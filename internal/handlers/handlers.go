package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/vibes/draft-board/internal/models"
	"github.com/vibes/draft-board/internal/repository"
	"github.com/vibes/draft-board/internal/snake"
	"github.com/vibes/draft-board/internal/validation"
)

type SSEEvent struct {
	Type string
	Data interface{}
}

type Handler struct {
	draftRepo  *repository.DraftRepository
	teamRepo   *repository.TeamRepository
	playerRepo *repository.PlayerRepository
	pickRepo   *repository.PickRepository
	queueRepo  *repository.QueueRepository
	auditRepo  *repository.AuditRepository

	// SSE: map of draft ID to channels
	sseClients map[int]map[chan SSEEvent]bool
	sseMutex   sync.RWMutex
}

func NewHandler(
	draftRepo *repository.DraftRepository,
	teamRepo *repository.TeamRepository,
	playerRepo *repository.PlayerRepository,
	pickRepo *repository.PickRepository,
	queueRepo *repository.QueueRepository,
	auditRepo *repository.AuditRepository,
) *Handler {
	return &Handler{
		draftRepo:  draftRepo,
		teamRepo:   teamRepo,
		playerRepo: playerRepo,
		pickRepo:   pickRepo,
		queueRepo:  queueRepo,
		auditRepo:  auditRepo,
		sseClients: make(map[int]map[chan SSEEvent]bool),
	}
}

func (h *Handler) broadcastEvent(draftID int, event SSEEvent) {
	h.sseMutex.RLock()
	clients, exists := h.sseClients[draftID]
	h.sseMutex.RUnlock()

	if !exists {
		return
	}

	for client := range clients {
		select {
		case client <- event:
		default:
			// Client channel is full, skip
		}
	}
}

func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	drafts, err := h.draftRepo.List()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var content strings.Builder
	content.WriteString(`
		<div class="mb-8">
			<h1 class="text-4xl font-bold mb-2 text-tokyo-night-accent">Fantasy Draft Board</h1>
			<p class="text-tokyo-night-fg-dim">Manage your offline fantasy football drafts</p>
		</div>
		<a href="/draft/new" class="inline-block mb-6 px-6 py-3 bg-tokyo-night-accent hover:bg-tokyo-night-accent-hover text-white rounded-lg font-semibold transition-colors">
			Create New Draft
		</a>
		<div class="mt-8">
			<h2 class="text-2xl font-semibold mb-4 text-tokyo-night-fg">Your Drafts</h2>
	`)

	if len(drafts) == 0 {
		content.WriteString(`
			<div class="bg-tokyo-night-bg-light rounded-lg p-8 text-center border border-tokyo-night-border">
				<p class="text-tokyo-night-fg-dim">No drafts yet. Create your first draft!</p>
			</div>
		`)
	} else {
		content.WriteString(`<div class="grid gap-4 md:grid-cols-2 lg:grid-cols-3">`)
		for _, d := range drafts {
			statusColor := "text-tokyo-night-success"
			if d.Status == "paused" {
				statusColor = "text-tokyo-night-warning"
			} else if d.Status == "completed" {
				statusColor = "text-tokyo-night-fg-dim"
			}
			content.WriteString(fmt.Sprintf(`
				<a href="/draft/%d" class="block bg-tokyo-night-bg-light rounded-lg p-6 border border-tokyo-night-border hover:border-tokyo-night-accent transition-colors">
					<h3 class="text-xl font-semibold mb-2 text-tokyo-night-fg">%s</h3>
					<div class="flex items-center gap-4 text-sm text-tokyo-night-fg-dim">
						<span class="%s">%s</span>
						<span>%d teams</span>
					</div>
				</a>
			`, d.ID, d.Name, statusColor, d.Status, d.NumTeams))
		}
		content.WriteString(`</div>`)
	}
	content.WriteString(`</div>`)

	w.Header().Set("Content-Type", "text/html")
	if err := renderTemplate(w, content.String(), "Fantasy Draft Board"); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) NewDraft(w http.ResponseWriter, r *http.Request) {
	var content strings.Builder
	content.WriteString(`
		<div class="max-w-2xl mx-auto">
			<div class="mb-8">
				<a href="/" class="text-tokyo-night-fg-dim hover:text-tokyo-night-accent mb-4 inline-block">← Back</a>
				<h1 class="text-4xl font-bold mb-2 text-tokyo-night-accent">Create New Draft</h1>
				<p class="text-tokyo-night-fg-dim">Configure your draft settings</p>
			</div>
			<div class="bg-tokyo-night-bg-light rounded-lg border border-tokyo-night-border p-6">
				<form method="POST" action="/draft/create" class="space-y-6">
					<div>
						<label class="block text-sm font-medium mb-2 text-tokyo-night-fg">Draft Name</label>
						<input type="text" name="name" required 
							class="w-full px-4 py-2 bg-tokyo-night-bg border border-tokyo-night-border rounded-lg text-tokyo-night-fg focus:outline-none focus:border-tokyo-night-accent">
					</div>
					<div>
						<label class="block text-sm font-medium mb-2 text-tokyo-night-fg">Number of Teams</label>
						<select name="num_teams" required 
							class="w-full px-4 py-2 bg-tokyo-night-bg border border-tokyo-night-border rounded-lg text-tokyo-night-fg focus:outline-none focus:border-tokyo-night-accent">
							<option value="8">8 Teams</option>
							<option value="10" selected>10 Teams</option>
							<option value="12">12 Teams</option>
							<option value="14">14 Teams</option>
						</select>
					</div>
					<div>
						<label class="block text-sm font-medium mb-2 text-tokyo-night-fg">Scoring Format</label>
						<select name="scoring_format" required 
							class="w-full px-4 py-2 bg-tokyo-night-bg border border-tokyo-night-border rounded-lg text-tokyo-night-fg focus:outline-none focus:border-tokyo-night-accent">
							<option value="Standard">Standard</option>
							<option value="Half-PPR">Half-PPR</option>
							<option value="PPR" selected>PPR</option>
						</select>
					</div>
					<div>
						<label class="block text-sm font-medium mb-2 text-tokyo-night-fg">Draft Type</label>
						<select name="draft_type" required 
							class="w-full px-4 py-2 bg-tokyo-night-bg border border-tokyo-night-border rounded-lg text-tokyo-night-fg focus:outline-none focus:border-tokyo-night-accent">
							<option value="Redraft" selected>Redraft</option>
							<option value="Dynasty">Dynasty</option>
						</select>
					</div>
					<div>
						<label class="block text-sm font-medium mb-2 text-tokyo-night-fg">Max Rounds</label>
						<input type="number" name="max_rounds" value="16" min="1" max="30" 
							class="w-full px-4 py-2 bg-tokyo-night-bg border border-tokyo-night-border rounded-lg text-tokyo-night-fg focus:outline-none focus:border-tokyo-night-accent">
					</div>
					<button type="submit" 
						class="w-full px-6 py-3 bg-tokyo-night-accent hover:bg-tokyo-night-accent-hover text-white rounded-lg font-semibold transition-colors">
						Create Draft
					</button>
				</form>
			</div>
		</div>
	`)
	renderTemplate(w, content.String(), "Create New Draft")
}

