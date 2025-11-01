# Fantasy Football Offline Draft Webapp - Implementation Plan

## Project Overview
Convert the Excel-based offline fantasy football draft tracker into a modern webapp using **HTMX**, **Go**, **Templ**, and **SQLite**.

## Current Features (from Excel)
- Support for 8, 10, 12, and 14 team leagues
- Multiple scoring formats: Standard, Half-PPR, PPR
- Dynasty and Redraft rankings
- Player database with 1400+ players
- Team management and draft order configuration
- Live draft board with snake draft support
- Position filtering (QB, RB, WR, TE, K, D/ST, IDP positions)
- Team statistics tracking
- Players drafted list by position
- Big board view showing each team's picks
- Available players list with position filtering

---

## Tech Stack
- **Backend**: Go (Golang)
- **Frontend**: HTMX + Templ templates
- **Database**: SQLite
- **CSS**: TailwindCSS or simple CSS
- **Deployment**: Single binary with embedded assets

---

## Phase 1: Foundation & Database Setup

### 1.1 Database Schema Design
```sql
-- Players table
CREATE TABLE players (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    team TEXT NOT NULL,
    position TEXT NOT NULL,
    bye_week INTEGER,
    dynasty_rank INTEGER,
    sf_rank INTEGER,
    std_rank INTEGER,
    half_ppr_rank INTEGER,
    ppr_rank INTEGER,
    is_custom BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Drafts table (one per draft session)
CREATE TABLE drafts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    num_teams INTEGER NOT NULL CHECK(num_teams IN (8, 10, 12, 14)),
    scoring_format TEXT NOT NULL CHECK(scoring_format IN ('Standard', 'Half-PPR', 'PPR')),
    draft_type TEXT NOT NULL CHECK(draft_type IN ('Redraft', 'Dynasty')),
    qb_setting TEXT DEFAULT '1QB',
    snake_draft BOOLEAN DEFAULT TRUE,
    status TEXT DEFAULT 'setup' CHECK(status IN ('setup', 'active', 'paused', 'completed')),
    max_rounds INTEGER DEFAULT 16,
    commissioner_id TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed BOOLEAN DEFAULT FALSE
);

-- Teams table
CREATE TABLE teams (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    draft_id INTEGER NOT NULL,
    team_name TEXT NOT NULL,
    owner_name TEXT,
    draft_position INTEGER NOT NULL,
    FOREIGN KEY (draft_id) REFERENCES drafts(id) ON DELETE CASCADE
);

-- Picks table (tracks all draft picks)
CREATE TABLE picks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    draft_id INTEGER NOT NULL,
    team_id INTEGER NOT NULL,
    player_id INTEGER NOT NULL,
    round INTEGER NOT NULL,
    overall_pick INTEGER NOT NULL,
    is_traded BOOLEAN DEFAULT FALSE,
    adp_rank INTEGER,
    picked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (draft_id) REFERENCES drafts(id) ON DELETE CASCADE,
    FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE,
    FOREIGN KEY (player_id) REFERENCES players(id),
    UNIQUE(draft_id, player_id) -- Prevent duplicate picks in same draft
);

-- Position settings for filtering
CREATE TABLE position_settings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    draft_id INTEGER NOT NULL,
    position TEXT NOT NULL,
    enabled BOOLEAN DEFAULT TRUE,
    FOREIGN KEY (draft_id) REFERENCES drafts(id) ON DELETE CASCADE
);

-- Player queue/watchlist table
CREATE TABLE draft_queue (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    draft_id INTEGER NOT NULL,
    team_id INTEGER NOT NULL,
    player_id INTEGER NOT NULL,
    queue_order INTEGER NOT NULL,
    added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (draft_id) REFERENCES drafts(id) ON DELETE CASCADE,
    FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE,
    FOREIGN KEY (player_id) REFERENCES players(id),
    UNIQUE(draft_id, team_id, player_id)
);

-- Audit log for tracking actions
CREATE TABLE audit_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    draft_id INTEGER NOT NULL,
    action_type TEXT NOT NULL, -- 'pick', 'undo', 'trade', 'pause', 'resume', 'complete'
    entity_id INTEGER,
    details TEXT, -- JSON with additional info
    performed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (draft_id) REFERENCES drafts(id) ON DELETE CASCADE
);

-- Indexes for performance
CREATE INDEX idx_players_position ON players(position);
CREATE INDEX idx_players_team ON players(team);
CREATE INDEX idx_players_name ON players(name);
CREATE INDEX idx_players_rank ON players(ppr_rank, half_ppr_rank, std_rank, dynasty_rank);
CREATE INDEX idx_picks_draft ON picks(draft_id);
CREATE INDEX idx_picks_team ON picks(team_id);
CREATE INDEX idx_picks_player_draft ON picks(player_id, draft_id);
CREATE INDEX idx_teams_draft ON teams(draft_id);
CREATE INDEX idx_draft_queue_draft ON draft_queue(draft_id);
CREATE INDEX idx_audit_log_draft ON audit_log(draft_id);
```

