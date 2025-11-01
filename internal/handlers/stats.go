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

	var steals, reaches, allPicks []ValuePick
	picksWithoutADP := 0

	for _, pick := range picks {
		player, err := h.playerRepo.GetByID(pick.PlayerID)
		if err != nil {
			continue
		}

		team, err := h.teamRepo.GetByID(pick.TeamID)
		if err != nil {
			continue
		}

		// Skip picks without ADP rank
		if pick.ADPRank == nil {
			picksWithoutADP++
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

		// Add to all picks list
		allPicks = append(allPicks, vp)

		// Steal: drafted 5+ spots later than ADP (positive valueDiff means picked later = better value)
		if valueDiff >= 5 {
			steals = append(steals, vp)
		}
		// Reach: drafted 5+ spots earlier than ADP (negative valueDiff means picked earlier = reach)
		if valueDiff <= -5 {
			reaches = append(reaches, ValuePick{
				PlayerName:  vp.PlayerName,
				TeamName:    vp.TeamName,
				ADPRank:     vp.ADPRank,
				OverallPick: vp.OverallPick,
				ValueDiff:   -vp.ValueDiff, // Make positive for display
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
	`)

	if len(picks) == 0 {
		content.WriteString(`
			<div class="bg-tokyo-night-bg-light rounded-lg p-8 text-center border border-tokyo-night-border">
				<p class="text-tokyo-night-fg-dim">No picks have been made yet.</p>
			</div>
		`)
		renderTemplate(w, content.String(), "Value Picks Analysis")
		return
	}

	if picksWithoutADP == len(picks) {
		content.WriteString(`
			<div class="bg-tokyo-night-bg-light rounded-lg p-8 border border-tokyo-night-border mb-6">
				<p class="text-tokyo-night-fg-dim mb-2">⚠️ No ADP data available for the drafted players.</p>
				<p class="text-sm text-tokyo-night-fg-dim">Value picks analysis requires players to have ADP rankings. Make sure players are imported with rank data (std_rank, half_ppr_rank, ppr_rank, or dynasty_rank).</p>
			</div>
		`)
	}

	if picksWithoutADP > 0 && picksWithoutADP < len(picks) {
		content.WriteString(fmt.Sprintf(`
			<div class="bg-tokyo-night-bg-light rounded-lg p-4 border border-tokyo-night-border mb-6">
				<p class="text-sm text-tokyo-night-fg-dim">Note: %d out of %d picks don't have ADP data and are excluded from this analysis.</p>
			</div>
		`, picksWithoutADP, len(picks)))
	}

	if len(steals) == 0 && len(reaches) == 0 && len(allPicks) > 0 {
		// Show all picks sorted by value difference if no significant steals/reaches
		content.WriteString(`
			<div class="mb-4">
				<p class="text-tokyo-night-fg-dim">No significant steals (5+ spots) or reaches (5+ spots) found. Showing all picks sorted by value:</p>
			</div>
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
		// Sort by valueDiff descending (best steals first)
		for i := 0; i < len(allPicks); i++ {
			for j := i + 1; j < len(allPicks); j++ {
				if allPicks[i].ValueDiff < allPicks[j].ValueDiff {
					allPicks[i], allPicks[j] = allPicks[j], allPicks[i]
				}
			}
		}
		for _, vp := range allPicks {
			valueColor := "text-tokyo-night-fg-dim"
			valueSign := ""
			if vp.ValueDiff > 0 {
				valueColor = "text-tokyo-night-success"
				valueSign = "+"
			} else if vp.ValueDiff < 0 {
				valueColor = "text-tokyo-night-error"
			}
			content.WriteString(fmt.Sprintf(`
				<tr class="hover:bg-tokyo-night-bg-dark transition-colors">
					<td class="px-4 py-2 border-b border-tokyo-night-border font-medium text-tokyo-night-fg">%s (%s)</td>
					<td class="px-4 py-2 border-b border-tokyo-night-border text-tokyo-night-fg-dim">%s</td>
					<td class="px-4 py-2 border-b border-tokyo-night-border text-tokyo-night-fg-dim">%d</td>
					<td class="px-4 py-2 border-b border-tokyo-night-border text-tokyo-night-fg-dim">%d</td>
					<td class="px-4 py-2 border-b border-tokyo-night-border %s font-semibold">%s%d</td>
				</tr>
			`, vp.PlayerName, vp.Position, vp.TeamName, vp.ADPRank, vp.OverallPick, valueColor, valueSign, vp.ValueDiff))
		}
		content.WriteString(`</tbody></table></div>`)
		renderTemplate(w, content.String(), "Value Picks Analysis")
		return
	}

	content.WriteString(`<div class="grid md:grid-cols-2 gap-8">
		<div>
			<h2 class="text-2xl font-semibold mb-4 text-tokyo-night-success">Steals (Drafted 5+ Spots Below ADP)</h2>
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

	// Sort steals by valueDiff descending
	for i := 0; i < len(steals); i++ {
		for j := i + 1; j < len(steals); j++ {
			if steals[i].ValueDiff < steals[j].ValueDiff {
				steals[i], steals[j] = steals[j], steals[i]
			}
		}
	}

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
		content.WriteString(`<tr><td colspan="5" class="px-4 py-8 text-center text-tokyo-night-fg-dim">No steals found (5+ spots)</td></tr>`)
	}
	content.WriteString(`</tbody></table></div></div>`)

	content.WriteString(`
		<div>
			<h2 class="text-2xl font-semibold mb-4 text-tokyo-night-error">Reaches (Drafted 5+ Spots Above ADP)</h2>
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

	// Sort reaches by ValueDiff descending (biggest reaches first)
	for i := 0; i < len(reaches); i++ {
		for j := i + 1; j < len(reaches); j++ {
			if reaches[i].ValueDiff < reaches[j].ValueDiff {
				reaches[i], reaches[j] = reaches[j], reaches[i]
			}
		}
	}

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
		content.WriteString(`<tr><td colspan="5" class="px-4 py-8 text-center text-tokyo-night-fg-dim">No reaches found (5+ spots)</td></tr>`)
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