func (h *Handler) CreateDraft(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	numTeams, _ := strconv.Atoi(r.FormValue("num_teams"))
	maxRounds, _ := strconv.Atoi(r.FormValue("max_rounds"))
	if maxRounds == 0 {
		maxRounds = 16
	}

	draft := &models.Draft{
		Name:           r.FormValue("name"),
		NumTeams:       numTeams,
		ScoringFormat:  r.FormValue("scoring_format"),
		DraftType:      r.FormValue("draft_type"),
		QBSetting:      "1QB",
		SnakeDraft:     true,
		Status:         "setup",
		MaxRounds:      maxRounds,
		CommissionerID: uuid.New().String(),
		Completed:      false,
	}

	if err := validation.ValidateDraft(draft); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.draftRepo.Create(draft); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/draft/%d/setup", draft.ID), http.StatusSeeOther)
}

func (h *Handler) DraftSetup(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid draft ID", http.StatusBadRequest)
		return
	}

	draft, err := h.draftRepo.GetByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	teams, err := h.teamRepo.GetByDraft(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var content strings.Builder
	content.WriteString(`
		<div class="mb-8">
			<a href="/" class="text-tokyo-night-fg-dim hover:text-tokyo-night-accent mb-4 inline-block">← Back</a>
			<h1 class="text-4xl font-bold mb-2 text-tokyo-night-accent">Setup Draft: ` + draft.Name + `</h1>
			<div class="flex items-center gap-4 text-sm text-tokyo-night-fg-dim">
				<span class="px-3 py-1 rounded-full bg-tokyo-night-bg-light border border-tokyo-night-border">` + draft.Status + `</span>
				<span>Teams: ` + fmt.Sprintf("%d/%d", len(teams), draft.NumTeams) + `</span>
			</div>
		</div>
		<div class="grid md:grid-cols-2 gap-8">
			<div class="bg-tokyo-night-bg-light rounded-lg border border-tokyo-night-border p-6">
				<h2 class="text-2xl font-semibold mb-4 text-tokyo-night-fg">Add Team</h2>
				<form method="POST" action="/draft/` + fmt.Sprintf("%d", id) + `/teams" class="space-y-4">
					<div>
						<label class="block text-sm font-medium mb-2 text-tokyo-night-fg">Team Name</label>
						<input type="text" name="team_name" required maxlength="50" 
							class="w-full px-4 py-2 bg-tokyo-night-bg border border-tokyo-night-border rounded-lg text-tokyo-night-fg focus:outline-none focus:border-tokyo-night-accent">
					</div>
					<div>
						<label class="block text-sm font-medium mb-2 text-tokyo-night-fg">Owner Name (optional)</label>
						<input type="text" name="owner_name" maxlength="50" 
							class="w-full px-4 py-2 bg-tokyo-night-bg border border-tokyo-night-border rounded-lg text-tokyo-night-fg focus:outline-none focus:border-tokyo-night-accent">
					</div>
					<div>
						<label class="block text-sm font-medium mb-2 text-tokyo-night-fg">Draft Position</label>
						<input type="number" name="draft_position" min="1" max="` + fmt.Sprintf("%d", draft.NumTeams) + `" required 
							class="w-full px-4 py-2 bg-tokyo-night-bg border border-tokyo-night-border rounded-lg text-tokyo-night-fg focus:outline-none focus:border-tokyo-night-accent">
					</div>
					<button type="submit" 
						class="w-full px-4 py-2 bg-tokyo-night-accent hover:bg-tokyo-night-accent-hover text-white rounded-lg font-semibold transition-colors">
						Add Team
					</button>
				</form>
			</div>
			<div class="bg-tokyo-night-bg-light rounded-lg border border-tokyo-night-border p-6">
				<h2 class="text-2xl font-semibold mb-4 text-tokyo-night-fg">Teams (` + fmt.Sprintf("%d/%d", len(teams), draft.NumTeams) + `)</h2>
	`)

	if len(teams) == 0 {
		content.WriteString(`
			<p class="text-tokyo-night-fg-dim">No teams added yet.</p>
		`)
	} else {
		content.WriteString(`<div class="space-y-2">`)
		for _, team := range teams {
			content.WriteString(fmt.Sprintf(`
				<div class="flex items-center justify-between p-3 bg-tokyo-night-bg rounded border border-tokyo-night-border">
					<div>
						<span class="font-semibold text-tokyo-night-fg">%d. %s</span>
						<span class="text-sm text-tokyo-night-fg-dim ml-2">%s</span>
					</div>
				</div>
			`, team.DraftPosition, team.TeamName, team.OwnerName))
		}
		content.WriteString(`</div>`)
	}
	content.WriteString(`</div>`)

	if len(teams) == draft.NumTeams {
		content.WriteString(`
			<div class="mt-8 text-center">
				<form method="POST" action="/draft/` + fmt.Sprintf("%d", id) + `/start">
					<button type="submit" 
						class="px-8 py-3 bg-tokyo-night-success hover:bg-green-600 text-white rounded-lg font-semibold text-lg transition-colors">
						Start Draft
					</button>
				</form>
			</div>
		`)
	}
	content.WriteString(`</div>`)

	renderTemplate(w, content.String(), "Setup Draft: "+draft.Name)
}