### 1.2 Project Structure
```
fantasy-draft/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── database/
│   │   ├── db.go                # Database connection & setup
│   │   ├── migrations.go        # Schema migrations
│   │   └── seed.go              # Seed player data from Excel
│   ├── handlers/
│   │   ├── draft.go             # Draft CRUD operations
│   │   ├── pick.go              # Draft pick actions
│   │   ├── team.go              # Team management
│   │   ├── player.go            # Player queries
│   │   └── stats.go             # Statistics endpoints
│   ├── models/
│   │   ├── draft.go
│   │   ├── team.go
│   │   ├── player.go
│   │   ├── pick.go
│   │   └── queue.go
│   ├── repository/
│   │   ├── draft_repo.go
│   │   ├── team_repo.go
│   │   ├── player_repo.go
│   │   ├── pick_repo.go
│   │   └── queue_repo.go
│   ├── validation/
│   │   ├── draft_validation.go  # All validation rules
│   │   ├── team_validation.go
│   │   ├── pick_validation.go
│   │   └── errors.go            # Custom error types
│   └── snake/
│       └── draft_order.go       # Snake draft calculation logic
├── web/
│   ├── templates/
│   │   ├── layout.templ         # Base layout
│   │   ├── home.templ           # Landing page
│   │   ├── draft/
│   │   │   ├── setup.templ      # Draft configuration
│   │   │   ├── board.templ      # Main draft board
│   │   │   ├── big_board.templ  # Compact grid view
│   │   │   ├── player_list.templ
│   │   │   ├── team_picks.templ
│   │   │   ├── stats.templ
│   │   │   ├── stats_franchise.templ  # Stats by NFL team
│   │   │   ├── drafted_by_position.templ
│   │   │   └── value_picks.templ
│   │   └── components/
│   │       ├── player_row.templ
│   │       ├── pick_card.templ
│   │       ├── team_summary.templ
│   │       ├── queue_item.templ
│   │       ├── error.templ
│   │       └── toast.templ
│   └── static/
│       ├── css/
│       │   └── styles.css
│       └── js/
│           └── app.js           # Minimal JS for HTMX enhancements
├── go.mod
├── go.sum
└── README.md
```

### 1.3 Initial Setup Tasks
- [ ] Initialize Go module
- [ ] Set up SQLite database connection
- [ ] Create migration system
- [ ] Extract player data from Excel and seed database
- [ ] Set up Templ for HTML templating
- [ ] Configure basic HTTP server with routes

### 1.4 Data Migration Script
Create a Go script to parse the Excel file and populate the players table:
```go
// Parse HELP_PlayerDB sheet
// Map columns: Dynasty, SF, STD, Half-PPR, PPR, Player, Team, Pos, Bye
// Insert into players table
```

### 1.5 Snake Draft Calculation Algorithm

**Critical Implementation - Must be exact:**

```go
package snake

import (
    "errors"
    "fmt"
    "math"
)

// CalculateCurrentTeam returns the team that should pick for the given pick number
func CalculateCurrentTeam(pickNumber int, numTeams int, teams []Team) (*Team, error) {
    if pickNumber < 1 {
        return nil, errors.New("invalid pick number")
    }
    
    // Calculate which round we're in (1-indexed)
    round := int(math.Ceil(float64(pickNumber) / float64(numTeams)))
    
    // Calculate position within the round (1-indexed)
    positionInRound := ((pickNumber - 1) % numTeams) + 1
    
    // Determine draft position based on snake pattern
    var draftPosition int
    if round%2 == 1 {
        // Odd rounds (1, 3, 5, ...): normal order (1, 2, 3, ...)
        draftPosition = positionInRound
    } else {
        // Even rounds (2, 4, 6, ...): reverse order (N, N-1, N-2, ...)
        draftPosition = numTeams - positionInRound + 1
    }
    
    // Find team with this draft position
    for _, team := range teams {
        if team.DraftPosition == draftPosition {
            return &team, nil
        }
    }
    
    return nil, fmt.Errorf("no team found for draft position %d", draftPosition)
}

// CalculateRound returns the round number for a given pick
func CalculateRound(pickNumber int, numTeams int) int {
    return int(math.Ceil(float64(pickNumber) / float64(numTeams)))
}

// Example:
// 10-team league
// Pick 1: Round 1, Position 1 (odd round) -> Team at position 1
// Pick 10: Round 1, Position 10 (odd round) -> Team at position 10
// Pick 11: Round 2, Position 1 (even round) -> Team at position 10 (reverse)
// Pick 20: Round 2, Position 10 (even round) -> Team at position 1 (reverse)
// Pick 21: Round 3, Position 1 (odd round) -> Team at position 1
```

### 1.6 Draft Completion Detection

```go
package draft

// CheckDraftCompletion determines if draft is complete
func CheckDraftCompletion(draft *Draft, pickCount int) bool {
    // Method 1: Based on configured max rounds
    if draft.MaxRounds > 0 {
        maxPicks := draft.NumTeams * draft.MaxRounds
        return pickCount >= maxPicks
    }
    
    // Method 2: Commissioner manually completes
    if draft.Status == "completed" {
        return true
    }
    
    return false
}

// Auto-complete draft when max picks reached
func (r *PickRepository) Create(pick *Pick) error {
    // Insert pick
    err := r.db.Create(pick)
    if err != nil {
        return err
    }
    
    // Check if draft should auto-complete
    draft, err := r.draftRepo.GetByID(pick.DraftID)
    if err != nil {
        return err
    }
    
    pickCount, err := r.CountByDraft(pick.DraftID)
    if err != nil {
        return err
    }
    
    if CheckDraftCompletion(draft, pickCount) {
        draft.Status = "completed"
        draft.Completed = true
        r.draftRepo.Update(draft)
        
        // Log to audit trail
        r.auditLog.Log(pick.DraftID, "complete", nil, "Draft auto-completed")
    }
    
    return nil
}
```

### 1.7 Validation Package

Create comprehensive validation with all business rules from spec:

```go
package validation

import "errors"

// All validation errors from spec
var (
    ErrInvalidLeagueSize    = errors.New("invalid league size. Must be 8, 10, 12, or 14")
    ErrInvalidScoringFormat = errors.New("invalid scoring format. Must be Standard, Half-PPR, or PPR")
    ErrInvalidDraftType     = errors.New("invalid draft type. Must be Redraft or Dynasty")
    ErrDraftNameRequired    = errors.New("draft name is required")
    ErrTeamNameRequired     = errors.New("team name is required")
    ErrTeamNameTooLong      = errors.New("team name must be between 1 and 50 characters")
    ErrDuplicateTeamName    = errors.New("team name already exists in this draft")
    ErrInvalidDraftPosition = errors.New("draft position must be between 1 and N")
    ErrDuplicateDraftPos    = errors.New("draft position already assigned")
    ErrIncompleteTeamRoster = errors.New("must have exactly N teams")
    ErrInvalidPlayer        = errors.New("invalid player ID")
    ErrPlayerAlreadyDrafted = errors.New("player has already been drafted")
    ErrInvalidTeam          = errors.New("invalid team ID")
    ErrNotTeamTurn          = errors.New("not this team's turn to pick")
    ErrDraftNotActive       = errors.New("cannot make picks in completed draft")
    ErrInvalidPickNumber    = errors.New("pick number must be sequential")
    ErrSearchQueryTooLong   = errors.New("search query too long (max 50 characters)")
    ErrInvalidPosition      = errors.New("invalid position filter")
    ErrInvalidSortOption    = errors.New("invalid sort option")
)

func ValidateDraft(draft *Draft) error {
    if draft.Name == "" {
        return ErrDraftNameRequired
    }
    if draft.NumTeams != 8 && draft.NumTeams != 10 && 
       draft.NumTeams != 12 && draft.NumTeams != 14 {
        return ErrInvalidLeagueSize
    }
    validFormats := map[string]bool{"Standard": true, "Half-PPR": true, "PPR": true}
    if !validFormats[draft.ScoringFormat] {
        return ErrInvalidScoringFormat
    }
    validTypes := map[string]bool{"Redraft": true, "Dynasty": true}
    if !validTypes[draft.DraftType] {
        return ErrInvalidDraftType
    }
    return nil
}

func ValidateTeam(team *Team, existingTeams []Team, numTeams int) error {
    if team.TeamName == "" {
        return ErrTeamNameRequired
    }
    if len(team.TeamName) > 50 {
        return ErrTeamNameTooLong
    }
    if team.DraftPosition < 1 || team.DraftPosition > numTeams {
        return ErrInvalidDraftPosition
    }
    for _, t := range existingTeams {
        if t.TeamName == team.TeamName && t.ID != team.ID {
            return ErrDuplicateTeamName
        }
        if t.DraftPosition == team.DraftPosition && t.ID != team.ID {
            return ErrDuplicateDraftPos
        }
    }
    return nil
}

func ValidatePick(pick *Pick, draft *Draft, teams []Team, pickCount int) error {
    // Validate pick is sequential
    if pick.OverallPick != pickCount+1 {
        return ErrInvalidPickNumber
    }
    
    // Validate correct team's turn
    currentTeam, err := snake.CalculateCurrentTeam(pick.OverallPick, draft.NumTeams, teams)
    if err != nil || currentTeam.ID != pick.TeamID {
        return ErrNotTeamTurn
    }
    
    // Validate draft is active
    if draft.Status != "active" {
        return ErrDraftNotActive
    }
    
    return nil
}
```

---

## Phase 2: Draft Setup & Configuration

### 2.1 Features
- Create new draft
- Configure settings:
  - Number of teams (8/10/12/14)
  - Scoring format (Standard/Half-PPR/PPR)
  - Draft type (Redraft/Dynasty)
  - Position filters (enable/disable positions)
  - **Max rounds** (default 16, configurable)
- Add teams with names and owners
- Set draft order (automatic snake calculation or manual)
- Save draft configuration
- **Pause/Resume draft** functionality
- **Draft URL sharing** for participants

### 2.2 Templates & Routes

**Routes:**
```
GET  /                          # Home page - list all drafts
GET  /draft/new                 # New draft setup form
POST /draft/create              # Create draft
GET  /draft/{id}/setup          # Edit draft settings
POST /draft/{id}/update         # Update draft settings
POST /draft/{id}/start          # Start draft (status -> active)
POST /draft/{id}/pause          # Pause draft
POST /draft/{id}/resume         # Resume draft
POST /draft/{id}/complete       # Manually complete draft
GET  /draft/{id}                # Main draft board
GET  /draft/{id}/big-board      # Big board grid view
DELETE /draft/{id}              # Delete draft
```

**Templates:**
- `home.templ` - List of all drafts with "Create New Draft" button
- `draft/setup.templ` - Multi-step form for draft configuration
- Team entry form with HTMX for dynamic team addition

### 2.3 HTMX Patterns
```html
<!-- Add team dynamically -->
<form hx-post="/draft/{id}/teams" 
      hx-target="#teams-list" 
      hx-swap="beforeend">
    <input name="team_name" placeholder="Team Name" required maxlength="50">
    <input name="owner_name" placeholder="Owner" maxlength="50">
    <button type="submit">Add Team</button>
</form>

<div id="teams-list">
    <!-- Teams will be added here -->
</div>

<!-- Draft URL Sharing -->
<div class="draft-share">
    <label>Share this draft:</label>
    <input type="text" 
           readonly 
           value="https://yourdomain.com/draft/{{.DraftID}}"
           id="draft-url">
    <button onclick="navigator.clipboard.writeText(document.getElementById('draft-url').value)">
        Copy Link
    </button>
</div>

<!-- Pause/Resume Controls -->
{{if eq .Draft.Status "active"}}
<button hx-post="/draft/{{.Draft.ID}}/pause"
        hx-target="#draft-controls"
        hx-swap="outerHTML">
    Pause Draft
</button>
{{else if eq .Draft.Status "paused"}}
<div class="paused-banner">
    ⏸ Draft Paused
</div>
<button hx-post="/draft/{{.Draft.ID}}/resume"
        hx-target="#draft-controls"
        hx-swap="outerHTML">
    Resume Draft
</button>
{{end}}
```

---

## Phase 3: Draft Board - Core Functionality

### 3.1 Draft Board Layout

**Main Sections:**
1. **Draft Controls** (top)
   - Current pick indicator
   - Round/pick number
   - On-the-clock team
   - Undo last pick button

2. **Available Players** (left sidebar, 30%)
   - Search bar
   - Position filters
   - Sortable by rank/name
   - Click to draft
   - Show player details (team, position, bye week, rank)

3. **Draft Board** (center, 40%)
   - Grid showing all picks
   - Column per team
   - Row per round
   - Snake pattern visualization
   - Highlight current pick slot

