package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

// StreamUpdates implements Server-Sent Events for real-time draft updates
func (h *Handler) StreamUpdates(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	draftID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid draft ID", http.StatusBadRequest)
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Create client channel
	clientChan := make(chan SSEEvent, 10)

	// Register client
	h.sseMutex.Lock()
	if h.sseClients[draftID] == nil {
		h.sseClients[draftID] = make(map[chan SSEEvent]bool)
	}
	h.sseClients[draftID][clientChan] = true
	h.sseMutex.Unlock()

	// Send initial connection event
	fmt.Fprintf(w, "event: connected\n")
	fmt.Fprintf(w, "data: {\"draft_id\": %d}\n\n", draftID)
	w.(http.Flusher).Flush()

	ctx := r.Context()
	for {
		select {
		case event := <-clientChan:
			eventJSON, _ := json.Marshal(event.Data)
			fmt.Fprintf(w, "event: %s\n", event.Type)
			fmt.Fprintf(w, "data: %s\n\n", eventJSON)
			w.(http.Flusher).Flush()

		case <-ctx.Done():
			// Client disconnected
			h.sseMutex.Lock()
			delete(h.sseClients[draftID], clientChan)
			if len(h.sseClients[draftID]) == 0 {
				delete(h.sseClients, draftID)
			}
			h.sseMutex.Unlock()
			close(clientChan)
			return
		}
	}
}

// GetFranchiseStats shows players drafted by NFL team
func (h *Handler) GetFranchiseStats(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	draftID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid draft ID", http.StatusBadRequest)
		return
	}

	picks, err := h.pickRepo.GetByDraft(draftID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	type FranchiseStat struct {
		TeamAbbr    string
		TotalCount  int
		QBCount     int
		RBCount     int
		WRCount     int
		TECount     int
		KCount      int
		DSTCount    int
		OtherCount  int
	}

	stats := make(map[string]*FranchiseStat)

	for _, pick := range picks {
		player, err := h.playerRepo.GetByID(pick.PlayerID)
		if err != nil {
			continue
		}

		if stats[player.Team] == nil {
			stats[player.Team] = &FranchiseStat{TeamAbbr: player.Team}
		}

		stat := stats[player.Team]
		stat.TotalCount++

		switch player.Position {
		case "QB":
			stat.QBCount++
		case "RB":
			stat.RBCount++
		case "WR":
			stat.WRCount++
		case "TE":
			stat.TECount++
		case "K":
			stat.KCount++
		case "D/ST":
			stat.DSTCount++
		default:
			stat.OtherCount++
		}
	}

	w.Header().Set("Content-Type", "text/html")
	var content strings.Builder
	content.WriteString(`
		<div class="mb-8">
			<a href="/draft/` + fmt.Sprintf("%d", draftID) + `" class="text-tokyo-night-fg-dim hover:text-tokyo-night-accent mb-4 inline-block">← Back to Draft</a>
			<h1 class="text-4xl font-bold mb-2 text-tokyo-night-accent">Stats by NFL Franchise</h1>
		</div>
		<div class="overflow-x-auto">
			<table class="w-full border-collapse bg-tokyo-night-bg-light rounded-lg overflow-hidden">
				<thead>
					<tr class="bg-tokyo-night-bg-dark">
						<th class="px-4 py-3 text-left font-semibold text-tokyo-night-fg border-b border-tokyo-night-border">Team</th>
						<th class="px-4 py-3 text-left font-semibold text-tokyo-night-fg border-b border-tokyo-night-border">Total</th>
						<th class="px-4 py-3 text-left font-semibold text-tokyo-night-fg border-b border-tokyo-night-border">QB</th>
						<th class="px-4 py-3 text-left font-semibold text-tokyo-night-fg border-b border-tokyo-night-border">RB</th>
						<th class="px-4 py-3 text-left font-semibold text-tokyo-night-fg border-b border-tokyo-night-border">WR</th>
						<th class="px-4 py-3 text-left font-semibold text-tokyo-night-fg border-b border-tokyo-night-border">TE</th>
						<th class="px-4 py-3 text-left font-semibold text-tokyo-night-fg border-b border-tokyo-night-border">K</th>
						<th class="px-4 py-3 text-left font-semibold text-tokyo-night-fg border-b border-tokyo-night-border">D/ST</th>
						<th class="px-4 py-3 text-left font-semibold text-tokyo-night-fg border-b border-tokyo-night-border">Other</th>
					</tr>
				</thead>
				<tbody>
	`)

	for _, stat := range stats {
		content.WriteString(fmt.Sprintf(`
			<tr class="hover:bg-tokyo-night-bg-dark transition-colors">
				<td class="px-4 py-2 border-b border-tokyo-night-border font-medium text-tokyo-night-fg">%s</td>
				<td class="px-4 py-2 border-b border-tokyo-night-border text-tokyo-night-fg">%d</td>
				<td class="px-4 py-2 border-b border-tokyo-night-border text-tokyo-night-fg-dim">%d</td>
				<td class="px-4 py-2 border-b border-tokyo-night-border text-tokyo-night-fg-dim">%d</td>
				<td class="px-4 py-2 border-b border-tokyo-night-border text-tokyo-night-fg-dim">%d</td>
				<td class="px-4 py-2 border-b border-tokyo-night-border text-tokyo-night-fg-dim">%d</td>
				<td class="px-4 py-2 border-b border-tokyo-night-border text-tokyo-night-fg-dim">%d</td>
				<td class="px-4 py-2 border-b border-tokyo-night-border text-tokyo-night-fg-dim">%d</td>
				<td class="px-4 py-2 border-b border-tokyo-night-border text-tokyo-night-fg-dim">%d</td>
			</tr>
		`, stat.TeamAbbr, stat.TotalCount, stat.QBCount, stat.RBCount, stat.WRCount, stat.TECount, stat.KCount, stat.DSTCount, stat.OtherCount))
	}

	content.WriteString(`</tbody></table></div>`)
	renderTemplate(w, content.String(), "Stats by Franchise")
}