func (h *Handler) CreateTeam(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	draftID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid draft ID", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	draft, err := h.draftRepo.GetByID(draftID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	draftPosition, _ := strconv.Atoi(r.FormValue("draft_position"))

	team := &models.Team{
		DraftID:       draftID,
		TeamName:      r.FormValue("team_name"),
		OwnerName:     r.FormValue("owner_name"),
		DraftPosition: draftPosition,
	}

	existingTeams, _ := h.teamRepo.GetByDraft(draftID)
	if err := validation.ValidateTeam(team, existingTeams, draft.NumTeams); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.teamRepo.Create(team); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/draft/%d/setup", draftID), http.StatusSeeOther)
}

func (h *Handler) UpdateDraft(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid draft ID", http.StatusBadRequest)
		return
	}

	draft, err := h.draftRepo.GetByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if name := r.FormValue("name"); name != "" {
		draft.Name = name
	}
	if numTeams := r.FormValue("num_teams"); numTeams != "" {
		if nt, err := strconv.Atoi(numTeams); err == nil {
			draft.NumTeams = nt
		}
	}
	if scoringFormat := r.FormValue("scoring_format"); scoringFormat != "" {
		draft.ScoringFormat = scoringFormat
	}
	if draftType := r.FormValue("draft_type"); draftType != "" {
		draft.DraftType = draftType
	}
	if maxRounds := r.FormValue("max_rounds"); maxRounds != "" {
		if mr, err := strconv.Atoi(maxRounds); err == nil {
			draft.MaxRounds = mr
		}
	}

	if err := validation.ValidateDraft(draft); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.draftRepo.Update(draft); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/draft/%d/setup", id), http.StatusSeeOther)
}

func (h *Handler) StartDraft(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid draft ID", http.StatusBadRequest)
		return
	}

	draft, err := h.draftRepo.GetByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	teamCount, err := h.teamRepo.CountByDraft(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := validation.ValidateTeamRosterCount(teamCount, draft.NumTeams); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	draft.Status = "active"
	if err := h.draftRepo.Update(draft); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.auditRepo.Log(id, "start", nil, "Draft started")

	http.Redirect(w, r, fmt.Sprintf("/draft/%d", id), http.StatusSeeOther)
}

func (h *Handler) PauseDraft(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid draft ID", http.StatusBadRequest)
		return
	}

	draft, err := h.draftRepo.GetByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	draft.Status = "paused"
	if err := h.draftRepo.Update(draft); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.auditRepo.Log(id, "pause", nil, "Draft paused")

	http.Redirect(w, r, fmt.Sprintf("/draft/%d", id), http.StatusSeeOther)
}

func (h *Handler) ResumeDraft(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid draft ID", http.StatusBadRequest)
		return
	}

	draft, err := h.draftRepo.GetByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	draft.Status = "active"
	if err := h.draftRepo.Update(draft); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.auditRepo.Log(id, "resume", nil, "Draft resumed")

	http.Redirect(w, r, fmt.Sprintf("/draft/%d", id), http.StatusSeeOther)
}

func (h *Handler) CompleteDraft(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid draft ID", http.StatusBadRequest)
		return
	}

	draft, err := h.draftRepo.GetByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	draft.Status = "completed"
	draft.Completed = true
	if err := h.draftRepo.Update(draft); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.auditRepo.Log(id, "complete", nil, "Draft completed")

	http.Redirect(w, r, fmt.Sprintf("/draft/%d", id), http.StatusSeeOther)
}