4. **Team Stats** (right sidebar, 30%)
   - Position breakdown per team
   - Recent picks
   - Team needs indicator

### 3.2 Routes

```
GET  /draft/{id}/board              # Main board view
GET  /draft/{id}/big-board          # Compact grid view (separate)
GET  /draft/{id}/players            # Get available players (filtered)
POST /draft/{id}/pick               # Make a pick
POST /draft/{id}/undo               # Undo last pick
POST /draft/{id}/trade              # Record traded pick
GET  /draft/{id}/teams/{team_id}    # Get team details
GET  /draft/{id}/current            # Get current pick info
GET  /draft/{id}/stats              # Get draft statistics
GET  /draft/{id}/stats/franchise    # Stats by NFL team
GET  /draft/{id}/stats/position     # Players drafted by position
GET  /draft/{id}/drafted-players    # Alternative drafted view
```

### 3.3 HTMX Implementation

**Making a Pick:**
```html
<!-- Player row in available players list -->
<div class="player-row" 
     hx-post="/draft/{id}/pick"
     hx-vals='{"player_id": {{.ID}}}'
     hx-target="#draft-board"
     hx-swap="outerHTML"
     hx-trigger="click">
    <span class="rank">{{.Rank}}</span>
    <span class="name">{{.Name}}</span>
    <span class="position">{{.Position}}</span>
    <span class="team">{{.Team}}</span>
    <span class="bye">Bye: {{.ByeWeek}}</span>
</div>
```

**Auto-refresh Components:**
```html
<!-- Current pick indicator -->
<div hx-get="/draft/{id}/current-pick" 
     hx-trigger="every 2s"
     hx-swap="outerHTML">
    <div class="current-pick">
        <strong>On the Clock:</strong> {{.TeamName}}
        <span>Round {{.Round}}, Pick {{.Pick}}</span>
    </div>
</div>
```

**Player List Filtering:**
```html
<div class="filters">
    <!-- Position checkboxes -->
    <input type="checkbox" name="QB" checked
           hx-get="/draft/{id}/players"
           hx-target="#players-list"
           hx-trigger="change"
           hx-include="[name='position']">
    <!-- Search -->
    <input type="text" name="search"
           hx-get="/draft/{id}/players"
           hx-target="#players-list"
           hx-trigger="keyup changed delay:300ms">
</div>

<div id="players-list">
    <!-- Player rows here -->
</div>
```

### 3.4 Pick Logic

**Go Handler:**
```go
func (h *DraftHandler) MakePick(w http.ResponseWriter, r *http.Request) {
    // 1. Parse draft_id and player_id
    draftID := chi.URLParam(r, "id")
    playerID := r.FormValue("player_id")
    
    // 2. Get draft and validate
    draft, err := h.draftRepo.GetByID(draftID)
    if err != nil {
        return component.Error("Draft not found").Render(r.Context(), w)
    }
    
    // 3. Get current pick number
    pickCount, _ := h.pickRepo.CountByDraft(draftID)
    currentPickNumber := pickCount + 1
    
    // 4. Calculate which team's turn
    teams, _ := h.teamRepo.GetByDraft(draftID)
    currentTeam, err := snake.CalculateCurrentTeam(currentPickNumber, draft.NumTeams, teams)
    if err != nil {
        return component.Error("Cannot determine current team").Render(r.Context(), w)
    }
    
    // 5. Get player and ADP rank
    player, _ := h.playerRepo.GetByID(playerID)
    adpRank := getADPRank(player, draft.DraftType, draft.ScoringFormat)
    
    // 6. Calculate round number
    round := snake.CalculateRound(currentPickNumber, draft.NumTeams)
    
    // 7. Create pick
    pick := &Pick{
        DraftID:     draft.ID,
        TeamID:      currentTeam.ID,
        PlayerID:    player.ID,
        Round:       round,
        OverallPick: currentPickNumber,
        ADPRank:     adpRank,
    }
    
    // 8. Validate pick
    if err := validation.ValidatePick(pick, draft, teams, pickCount); err != nil {
        return component.Error(err.Error()).Render(r.Context(), w)
    }
    
    // 9. Insert pick (includes auto-complete check)
    if err := h.pickRepo.Create(pick); err != nil {
        return component.Error("Failed to record pick").Render(r.Context(), w)
    }
    
    // 10. Log to audit trail
    h.auditLog.Log(draft.ID, "pick", pick.ID, fmt.Sprintf("%s drafted by %s", player.Name, currentTeam.TeamName))
    
    // 11. Return updated board + success toast
    return templ.Handler(
        component.DraftBoard(updatedBoard),
        component.Toast("success", fmt.Sprintf("%s drafted!", player.Name)),
    ).ServeHTTP(w, r)
}

func getADPRank(player *Player, draftType, scoringFormat string) int {
    if draftType == "Dynasty" {
        return player.DynastyRank
    }
    switch scoringFormat {
    case "PPR":
        return player.PPRRank
    case "Half-PPR":
        return player.HalfPPRRank
    default:
        return player.StdRank
    }
}
```

### 3.5 Big Board View (Separate from Main Board)

The Big Board is a **compact grid layout** showing all picks in a table format.

**Template: `templates/draft/big_board.templ`**