// GetDraftedByPosition shows players drafted organized by position
func (h *Handler) GetDraftedByPosition(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	draftID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid draft ID", http.StatusBadRequest)
		return
	}

	picks, err := h.pickRepo.GetByDraft(draftID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	type DraftedPlayer struct {
		PlayerName  string
		TeamName    string
		OverallPick int
		Position    string
	}

	byPosition := make(map[string][]DraftedPlayer)

	for _, pick := range picks {
		player, err := h.playerRepo.GetByID(pick.PlayerID)
		if err != nil {
			continue
		}

		team, err := h.teamRepo.GetByID(pick.TeamID)
		if err != nil {
			continue
		}

		byPosition[player.Position] = append(byPosition[player.Position], DraftedPlayer{
			PlayerName:  player.Name,
			TeamName:    team.TeamName,
			OverallPick: pick.OverallPick,
			Position:    player.Position,
		})
	}

	w.Header().Set("Content-Type", "text/html")
	var content strings.Builder
	content.WriteString(`
		<div class="mb-8">
			<a href="/draft/` + fmt.Sprintf("%d", draftID) + `" class="text-tokyo-night-fg-dim hover:text-tokyo-night-accent mb-4 inline-block">← Back to Draft</a>
			<h1 class="text-4xl font-bold mb-2 text-tokyo-night-accent">Players Drafted by Position</h1>
		</div>
		<div class="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
	`)

	positions := []string{"QB", "RB", "WR", "TE", "K", "D/ST"}
	for _, pos := range positions {
		players := byPosition[pos]
		content.WriteString(fmt.Sprintf(`
			<div class="bg-tokyo-night-bg-light rounded-lg border border-tokyo-night-border p-6">
				<h2 class="text-2xl font-semibold mb-4 text-tokyo-night-fg">%s (%d)</h2>
				<div class="space-y-2">
		`, pos, len(players)))
		for _, p := range players {
			content.WriteString(fmt.Sprintf(`
				<div class="p-3 bg-tokyo-night-bg rounded border border-tokyo-night-border">
					<div class="font-medium text-tokyo-night-fg">%s</div>
					<div class="text-sm text-tokyo-night-fg-dim">%s - Pick %d</div>
				</div>
			`, p.PlayerName, p.TeamName, p.OverallPick))
		}
		if len(players) == 0 {
			content.WriteString(`<p class="text-tokyo-night-fg-dim">No players drafted</p>`)
		}
		content.WriteString(`</div></div>`)
	}
	content.WriteString(`</div>`)

	renderTemplate(w, content.String(), "Drafted by Position")
}