func (h *Handler) DeleteDraft(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid draft ID", http.StatusBadRequest)
		return
	}

	if err := h.draftRepo.Delete(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) GetDraftBoard(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid draft ID", http.StatusBadRequest)
		return
	}

	draft, err := h.draftRepo.GetByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	teams, err := h.teamRepo.GetByDraft(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	picks, err := h.pickRepo.GetByDraft(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pickCount, _ := h.pickRepo.CountByDraft(id)
	currentPick := pickCount + 1

	var currentTeam *models.Team
	if draft.IsActive() {
		snakeTeams := make([]snake.Team, len(teams))
		for i, t := range teams {
			snakeTeams[i] = snake.Team{
				ID:            t.ID,
				DraftPosition: t.DraftPosition,
			}
		}
		if team, err := snake.CalculateCurrentTeam(currentPick, draft.NumTeams, snakeTeams); err == nil {
			for _, t := range teams {
				if t.ID == team.ID {
					currentTeam = &t
					break
				}
			}
		}
	}

	var content strings.Builder
	content.WriteString(`
		<div class="mb-8">
			<div class="flex items-center justify-between mb-4">
				<div>
					<a href="/" class="text-tokyo-night-fg-dim hover:text-tokyo-night-accent mb-2 inline-block">← Back</a>
					<h1 class="text-4xl font-bold mb-2 text-tokyo-night-accent">` + draft.Name + `</h1>
				</div>
			</div>
			<div class="flex flex-wrap items-center gap-4 mb-6">
	`)

	statusColor := "bg-tokyo-night-success"
	if draft.Status == "paused" {
		statusColor = "bg-tokyo-night-warning"
	} else if draft.Status == "completed" {
		statusColor = "bg-tokyo-night-fg-dim"
	}

	content.WriteString(fmt.Sprintf(`
				<span class="px-3 py-1 rounded-full %s text-white text-sm font-medium">%s</span>
				<span class="text-tokyo-night-fg-dim">Round %d</span>
				<span class="text-tokyo-night-fg-dim">Pick %d</span>
	`, statusColor, draft.Status, snake.CalculateRound(currentPick, draft.NumTeams), currentPick))

	if currentTeam != nil {
		content.WriteString(fmt.Sprintf(`
				<span class="px-4 py-2 bg-tokyo-night-accent text-white rounded-lg font-semibold">On the Clock: %s</span>
		`, currentTeam.TeamName))
	}
	content.WriteString(`</div></div>`)

	// SSE connection for real-time updates (only if draft is active)
	if draft.IsActive() || draft.IsPaused() {
		content.WriteString(fmt.Sprintf(`
			<script>
				(function() {
					const draftId = %d;
					const eventSource = new EventSource('/draft/' + draftId + '/stream');
					
					eventSource.addEventListener('pick-made', function(event) {
						console.log('SSE: Pick made', event.data);
						eventSource.close();
						location.reload();
					});
					
					eventSource.addEventListener('pick-undone', function(event) {
						console.log('SSE: Pick undone', event.data);
						eventSource.close();
						location.reload();
					});
					
					eventSource.addEventListener('draft-completed', function(event) {
						console.log('SSE: Draft completed', event.data);
						eventSource.close();
						setTimeout(() => location.reload(), 1000);
					});
					
					eventSource.addEventListener('connected', function(event) {
						console.log('SSE: Connected to stream', event.data);
					});
					
					eventSource.onerror = function(event) {
						console.error('SSE: Connection error', event);
						// Reconnect after 3 seconds
						setTimeout(function() {
							if (eventSource.readyState === EventSource.CLOSED) {
								location.reload();
							}
						}, 3000);
					};
					
					// Clean up on page unload
					window.addEventListener('beforeunload', function() {
						eventSource.close();
					});
				})();
			</script>
		`, id))
	}

	// Draft board grid
	content.WriteString(`<div class="overflow-x-auto mb-8" id="draft-board">`)
	content.WriteString(`<table class="w-full border-collapse bg-tokyo-night-bg-light rounded-lg overflow-hidden">`)
	content.WriteString(`<thead><tr class="bg-tokyo-night-bg-dark"><th class="px-4 py-3 text-left font-semibold text-tokyo-night-fg border-b border-tokyo-night-border">Round</th>`)
	for _, team := range teams {
		content.WriteString(fmt.Sprintf(`<th class="px-4 py-3 text-left font-semibold text-tokyo-night-fg border-b border-tokyo-night-border">%s</th>`, team.TeamName))
	}
	content.WriteString(`</tr></thead><tbody>`)

	maxRound := draft.MaxRounds
	if len(picks) > 0 {
		lastRound := picks[len(picks)-1].Round
		if lastRound > maxRound {
			maxRound = lastRound
		}
	}

	pickMap := make(map[int]map[int]*models.Pick)
	for i := range picks {
		pick := &picks[i]
		if pickMap[pick.Round] == nil {
			pickMap[pick.Round] = make(map[int]*models.Pick)
		}
		pickMap[pick.Round][pick.TeamID] = pick
	}

	for round := 1; round <= maxRound; round++ {
		isCurrentRound := snake.CalculateRound(currentPick, draft.NumTeams) == round
		rowClass := ""
		if isCurrentRound {
			rowClass = "bg-tokyo-night-accent/10"
		}
		content.WriteString(fmt.Sprintf(`<tr class="%s">`, rowClass))
		content.WriteString(fmt.Sprintf(`<td class="px-4 py-2 font-medium text-tokyo-night-fg border-b border-tokyo-night-border">Round %d</td>`, round))
		for _, team := range teams {
			if pick, ok := pickMap[round][team.ID]; ok {
				player, _ := h.playerRepo.GetByID(pick.PlayerID)
				if player != nil {
					content.WriteString(fmt.Sprintf(`<td class="px-4 py-2 border-b border-tokyo-night-border">
						<div class="font-medium text-tokyo-night-fg">%s</div>
						<div class="text-sm text-tokyo-night-fg-dim">%s - %s</div>
					</td>`, player.Name, player.Position, player.Team))
				} else {
					content.WriteString(`<td class="px-4 py-2 border-b border-tokyo-night-border text-tokyo-night-fg-dim">-</td>`)
				}
			} else {
				content.WriteString(`<td class="px-4 py-2 border-b border-tokyo-night-border text-tokyo-night-fg-dim">-</td>`)
			}
		}
		content.WriteString(`</tr>`)
	}
	content.WriteString(`</tbody></table></div>`)

	// Navigation links
	content.WriteString(`
		<div class="mb-6 flex flex-wrap gap-4">
			<a href="/draft/` + fmt.Sprintf("%d", id) + `/players" class="px-4 py-2 bg-tokyo-night-bg-light hover:bg-tokyo-night-bg border border-tokyo-night-border rounded-lg transition-colors">
				View Available Players
			</a>
			<a href="/draft/` + fmt.Sprintf("%d", id) + `/big-board" class="px-4 py-2 bg-tokyo-night-bg-light hover:bg-tokyo-night-bg border border-tokyo-night-border rounded-lg transition-colors">
				Big Board
			</a>
			<a href="/draft/` + fmt.Sprintf("%d", id) + `/stats/franchise" class="px-4 py-2 bg-tokyo-night-bg-light hover:bg-tokyo-night-bg border border-tokyo-night-border rounded-lg transition-colors">
				Stats by Franchise
			</a>
			<a href="/draft/` + fmt.Sprintf("%d", id) + `/stats/position" class="px-4 py-2 bg-tokyo-night-bg-light hover:bg-tokyo-night-bg border border-tokyo-night-border rounded-lg transition-colors">
				Drafted by Position
			</a>
			<a href="/draft/` + fmt.Sprintf("%d", id) + `/stats/value-picks" class="px-4 py-2 bg-tokyo-night-bg-light hover:bg-tokyo-night-bg border border-tokyo-night-border rounded-lg transition-colors">
				Value Picks
			</a>
			<a href="/draft/` + fmt.Sprintf("%d", id) + `/export/csv" class="px-4 py-2 bg-tokyo-night-bg-light hover:bg-tokyo-night-bg border border-tokyo-night-border rounded-lg transition-colors">
				Export CSV
			</a>
			<a href="/draft/` + fmt.Sprintf("%d", id) + `/export/json" class="px-4 py-2 bg-tokyo-night-bg-light hover:bg-tokyo-night-bg border border-tokyo-night-border rounded-lg transition-colors">
				Export JSON
			</a>
		</div>
	`)

	// Undo button if draft is active
	if draft.IsActive() {
		pickCount, _ := h.pickRepo.CountByDraft(id)
		if pickCount > 0 {
			content.WriteString(fmt.Sprintf(`
				<form method="POST" action="/draft/%d/undo" class="inline-block">
					<button type="submit" class="px-4 py-2 bg-tokyo-night-warning hover:bg-yellow-600 text-white rounded-lg font-semibold transition-colors">
						Undo Last Pick
					</button>
				</form>
			`, id))
		}
	}

	// Control buttons
	content.WriteString(`<div class="flex gap-4 mt-6">`)
	if draft.IsActive() {
		content.WriteString(fmt.Sprintf(`
			<form method="POST" action="/draft/%d/pause">
				<button type="submit" class="px-4 py-2 bg-tokyo-night-warning hover:bg-yellow-600 text-white rounded-lg font-semibold transition-colors">
					Pause Draft
				</button>
			</form>
		`, id))
	} else if draft.IsPaused() {
		content.WriteString(fmt.Sprintf(`
			<form method="POST" action="/draft/%d/resume">
				<button type="submit" class="px-4 py-2 bg-tokyo-night-success hover:bg-green-600 text-white rounded-lg font-semibold transition-colors">
					Resume Draft
				</button>
			</form>
		`, id))
	}
	if !draft.IsCompleted() {
		content.WriteString(fmt.Sprintf(`
			<form method="POST" action="/draft/%d/complete">
				<button type="submit" class="px-4 py-2 bg-tokyo-night-fg-dim hover:bg-gray-600 text-white rounded-lg font-semibold transition-colors">
					Complete Draft
				</button>
			</form>
		`, id))
	}
	content.WriteString(`</div>`)

	renderTemplate(w, content.String(), draft.Name)
}

func (h *Handler) GetAvailablePlayers(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid draft ID", http.StatusBadRequest)
		return
	}

	draft, err := h.draftRepo.GetByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	positions := r.URL.Query()["position"]
	search := r.URL.Query().Get("search")

	filters := repository.PlayerFilters{
		Positions:     positions,
		Search:        search,
		DraftType:     draft.DraftType,
		ScoringFormat: draft.ScoringFormat,
		Limit:         100,
	}

	players, err := h.playerRepo.GetAvailable(id, filters)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var content strings.Builder
	content.WriteString(`
		<div class="mb-8">
			<a href="/draft/` + fmt.Sprintf("%d", id) + `" class="text-tokyo-night-fg-dim hover:text-tokyo-night-accent mb-4 inline-block">← Back to Draft</a>
			<h1 class="text-4xl font-bold mb-2 text-tokyo-night-accent">Available Players</h1>
		</div>
		<div class="mb-6">
			<form method="GET" action="/draft/` + fmt.Sprintf("%d", id) + `/players" class="flex gap-4">
				<input type="text" name="search" placeholder="Search players..." value="` + search + `" 
					class="flex-1 px-4 py-2 bg-tokyo-night-bg-light border border-tokyo-night-border rounded-lg text-tokyo-night-fg focus:outline-none focus:border-tokyo-night-accent">
				<button type="submit" class="px-6 py-2 bg-tokyo-night-accent hover:bg-tokyo-night-accent-hover text-white rounded-lg font-semibold transition-colors">
					Search
				</button>
			</form>
		</div>
		<div class="overflow-x-auto">
			<table class="w-full border-collapse bg-tokyo-night-bg-light rounded-lg overflow-hidden">
				<thead>
					<tr class="bg-tokyo-night-bg-dark">
						<th class="px-4 py-3 text-left font-semibold text-tokyo-night-fg border-b border-tokyo-night-border">Rank</th>
						<th class="px-4 py-3 text-left font-semibold text-tokyo-night-fg border-b border-tokyo-night-border">Name</th>
						<th class="px-4 py-3 text-left font-semibold text-tokyo-night-fg border-b border-tokyo-night-border">Team</th>
						<th class="px-4 py-3 text-left font-semibold text-tokyo-night-fg border-b border-tokyo-night-border">Position</th>
						<th class="px-4 py-3 text-left font-semibold text-tokyo-night-fg border-b border-tokyo-night-border">Bye</th>
						<th class="px-4 py-3 text-left font-semibold text-tokyo-night-fg border-b border-tokyo-night-border">Action</th>
					</tr>
				</thead>
				<tbody>
	`)

	for _, player := range players {
		rank := "-"
		if r := player.GetADPRank(draft.DraftType, draft.ScoringFormat); r != nil {
			rank = fmt.Sprintf("%d", *r)
		}
		bye := "-"
		if player.ByeWeek != nil {
			bye = fmt.Sprintf("%d", *player.ByeWeek)
		}

		content.WriteString(`<tr class="hover:bg-tokyo-night-bg-dark transition-colors">`)
		content.WriteString(fmt.Sprintf(`<td class="px-4 py-2 border-b border-tokyo-night-border text-tokyo-night-fg">%s</td>`, rank))
		content.WriteString(fmt.Sprintf(`<td class="px-4 py-2 border-b border-tokyo-night-border font-medium text-tokyo-night-fg">%s</td>`, player.Name))
		content.WriteString(fmt.Sprintf(`<td class="px-4 py-2 border-b border-tokyo-night-border text-tokyo-night-fg-dim">%s</td>`, player.Team))
		content.WriteString(fmt.Sprintf(`<td class="px-4 py-2 border-b border-tokyo-night-border text-tokyo-night-fg-dim">%s</td>`, player.Position))
		content.WriteString(fmt.Sprintf(`<td class="px-4 py-2 border-b border-tokyo-night-border text-tokyo-night-fg-dim">%s</td>`, bye))

		if draft.CanMakePicks() {
			content.WriteString(fmt.Sprintf(`<td class="px-4 py-2 border-b border-tokyo-night-border">
				<form method="POST" action="/draft/%d/pick" class="inline">
					<input type="hidden" name="player_id" value="%d">
					<button type="submit" class="px-3 py-1 bg-tokyo-night-accent hover:bg-tokyo-night-accent-hover text-white rounded text-sm font-semibold transition-colors">
						Draft
					</button>
				</form>
			</td>`, id, player.ID))
		} else {
			content.WriteString(`<td class="px-4 py-2 border-b border-tokyo-night-border text-tokyo-night-fg-dim">-</td>`)
		}
		content.WriteString(`</tr>`)
	}

	content.WriteString(`</tbody></table></div>`)
	renderTemplate(w, content.String(), "Available Players")
}