```html
<div class="big-board-view">
    <h2>Big Board</h2>
    <button onclick="window.print()" class="no-print">Print</button>
    
    <table class="big-board-grid">
        <thead>
            <tr>
                <th>Round</th>
                {{range .Teams}}
                <th class="team-header">
                    <div class="team-name">{{.TeamName}}</div>
                    <div class="owner-name">{{.OwnerName}}</div>
                    <div class="position-counts">
                        {{.PositionSummary}}
                    </div>
                </th>
                {{end}}
            </tr>
        </thead>
        <tbody>
            {{range .Rounds}}
            <tr>
                <td class="round-label">Round {{.Number}}</td>
                {{range .Picks}}
                <td class="pick-cell {{if .IsTraded}}traded{{end}}">
                    {{if .Player}}
                        <div class="player-name">{{.Player.Name}}</div>
                        <div class="player-details">
                            <span class="position-badge">{{.Player.Position}}</span>
                            <span class="team-abbr">{{.Player.Team}}</span>
                        </div>
                        {{if .IsTraded}}<span class="traded-badge">TRADED</span>{{end}}
                    {{else}}
                        <span class="empty-pick">-</span>
                    {{end}}
                </td>
                {{end}}
            </tr>
            {{end}}
        </tbody>
    </table>
</div>

<style media="print">
    .no-print { display: none; }
    .big-board-grid { 
        font-size: 9pt;
        page-break-after: always;
    }
    @page { 
        size: landscape;
        margin: 0.5in;
    }
</style>
```

### 3.6 Stats by Franchise View

Shows players drafted organized by NFL team.

**Template: `templates/draft/stats_franchise.templ`**

```html
<div class="stats-franchise">
    <h2>Players Drafted by NFL Team</h2>
    
    <table class="franchise-stats-table">
        <thead>
            <tr>
                <th>NFL Team</th>
                <th>Total Drafted</th>
                <th>QB</th>
                <th>RB</th>
                <th>WR</th>
                <th>TE</th>
                <th>Other</th>
            </tr>
        </thead>
        <tbody>
            {{range .FranchiseStats}}
            <tr>
                <td>
                    <strong>{{.TeamAbbr}}</strong>
                    <span class="team-full-name">{{.TeamFullName}}</span>
                </td>
                <td>{{.TotalCount}}</td>
                <td>{{.QBCount}}</td>
                <td>{{.RBCount}}</td>
                <td>{{.WRCount}}</td>
                <td>{{.TECount}}</td>
                <td>{{.OtherCount}}</td>
            </tr>
            {{end}}
        </tbody>
    </table>
</div>
```

**Handler:**
```go
func (h *StatsHandler) GetFranchiseStats(w http.ResponseWriter, r *http.Request) {
    draftID := chi.URLParam(r, "id")
    
    // Query: SELECT players.team, players.position, COUNT(*) 
    //        FROM picks JOIN players ON picks.player_id = players.id
    //        WHERE picks.draft_id = ?
    //        GROUP BY players.team, players.position
    
    stats := h.statsRepo.GetFranchiseStats(draftID)
    
    return component.StatsFranchise(stats).Render(r.Context(), w)
}
```

### 3.7 Players Drafted by Position View

Shows all drafted players organized by position (QB column, RB column, etc.).

**Template: `templates/draft/drafted_by_position.templ`**

```html
<div class="drafted-by-position">
    <div class="position-summary">
        <span>QB: {{.QBCount}}/{{.TotalQB}}</span>
        <span>RB: {{.RBCount}}/{{.TotalRB}}</span>
        <span>WR: {{.WRCount}}/{{.TotalWR}}</span>
        <span>TE: {{.TECount}}/{{.TotalTE}}</span>
        <span>K: {{.KCount}}/{{.TotalK}}</span>
        <span>D/ST: {{.DSTCount}}/{{.TotalDST}}</span>
    </div>
    
    <div class="position-columns">
        <div class="position-column">
            <h3>Quarterbacks ({{.QBCount}})</h3>
            {{range .QBDrafted}}
                <div class="drafted-player">
                    {{.PlayerName}} - {{.TeamName}}
                    <span class="pick-number">Pick {{.OverallPick}}</span>
                </div>
            {{end}}
        </div>
        
        <div class="position-column">
            <h3>Running Backs ({{.RBCount}})</h3>
            {{range .RBDrafted}}
                <div class="drafted-player">
                    {{.PlayerName}} - {{.TeamName}}
                    <span class="pick-number">Pick {{.OverallPick}}</span>
                </div>
            {{end}}
        </div>
        
        <!-- Repeat for WR, TE, K, D/ST -->
    </div>
</div>
```

---

## Phase 4: Real-time Updates & Enhancements

### 4.1 Features
- Live draft board updates (simulate multiple users)
- **Player queue/watchlist** per team
- Export draft results (CSV, PDF)
- Draft timer (optional countdown per pick)
- Pick notifications
- **Trade pick recording**
- Player comparison view

### 4.2 HTMX Server-Sent Events (SSE)
```html
<!-- Auto-update board for all viewers -->
<div hx-sse="connect:/draft/{id}/stream">
    <div hx-sse="swap:pick-made" 
         hx-swap="afterbegin"
         id="draft-board">
        <!-- Board updates here -->
    </div>
</div>
```

**Go SSE Handler:**
```go
func (h *DraftHandler) StreamUpdates(w http.ResponseWriter, r *http.Request) {
    // Set SSE headers
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")
    
    // Listen for draft updates
    // Send events when picks are made
}
```

### 4.3 Player Queue/Watchlist

Each team can maintain a queue of players they're interested in drafting.

**Routes:**
```
POST   /draft/{id}/queue              # Add player to queue
GET    /draft/{id}/queue              # Get team's queue
DELETE /draft/{id}/queue/{queue_id}   # Remove from queue
PUT    /draft/{id}/queue/reorder      # Reorder queue
```

**Template Component:**
```html
<div id="player-queue" class="queue-sidebar">
    <h3>Your Queue</h3>
    <p class="queue-help">Add players you're targeting</p>
    
    <div class="queue-list" id="queue-items">
        {{range .QueuedPlayers}}
        <div class="queue-item" data-queue-id="{{.ID}}">
            <div class="player-info">
                <strong>{{.Player.Name}}</strong>
                <span class="player-meta">
                    {{.Player.Position}} - {{.Player.Team}}
                    Bye: {{.Player.ByeWeek}}
                </span>
            </div>
            <button class="remove-btn"
                    hx-delete="/draft/{{$.DraftID}}/queue/{{.ID}}"
                    hx-target="#player-queue"
                    hx-swap="outerHTML">
                ×
            </button>
        </div>
        {{end}}
        
        {{if eq (len .QueuedPlayers) 0}}
        <p class="empty-queue">No players in queue</p>
        {{end}}
    </div>
    
    <button class="draft-from-queue"
            {{if gt (len .QueuedPlayers) 0}}
            hx-post="/draft/{{$.DraftID}}/pick"
            hx-vals='{"player_id": "{{(index .QueuedPlayers 0).PlayerID}}"}'
            hx-target="#draft-board"
            {{else}}
            disabled
            {{end}}>
        Draft Next in Queue
    </button>
</div>
```