// GetValuePicks shows steals and reaches based on ADP
func (h *Handler) GetValuePicks(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	draftID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid draft ID", http.StatusBadRequest)
		return
	}

	_, err = h.draftRepo.GetByID(draftID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	picks, err := h.pickRepo.GetByDraft(draftID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	type ValuePick struct {
		PlayerName  string
		TeamName    string
		ADPRank     int
		OverallPick int
		ValueDiff   int
		Position    string
	}

	var steals, reaches []ValuePick

	for _, pick := range picks {
		if pick.ADPRank == nil {
			continue
		}

		player, err := h.playerRepo.GetByID(pick.PlayerID)
		if err != nil {
			continue
		}

		team, err := h.teamRepo.GetByID(pick.TeamID)
		if err != nil {
			continue
		}

		valueDiff := *pick.ADPRank - pick.OverallPick

		vp := ValuePick{
			PlayerName:  player.Name,
			TeamName:    team.TeamName,
			ADPRank:     *pick.ADPRank,
			OverallPick: pick.OverallPick,
			ValueDiff:   valueDiff,
			Position:    player.Position,
		}

		if valueDiff >= 10 {
			steals = append(steals, vp)
		} else if valueDiff <= -10 {
			reaches = append(reaches, ValuePick{
				PlayerName:  vp.PlayerName,
				TeamName:    vp.TeamName,
				ADPRank:     vp.ADPRank,
				OverallPick: vp.OverallPick,
				ValueDiff:   -vp.ValueDiff,
				Position:    vp.Position,
			})
		}
	}

	w.Header().Set("Content-Type", "text/html")
	var content strings.Builder
	content.WriteString(`
		<div class="mb-8">
			<a href="/draft/` + fmt.Sprintf("%d", draftID) + `" class="text-tokyo-night-fg-dim hover:text-tokyo-night-accent mb-4 inline-block">← Back to Draft</a>
			<h1 class="text-4xl font-bold mb-2 text-tokyo-night-accent">Draft Value Analysis</h1>
		</div>
		<div class="grid md:grid-cols-2 gap-8">
			<div>
				<h2 class="text-2xl font-semibold mb-4 text-tokyo-night-success">Steals (Drafted Below ADP)</h2>
				<div class="overflow-x-auto">
					<table class="w-full border-collapse bg-tokyo-night-bg-light rounded-lg overflow-hidden">
						<thead>
							<tr class="bg-tokyo-night-bg-dark">
								<th class="px-4 py-3 text-left font-semibold text-tokyo-night-fg border-b border-tokyo-night-border">Player</th>
								<th class="px-4 py-3 text-left font-semibold text-tokyo-night-fg border-b border-tokyo-night-border">Team</th>
								<th class="px-4 py-3 text-left font-semibold text-tokyo-night-fg border-b border-tokyo-night-border">ADP</th>
								<th class="px-4 py-3 text-left font-semibold text-tokyo-night-fg border-b border-tokyo-night-border">Picked</th>
								<th class="px-4 py-3 text-left font-semibold text-tokyo-night-fg border-b border-tokyo-night-border">Value</th>
							</tr>
						</thead>
						<tbody>
	`)

	for _, steal := range steals {
		content.WriteString(fmt.Sprintf(`
			<tr class="hover:bg-tokyo-night-bg-dark transition-colors">
				<td class="px-4 py-2 border-b border-tokyo-night-border font-medium text-tokyo-night-fg">%s (%s)</td>
				<td class="px-4 py-2 border-b border-tokyo-night-border text-tokyo-night-fg-dim">%s</td>
				<td class="px-4 py-2 border-b border-tokyo-night-border text-tokyo-night-fg-dim">%d</td>
				<td class="px-4 py-2 border-b border-tokyo-night-border text-tokyo-night-fg-dim">%d</td>
				<td class="px-4 py-2 border-b border-tokyo-night-border text-tokyo-night-success font-semibold">+%d</td>
			</tr>
		`, steal.PlayerName, steal.Position, steal.TeamName, steal.ADPRank, steal.OverallPick, steal.ValueDiff))
	}
	if len(steals) == 0 {
		content.WriteString(`<tr><td colspan="5" class="px-4 py-8 text-center text-tokyo-night-fg-dim">No steals found</td></tr>`)
	}
	content.WriteString(`</tbody></table></div></div>`)

	content.WriteString(`
		<div>
			<h2 class="text-2xl font-semibold mb-4 text-tokyo-night-error">Reaches (Drafted Above ADP)</h2>
			<div class="overflow-x-auto">
				<table class="w-full border-collapse bg-tokyo-night-bg-light rounded-lg overflow-hidden">
					<thead>
						<tr class="bg-tokyo-night-bg-dark">
							<th class="px-4 py-3 text-left font-semibold text-tokyo-night-fg border-b border-tokyo-night-border">Player</th>
							<th class="px-4 py-3 text-left font-semibold text-tokyo-night-fg border-b border-tokyo-night-border">Team</th>
							<th class="px-4 py-3 text-left font-semibold text-tokyo-night-fg border-b border-tokyo-night-border">ADP</th>
							<th class="px-4 py-3 text-left font-semibold text-tokyo-night-fg border-b border-tokyo-night-border">Picked</th>
							<th class="px-4 py-3 text-left font-semibold text-tokyo-night-fg border-b border-tokyo-night-border">Reach</th>
						</tr>
					</thead>
					<tbody>
	`)

	for _, reach := range reaches {
		content.WriteString(fmt.Sprintf(`
			<tr class="hover:bg-tokyo-night-bg-dark transition-colors">
				<td class="px-4 py-2 border-b border-tokyo-night-border font-medium text-tokyo-night-fg">%s (%s)</td>
				<td class="px-4 py-2 border-b border-tokyo-night-border text-tokyo-night-fg-dim">%s</td>
				<td class="px-4 py-2 border-b border-tokyo-night-border text-tokyo-night-fg-dim">%d</td>
				<td class="px-4 py-2 border-b border-tokyo-night-border text-tokyo-night-fg-dim">%d</td>
				<td class="px-4 py-2 border-b border-tokyo-night-border text-tokyo-night-error font-semibold">-%d</td>
			</tr>
		`, reach.PlayerName, reach.Position, reach.TeamName, reach.ADPRank, reach.OverallPick, reach.ValueDiff))
	}
	if len(reaches) == 0 {
		content.WriteString(`<tr><td colspan="5" class="px-4 py-8 text-center text-tokyo-night-fg-dim">No reaches found</td></tr>`)
	}
	content.WriteString(`</tbody></table></div></div></div>`)

	renderTemplate(w, content.String(), "Value Picks Analysis")
}

// ExportCSV exports draft results as CSV
func (h *Handler) ExportCSV(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	draftID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid draft ID", http.StatusBadRequest)
		return
	}

	picks, err := h.pickRepo.GetByDraft(draftID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=draft-%d.csv", draftID))

	fmt.Fprintf(w, "Round,Overall Pick,Team,Player,Position,NFL Team,ADP Rank\n")

	for _, pick := range picks {
		player, _ := h.playerRepo.GetByID(pick.PlayerID)
		team, _ := h.teamRepo.GetByID(pick.TeamID)

		playerName := ""
		position := ""
		teamAbbr := ""
		if player != nil {
			playerName = player.Name
			position = player.Position
			teamAbbr = player.Team
		}

		teamName := ""
		if team != nil {
			teamName = team.TeamName
		}

		adpRank := ""
		if pick.ADPRank != nil {
			adpRank = fmt.Sprintf("%d", *pick.ADPRank)
		}

		fmt.Fprintf(w, "%d,%d,%s,%s,%s,%s,%s\n",
			pick.Round, pick.OverallPick, teamName, playerName, position, teamAbbr, adpRank)
	}
}

// ExportJSON exports draft results as JSON
func (h *Handler) ExportJSON(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	draftID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid draft ID", http.StatusBadRequest)
		return
	}

	draft, err := h.draftRepo.GetByID(draftID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	teams, _ := h.teamRepo.GetByDraft(draftID)
	picks, _ := h.pickRepo.GetByDraft(draftID)

	type ExportPick struct {
		Round       int     `json:"round"`
		OverallPick int     `json:"overall_pick"`
		TeamName    string  `json:"team_name"`
		PlayerName  string  `json:"player_name"`
		Position    string  `json:"position"`
		NFLTeam     string  `json:"nfl_team"`
		ADPRank     *int    `json:"adp_rank"`
	}

	exportPicks := make([]ExportPick, 0, len(picks))
	for _, pick := range picks {
		player, _ := h.playerRepo.GetByID(pick.PlayerID)
		team, _ := h.teamRepo.GetByID(pick.TeamID)

		ep := ExportPick{
			Round:       pick.Round,
			OverallPick: pick.OverallPick,
			ADPRank:     pick.ADPRank,
		}

		if team != nil {
			ep.TeamName = team.TeamName
		}

		if player != nil {
			ep.PlayerName = player.Name
			ep.Position = player.Position
			ep.NFLTeam = player.Team
		}

		exportPicks = append(exportPicks, ep)
	}

	result := map[string]interface{}{
		"draft": map[string]interface{}{
			"id":             draft.ID,
			"name":           draft.Name,
			"num_teams":      draft.NumTeams,
			"scoring_format": draft.ScoringFormat,
			"draft_type":     draft.DraftType,
			"status":         draft.Status,
		},
		"teams": teams,
		"picks": exportPicks,
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=draft-%d.json", draftID))
	json.NewEncoder(w).Encode(result)
}