func (h *Handler) MakePick(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	draftID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid draft ID", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	playerID, err := strconv.Atoi(r.FormValue("player_id"))
	if err != nil {
		http.Error(w, "Invalid player ID", http.StatusBadRequest)
		return
	}

	draft, err := h.draftRepo.GetByID(draftID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if !draft.CanMakePicks() {
		http.Error(w, validation.ErrDraftNotActive.Error(), http.StatusBadRequest)
		return
	}

	pickCount, _ := h.pickRepo.CountByDraft(draftID)
	currentPickNumber := pickCount + 1

	teams, _ := h.teamRepo.GetByDraft(draftID)
	snakeTeams := make([]snake.Team, len(teams))
	for i, t := range teams {
		snakeTeams[i] = snake.Team{
			ID:            t.ID,
			DraftPosition: t.DraftPosition,
		}
	}

	currentTeam, err := snake.CalculateCurrentTeam(currentPickNumber, draft.NumTeams, snakeTeams)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var teamName string
	for _, t := range teams {
		if t.ID == currentTeam.ID {
			teamName = t.TeamName
			break
		}
	}

	player, err := h.playerRepo.GetByID(playerID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	draftedPlayerIDs, _ := h.pickRepo.GetDraftedPlayerIDs(draftID)
	if err := validation.ValidatePlayerNotDrafted(playerID, draftedPlayerIDs); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	adpRank := player.GetADPRank(draft.DraftType, draft.ScoringFormat)
	round := snake.CalculateRound(currentPickNumber, draft.NumTeams)

	pick := &models.Pick{
		DraftID:     draftID,
		TeamID:      currentTeam.ID,
		PlayerID:    playerID,
		Round:       round,
		OverallPick: currentPickNumber,
		ADPRank:     adpRank,
		IsTraded:    false,
	}

	if err := validation.ValidatePick(pick, draft, teams, pickCount); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.pickRepo.Create(pick); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.auditRepo.Log(draftID, "pick", &pick.ID, fmt.Sprintf("%s drafted by team %d", player.Name, currentTeam.ID))

	// Broadcast SSE event
	h.broadcastEvent(draftID, SSEEvent{
		Type: "pick-made",
		Data: map[string]interface{}{
			"pick_id":      pick.ID,
			"player_id":    playerID,
			"player_name":  player.Name,
			"team_id":      currentTeam.ID,
			"team_name":    teamName,
			"round":        round,
			"overall_pick": currentPickNumber,
		},
	})

	// Check if draft is complete
	pickCount++
	if draft.CheckDraftCompletion(pickCount) {
		draft.Status = "completed"
		draft.Completed = true
		h.draftRepo.Update(draft)
		h.auditRepo.Log(draftID, "complete", nil, "Draft auto-completed")
		h.broadcastEvent(draftID, SSEEvent{
			Type: "draft-completed",
			Data: map[string]interface{}{
				"draft_id": draftID,
			},
		})
	}

	http.Redirect(w, r, fmt.Sprintf("/draft/%d", draftID), http.StatusSeeOther)
}

func (h *Handler) UndoPick(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	draftID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid draft ID", http.StatusBadRequest)
		return
	}

	lastPick, err := h.pickRepo.GetLast(draftID)
	if err != nil || lastPick == nil {
		http.Error(w, "No pick to undo", http.StatusBadRequest)
		return
	}

	if err := h.pickRepo.Delete(lastPick.ID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.auditRepo.Log(draftID, "undo", &lastPick.ID, fmt.Sprintf("Undid pick %d", lastPick.OverallPick))

	h.broadcastEvent(draftID, SSEEvent{
		Type: "pick-undone",
		Data: map[string]interface{}{
			"pick_id":  lastPick.ID,
			"draft_id": draftID,
		},
	})

	http.Redirect(w, r, fmt.Sprintf("/draft/%d", draftID), http.StatusSeeOther)
}

func (h *Handler) TradePick(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PickID   int    `json:"pick_id"`
		ToTeamID int    `json:"to_team_id"`
		Notes    string `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	pick, err := h.pickRepo.GetByID(req.PickID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	pick.TeamID = req.ToTeamID
	pick.IsTraded = true

	if err := h.pickRepo.Update(pick); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.auditRepo.Log(pick.DraftID, "trade", &pick.ID, req.Notes)

	http.Redirect(w, r, fmt.Sprintf("/draft/%d", pick.DraftID), http.StatusSeeOther)
}

func (h *Handler) GetCurrentPick(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid draft ID", http.StatusBadRequest)
		return
	}

	draft, err := h.draftRepo.GetByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	pickCount, _ := h.pickRepo.CountByDraft(id)
	currentPick := pickCount + 1

	teams, _ := h.teamRepo.GetByDraft(id)
	snakeTeams := make([]snake.Team, len(teams))
	for i, t := range teams {
		snakeTeams[i] = snake.Team{
			ID:            t.ID,
			DraftPosition: t.DraftPosition,
		}
	}

	currentTeam, err := snake.CalculateCurrentTeam(currentPick, draft.NumTeams, snakeTeams)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var team *models.Team
	for _, t := range teams {
		if t.ID == currentTeam.ID {
			team = &t
			break
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"pick_number": currentPick,
		"round":       snake.CalculateRound(currentPick, draft.NumTeams),
		"team_id":     team.ID,
		"team_name":   team.TeamName,
	})
}

func (h *Handler) GetTeams(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid draft ID", http.StatusBadRequest)
		return
	}

	teams, err := h.teamRepo.GetByDraft(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(teams)
}

func (h *Handler) UpdateTeam(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid team ID", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	team, err := h.teamRepo.GetByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	team.TeamName = r.FormValue("team_name")
	team.OwnerName = r.FormValue("owner_name")
	if pos := r.FormValue("draft_position"); pos != "" {
		team.DraftPosition, _ = strconv.Atoi(pos)
	}

	draft, _ := h.draftRepo.GetByID(team.DraftID)
	existingTeams, _ := h.teamRepo.GetByDraft(team.DraftID)
	if err := validation.ValidateTeam(team, existingTeams, draft.NumTeams); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.teamRepo.Update(team); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/draft/%d/setup", team.DraftID), http.StatusSeeOther)
}

func (h *Handler) DeleteTeam(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid team ID", http.StatusBadRequest)
		return
	}

	team, err := h.teamRepo.GetByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if err := h.teamRepo.Delete(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/draft/%d/setup", team.DraftID), http.StatusSeeOther)
}

func (h *Handler) GetQueue(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	draftID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid draft ID", http.StatusBadRequest)
		return
	}

	teamIDStr := r.URL.Query().Get("team_id")
	teamID, err := strconv.Atoi(teamIDStr)
	if err != nil {
		http.Error(w, "Invalid team ID", http.StatusBadRequest)
		return
	}

	items, err := h.queueRepo.GetByTeam(draftID, teamID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func (h *Handler) AddToQueue(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	draftID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid draft ID", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	teamID, _ := strconv.Atoi(r.FormValue("team_id"))
	playerID, _ := strconv.Atoi(r.FormValue("player_id"))

	maxOrder, _ := h.queueRepo.GetMaxOrder(draftID, teamID)

	queueItem := &models.QueueItem{
		DraftID:    draftID,
		TeamID:     teamID,
		PlayerID:   playerID,
		QueueOrder: maxOrder + 1,
	}

	if err := h.queueRepo.Create(queueItem); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/draft/%d", draftID), http.StatusSeeOther)
}

func (h *Handler) RemoveFromQueue(w http.ResponseWriter, r *http.Request) {
	queueIDStr := chi.URLParam(r, "queueId")
	queueID, err := strconv.Atoi(queueIDStr)
	if err != nil {
		http.Error(w, "Invalid queue ID", http.StatusBadRequest)
		return
	}

	if err := h.queueRepo.Delete(queueID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	idStr := chi.URLParam(r, "id")
	draftID, _ := strconv.Atoi(idStr)
	http.Redirect(w, r, fmt.Sprintf("/draft/%d", draftID), http.StatusSeeOther)
}

func (h *Handler) CreateCustomPlayer(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var byeWeek *int
	if bw := r.FormValue("bye_week"); bw != "" {
		if b, err := strconv.Atoi(bw); err == nil {
			byeWeek = &b
		}
	}

	player := &models.Player{
		Name:     r.FormValue("name"),
		Team:     r.FormValue("team"),
		Position: r.FormValue("position"),
		ByeWeek:  byeWeek,
		IsCustom: true,
	}

	if err := validation.ValidatePosition(player.Position); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.playerRepo.Create(player); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(player)
}

func (h *Handler) GetBigBoard(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid draft ID", http.StatusBadRequest)
		return
	}

	draft, err := h.draftRepo.GetByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	teams, err := h.teamRepo.GetByDraft(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	picks, err := h.pickRepo.GetByDraft(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Build pick map by round and team
	pickMap := make(map[int]map[int]*models.Pick)
	for i := range picks {
		pick := &picks[i]
		if pickMap[pick.Round] == nil {
			pickMap[pick.Round] = make(map[int]*models.Pick)
		}
		pickMap[pick.Round][pick.TeamID] = pick
	}

	w.Header().Set("Content-Type", "text/html")
	var content strings.Builder
	content.WriteString(`
		<div class="mb-8">
			<div class="flex items-center justify-between mb-4">
				<div>
					<a href="/draft/` + fmt.Sprintf("%d", id) + `" class="text-tokyo-night-fg-dim hover:text-tokyo-night-accent mb-4 inline-block">← Back to Draft</a>
					<h1 class="text-4xl font-bold mb-2 text-tokyo-night-accent">Big Board - ` + draft.Name + `</h1>
				</div>
				<button onclick="window.print()" class="no-print px-4 py-2 bg-tokyo-night-accent hover:bg-tokyo-night-accent-hover text-white rounded-lg font-semibold transition-colors">
					Print
				</button>
			</div>
		</div>
		<div class="overflow-x-auto">
			<table class="w-full border-collapse bg-tokyo-night-bg-light rounded-lg overflow-hidden text-sm">
				<thead>
					<tr class="bg-tokyo-night-bg-dark">
						<th class="px-3 py-2 text-left font-semibold text-tokyo-night-fg border-b border-tokyo-night-border">Round</th>
	`)

	for _, team := range teams {
		content.WriteString(fmt.Sprintf(`
			<th class="px-3 py-2 text-left font-semibold text-tokyo-night-fg border-b border-tokyo-night-border">
				<div>%s</div>
				<div class="text-xs text-tokyo-night-fg-dim font-normal">%s</div>
			</th>
		`, team.TeamName, team.OwnerName))
	}
	content.WriteString(`</tr></thead><tbody>`)

	maxRound := draft.MaxRounds
	if len(picks) > 0 {
		lastRound := picks[len(picks)-1].Round
		if lastRound > maxRound {
			maxRound = lastRound
		}
	}

	for round := 1; round <= maxRound; round++ {
		content.WriteString(fmt.Sprintf(`<tr><td class="px-3 py-2 font-medium text-tokyo-night-fg border-b border-tokyo-night-border">Round %d</td>`, round))
		for _, team := range teams {
			if pick, ok := pickMap[round][team.ID]; ok {
				player, _ := h.playerRepo.GetByID(pick.PlayerID)
				if player != nil {
					traded := ""
					if pick.IsTraded {
						traded = `<span class="text-red-500 text-xs">TRADED</span>`
					}
					content.WriteString(fmt.Sprintf(`<td class="px-3 py-2 border-b border-tokyo-night-border">
						<div class="font-medium text-tokyo-night-fg">%s</div>
						<div class="text-xs text-tokyo-night-fg-dim">%s - %s</div>
						%s
					</td>`, player.Name, player.Position, player.Team, traded))
				} else {
					content.WriteString(`<td class="px-3 py-2 border-b border-tokyo-night-border text-tokyo-night-fg-dim">-</td>`)
				}
			} else {
				content.WriteString(`<td class="px-3 py-2 border-b border-tokyo-night-border text-tokyo-night-fg-dim">-</td>`)
			}
		}
		content.WriteString(`</tr>`)
	}
	content.WriteString(`</tbody></table></div>`)

	content.WriteString(`
		<style media="print">
			.no-print { display: none !important; }
			table { font-size: 9pt; page-break-after: always; }
			@page { size: landscape; margin: 0.5in; }
			body { background: white; color: black; }
		</style>
	`)

	renderTemplate(w, content.String(), "Big Board - "+draft.Name)
}