**Add to Queue Button (on player rows):**
```html
<button class="add-queue-btn"
        hx-post="/draft/{{.DraftID}}/queue"
        hx-vals='{"player_id": "{{.Player.ID}}"}'
        hx-target="#player-queue"
        hx-swap="outerHTML">
    + Queue
</button>
```

**Handler:**
```go
func (h *QueueHandler) AddToQueue(w http.ResponseWriter, r *http.Request) {
    draftID := chi.URLParam(r, "id")
    playerID := r.FormValue("player_id")
    teamID := r.FormValue("team_id") // From session/cookie
    
    // Get current max order
    maxOrder := h.queueRepo.GetMaxOrder(draftID, teamID)
    
    queueItem := &QueueItem{
        DraftID:    draftID,
        TeamID:     teamID,
        PlayerID:   playerID,
        QueueOrder: maxOrder + 1,
    }
    
    err := h.queueRepo.Create(queueItem)
    if err != nil {
        return component.Error("Failed to add to queue").Render(r.Context(), w)
    }
    
    // Return updated queue sidebar
    queue := h.queueRepo.GetByTeam(draftID, teamID)
    return component.PlayerQueue(queue).Render(r.Context(), w)
}
```

### 4.4 Trade Pick Recording

Allow commissioners to record traded picks.

**UI:**
- Right-click or button on pick cell
- Modal dialog to select new team
- "TRADED" badge displayed on pick

**Route:**
```
POST /draft/{id}/trade
```

**Request Body:**
```json
{
    "pick_id": 42,
    "from_team_id": 3,
    "to_team_id": 7,
    "notes": "Round 3 pick swap"
}
```

**Handler:**
```go
func (h *DraftHandler) TradePick(w http.ResponseWriter, r *http.Request) {
    var req struct {
        PickID     int    `json:"pick_id"`
        FromTeamID int    `json:"from_team_id"`
        ToTeamID   int    `json:"to_team_id"`
        Notes      string `json:"notes"`
    }
    
    json.NewDecoder(r.Body).Decode(&req)
    
    // Validate commissioner permission (check session/cookie)
    if !isCommissioner(r) {
        w.WriteHeader(http.StatusForbidden)
        return component.Error("Only commissioner can trade picks").Render(r.Context(), w)
    }
    
    // Update pick
    pick, _ := h.pickRepo.GetByID(req.PickID)
    pick.TeamID = req.ToTeamID
    pick.IsTraded = true
    
    err := h.pickRepo.Update(pick)
    if err != nil {
        return component.Error("Failed to trade pick").Render(r.Context(), w)
    }
    
    // Log trade
    h.auditLog.Log(pick.DraftID, "trade", pick.ID, 
        fmt.Sprintf("Pick %d traded from team %d to team %d: %s", 
                    pick.OverallPick, req.FromTeamID, req.ToTeamID, req.Notes))
    
    return component.Toast("success", "Pick traded successfully").Render(r.Context(), w)
}
```

**Modal Template:**
```html
<div class="modal" id="trade-modal">
    <div class="modal-content">
        <h3>Trade Pick {{.Pick.OverallPick}}</h3>
        
        <form hx-post="/draft/{{.DraftID}}/trade"
              hx-target="#draft-board">
            <input type="hidden" name="pick_id" value="{{.Pick.ID}}">
            <input type="hidden" name="from_team_id" value="{{.Pick.TeamID}}">
            
            <label>Trade to:</label>
            <select name="to_team_id" required>
                {{range .Teams}}
                <option value="{{.ID}}">{{.TeamName}}</option>
                {{end}}
            </select>
            
            <label>Notes (optional):</label>
            <input type="text" name="notes" placeholder="e.g., Round 3 swap">
            
            <button type="submit">Confirm Trade</button>
            <button type="button" onclick="closeModal()">Cancel</button>
        </form>
    </div>
</div>
```

---

## Phase 5: Analytics & Reporting

### 5.1 Features
- Team roster summary
- Position breakdown (bar charts)
- Best available players by position
- **Draft grades** (based on ADP vs. pick number)
- **Value picks analysis** (steals and reaches)
- Bye week analysis
- Team strength comparison
- **Stats by NFL franchise**
- **Players drafted by position**

### 5.2 Templates
- `draft/stats.templ` - Comprehensive statistics page
- `draft/team_detail.templ` - Individual team analysis
- `draft/stats_franchise.templ` - Stats by NFL team
- `draft/drafted_by_position.templ` - Players drafted organized by position
- `draft/value_picks.templ` - Draft steals and reaches
- `draft/bye_weeks.templ` - Bye week analysis
- `draft/export.templ` - Export options

### 5.3 Routes
```
GET /draft/{id}/stats                  # Draft statistics dashboard
GET /draft/{id}/stats/teams            # Stats by fantasy team
GET /draft/{id}/stats/franchise        # Stats by NFL team
GET /draft/{id}/stats/position         # Players drafted by position
GET /draft/{id}/team/{team_id}         # Team detail page
GET /draft/{id}/value-picks            # Value picks report
GET /draft/{id}/bye-weeks              # Bye week analysis
GET /draft/{id}/export/csv             # Export as CSV
GET /draft/{id}/export/pdf             # Export as PDF
GET /draft/{id}/export/json            # Export as JSON
```

### 5.4 Value Picks Analysis

Track and display picks that were significantly above or below ADP.

**Template: `templates/draft/value_picks.templ`**

```html
<div class="value-picks-analysis">
    <h2>Draft Value Analysis</h2>
    
    <div class="analysis-sections">
        <section class="steals">
            <h3>🎯 Steals (Drafted Below ADP)</h3>
            <p>Players drafted later than expected</p>
            
            <table>
                <thead>
                    <tr>
                        <th>Player</th>
                        <th>Team</th>
                        <th>ADP</th>
                        <th>Picked</th>
                        <th>Value</th>
                    </tr>
                </thead>
                <tbody>
                    {{range .Steals}}
                    <tr>
                        <td>{{.Player.Name}} ({{.Player.Position}})</td>
                        <td>{{.Team.TeamName}}</td>
                        <td>{{.ADPRank}}</td>
                        <td>{{.OverallPick}}</td>
                        <td class="positive">+{{.ValueDiff}}</td>
                    </tr>
                    {{end}}
                </tbody>
            </table>
        </section>
        
        <section class="reaches">
            <h3>📈 Reaches (Drafted Above ADP)</h3>
            <p>Players drafted earlier than expected</p>
            
            <table>
                <thead>
                    <tr>
                        <th>Player</th>
                        <th>Team</th>
                        <th>ADP</th>
                        <th>Picked</th>
                        <th>Reach</th>
                    </tr>
                </thead>
                <tbody>
                    {{range .Reaches}}
                    <tr>
                        <td>{{.Player.Name}} ({{.Player.Position}})</td>
                        <td>{{.Team.TeamName}}</td>
                        <td>{{.ADPRank}}</td>
                        <td>{{.OverallPick}}</td>
                        <td class="negative">-{{.ValueDiff}}</td>
                    </tr>
                    {{end}}
                </tbody>
            </table>
        </section>
    </div>
    
    <section class="team-grades">
        <h3>Team Draft Grades</h3>
        {{range .TeamGrades}}
        <div class="team-grade">
            <strong>{{.TeamName}}</strong>
            <span class="grade">Grade: {{.Grade}}</span>
            <span class="avg-value">Avg Value: {{.AvgValue}}</span>
        </div>
        {{end}}
    </section>
</div>
```

**Handler:**
```go
func (h *StatsHandler) GetValuePicks(w http.ResponseWriter, r *http.Request) {
    draftID := chi.URLParam(r, "id")
    
    picks := h.pickRepo.GetByDraft(draftID)
    
    var steals, reaches []PickWithValue
    
    for _, pick := range picks {
        if pick.ADPRank == 0 {
            continue // No ADP data
        }
        
        valueDiff := pick.ADPRank - pick.OverallPick
        
        if valueDiff >= 10 { // Drafted 10+ spots later than ADP
            steals = append(steals, PickWithValue{
                Pick:      pick,
                ValueDiff: valueDiff,
            })
        } else if valueDiff <= -10 { // Drafted 10+ spots earlier
            reaches = append(reaches, PickWithValue{
                Pick:      pick,
                ValueDiff: -valueDiff,
            })
        }
    }
    
    // Sort by value
    sort.Slice(steals, func(i, j int) bool {
        return steals[i].ValueDiff > steals[j].ValueDiff
    })
    sort.Slice(reaches, func(i, j int) bool {
        return reaches[i].ValueDiff > reaches[j].ValueDiff
    })
    
    // Calculate team grades
    teamGrades := calculateTeamGrades(picks)
    
    data := ValuePicksData{
        Steals:     steals,
        Reaches:    reaches,
        TeamGrades: teamGrades,
    }
    
    return component.ValuePicks(data).Render(r.Context(), w)
}
```

### 5.5 Bye Week Analysis

**Template: `templates/draft/bye_weeks.templ`**

```html
<div class="bye-week-analysis">
    <h2>Bye Week Analysis</h2>
    
    {{range .Teams}}
    <section class="team-bye-analysis">
        <h3>{{.TeamName}}</h3>
        
        <div class="bye-week-grid">
            {{range $week := iterate 1 18}}
            <div class="bye-week-column">
                <div class="week-header">Week {{$week}}</div>
                {{$players := playersWithBye $.TeamPlayers $week}}
                {{if gt (len $players) 0}}
                <div class="players-on-bye {{if gt (len $players) 2}}conflict{{end}}">
                    {{range $players}}
                    <div class="player-bye">
                        {{.Name}} ({{.Position}})
                    </div>
                    {{end}}
                </div>
                {{else}}
                <div class="no-bye">-</div>
                {{end}}
            </div>
            {{end}}
        </div>
        
        {{if hasConflicts .ByeWeeks}}
        <div class="warning">
            ⚠️ Multiple starters on bye same week
        </div>
        {{end}}
    </section>
    {{end}}
</div>
```

---

## Phase 6: Polish & Deployment

### 6.1 UI/UX Improvements
- Responsive design (mobile support)
- Dark mode toggle
- Keyboard shortcuts (arrow keys for navigation)
- Print-friendly draft board
- Loading states and animations
- Error handling and validation
- Toast notifications for actions
- **Custom player addition form**
- **Draft URL sharing with copy button**
- **Commissioner controls** (highlighted)

### 6.2 Performance Optimization
- Database indexing (already in schema)
- Query optimization
- Template caching
- Static asset compression
- Connection pooling
- **Pagination for large drafts** (420+ picks)
- **Caching strategy** for available players
- **Optimized queries** with proper indexes

### 6.3 Deployment
- Build single binary with embedded assets
- Docker container option
- Environment configuration
- Database backup/restore
- Health check endpoint

### 6.4 Documentation
- User guide
- API documentation
- Deployment guide
- Development setup instructions

---

## Implementation Timeline

### Week 1: Foundation
- Set up project structure
- Database schema and migrations
- Seed player data
- Basic routing and templating

### Week 2: Draft Setup
- Create draft forms
- Team management
- Draft configuration
- Save/load drafts

### Week 3: Core Draft Board
- Player list with filtering
- Draft board grid
- Pick logic (snake draft)
- Basic HTMX interactions

### Week 4: Real-time Features
- SSE for live updates
- Undo functionality
- Player queue
- Current pick indicator

### Week 5: Analytics & Polish
- Statistics pages
- Team summaries
- Export functionality
- UI refinements

### Week 6: Testing & Deployment
- Integration testing
- Performance testing
- Bug fixes
- Deployment setup
- Documentation

---

## Key HTMX Patterns to Use

### 1. Out-of-Band Swaps (OOB)
Update multiple parts of the page from a single request:
```html
<!-- Main response -->
<div id="draft-board">...</div>

<!-- OOB updates -->
<div id="team-stats" hx-swap-oob="true">...</div>
<div id="available-count" hx-swap-oob="true">...</div>
```

### 2. Inline Editing
```html
<div hx-get="/draft/{id}/pick/{pick_id}/edit"
     hx-trigger="click"
     hx-swap="outerHTML">
    {{.PlayerName}}
</div>
```

### 3. Optimistic UI
Show immediate feedback before server confirmation:
```html
<div hx-post="/draft/{id}/pick"
     hx-on="htmx:beforeRequest: this.classList.add('picking')"
     hx-on="htmx:afterRequest: this.classList.remove('picking')">
```

### 4. Debounced Search
```html
<input type="search"
       hx-get="/draft/{id}/players/search"
       hx-trigger="keyup changed delay:300ms"
       hx-target="#search-results">
```

---

## Go Code Architecture

### Repository Pattern
```go
type DraftRepository interface {
    Create(draft *Draft) error
    GetByID(id int) (*Draft, error)
    Update(draft *Draft) error
    Delete(id int) error
    List() ([]*Draft, error)
}

type PlayerRepository interface {
    GetAvailable(draftID int, filters PlayerFilters) ([]*Player, error)
    GetByID(id int) (*Player, error)
    Search(query string) ([]*Player, error)
}

type PickRepository interface {
    Create(pick *Pick) error
    GetByDraft(draftID int) ([]*Pick, error)
    GetLast(draftID int) (*Pick, error)
    Delete(id int) error
}
```

### Handler Organization
```go
type DraftHandler struct {
    draftRepo  repository.DraftRepository
    playerRepo repository.PlayerRepository
    pickRepo   repository.PickRepository
    teamRepo   repository.TeamRepository
}

func (h *DraftHandler) RegisterRoutes(r *chi.Mux) {
    r.Get("/draft/{id}/board", h.GetBoard)
    r.Post("/draft/{id}/pick", h.MakePick)
    r.Post("/draft/{id}/undo", h.UndoPick)
    // ... other routes
}
```

---

## Testing Strategy

### Unit Tests
- Repository methods
- Draft logic (snake calculation)
- Pick validation
- Rank calculations

### Integration Tests
- HTTP handlers
- Database operations
- HTMX responses

### E2E Tests
- Complete draft flow
- Multi-team scenarios
- Undo/redo operations

---

## Future Enhancements (Post-MVP)

1. **Authentication & Multi-user Support**
   - User accounts
   - Draft permissions
   - Multiple commissioners

2. **Mock Draft Mode**
   - Simulate picks
   - Reset and try again
   - Save draft scenarios

3. **Mobile App**
   - Native iOS/Android
   - Or PWA with offline support

4. **Advanced Analytics**
   - Machine learning for pick suggestions
   - Historical draft data
   - Player projections

5. **Integrations**
   - Import from ESPN/Yahoo
   - Export to league platforms
   - API for external tools

---

## Resources & Libraries

### Go Libraries
- `github.com/mattn/go-sqlite3` - SQLite driver
- `github.com/a-h/templ` - Type-safe HTML templates
- `github.com/go-chi/chi` - HTTP router
- `github.com/golang-migrate/migrate` - Database migrations
- `github.com/jmoiron/sqlx` - SQL extensions

### Frontend
- HTMX - https://htmx.org/
- TailwindCSS (optional) - https://tailwindcss.com/
- AlpineJS (minimal JS if needed)

### Development Tools
- Air - Live reload for Go
- sqlc (optional) - Generate type-safe Go from SQL
- templ CLI - Template generation

---

## Success Criteria

### MVP (Phases 1-3)
- ✅ Create and configure drafts
- ✅ Add teams and set draft order
- ✅ View available players with filtering
- ✅ Make picks with snake draft logic
- ✅ View draft board in real-time
- ✅ Basic statistics per team

### Full Release (Phases 1-6)
- ✅ All MVP features
- ✅ Live updates across multiple viewers
- ✅ Undo/redo picks
- ✅ Player queue/watchlist
- ✅ Export draft results
- ✅ Comprehensive analytics
- ✅ Mobile-responsive design
- ✅ Easy deployment (single binary)

---

## Getting Started

### Development Setup
```bash
# Clone repository
git clone <repo-url>
cd fantasy-draft

# Install dependencies
go mod download

# Install templ CLI
go install github.com/a-h/templ/cmd/templ@latest

# Generate templates
templ generate

# Run migrations
go run cmd/migrate/main.go up

# Seed database from Excel
go run cmd/seed/main.go

# Run development server
go run cmd/server/main.go

# Visit http://localhost:8080
```

### Building for Production
```bash
# Generate templates
templ generate

# Build binary
go build -o fantasy-draft cmd/server/main.go

# Run
./fantasy-draft
```

---

## Summary

This plan provides a comprehensive roadmap for building a modern fantasy football draft webapp using HTMX, Go, Templ, and SQLite. The phased approach allows for incremental development and testing, with each phase building on the previous one. The use of HTMX minimizes JavaScript while providing a rich, interactive user experience. Go's simplicity and performance make it ideal for the backend, and SQLite provides a lightweight, embedded database solution perfect for this use case.

The resulting webapp will be easy to deploy (single binary), performant, and maintainable, while preserving all the functionality of the original Excel workbook and adding modern web capabilities like real-time updates and multi-user support.
