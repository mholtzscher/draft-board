# Implementation Plan Gap Analysis & Updates

**Date:** 2025-01-XX  
**Analysis:** Comparison of Implementation Plan vs. Specification Document

---

## Executive Summary

After comparing the Implementation Plan against the detailed Specification Document, **61 gaps** were identified across 8 categories. These gaps range from missing database fields to entire views and workflows not accounted for in the original plan.

**Breakdown:**
- 16 Missing Features
- 7 Schema Gaps
- 12 Route/API Gaps
- 6 Critical Business Logic Gaps
- 10 UI/UX Gaps
- 5 Validation Gaps
- 5 Performance/Scale Gaps

**Recommendation:** Integrate these gaps into the implementation plan across all 6 phases.

---

## 1. CRITICAL GAPS (Must Fix Immediately)

### 1.1 Database Schema Updates

**Add to Phase 1 - Foundation:**

```sql
-- Update drafts table
ALTER TABLE drafts ADD COLUMN status TEXT DEFAULT 'setup' 
    CHECK(status IN ('setup', 'active', 'paused', 'completed'));
ALTER TABLE drafts ADD COLUMN max_rounds INTEGER DEFAULT 16;
ALTER TABLE drafts ADD COLUMN commissioner_id TEXT; -- Simple UUID or session ID

-- Update picks table
ALTER TABLE picks ADD COLUMN is_traded BOOLEAN DEFAULT FALSE;
ALTER TABLE picks ADD COLUMN adp_rank INTEGER; -- ADP at time of pick

-- Update players table
ALTER TABLE players ADD COLUMN is_custom BOOLEAN DEFAULT FALSE;

-- Create draft_queue table (watchlist)
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

-- Create audit_log table (optional but recommended)
CREATE TABLE audit_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    draft_id INTEGER NOT NULL,
    action_type TEXT NOT NULL, -- 'pick', 'undo', 'trade', 'pause', 'resume'
    entity_id INTEGER, -- pick_id or team_id
    details TEXT, -- JSON with additional info
    performed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (draft_id) REFERENCES drafts(id) ON DELETE CASCADE
);

-- Indexes
CREATE INDEX idx_draft_queue_draft ON draft_queue(draft_id);
CREATE INDEX idx_audit_log_draft ON audit_log(draft_id);
```

### 1.2 Snake Draft Formula (Missing from Plan)

**Add to Phase 3 - Core Functionality:**

```go
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
        // Odd rounds: normal order (1, 2, 3, ...)
        draftPosition = positionInRound
    } else {
        // Even rounds: reverse order (N, N-1, N-2, ...)
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
```

### 1.3 Draft Completion Detection

**Add to Phase 3:**

```go
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

// Auto-update status when max picks reached
func (r *PickRepository) Create(pick *Pick) error {
    // ... insert pick ...
    
    // Check if draft is now complete
    draft, _ := r.draftRepo.GetByID(pick.DraftID)
    pickCount, _ := r.CountByDraft(pick.DraftID)
    
    if CheckDraftCompletion(draft, pickCount) {
        draft.Status = "completed"
        draft.Completed = true
        r.draftRepo.Update(draft)
    }
    
    return nil
}
```

---

## 2. MISSING VIEWS & FEATURES

### 2.1 Stats by Franchise View

**Add to Phase 5 - Analytics:**

**Template: `templates/draft/stats_franchise.templ`**

Shows how many players drafted from each NFL team.

```html
<div class="stats-franchise">
    <h2>Players Drafted by NFL Team</h2>
    <table>
        <thead>
            <tr>
                <th>NFL Team</th>
                <th>Players Drafted</th>
                <th>Positions</th>
            </tr>
        </thead>
        <tbody>
            {{range .FranchiseStats}}
            <tr>
                <td>{{.TeamAbbr}} ({{.TeamName}})</td>
                <td>{{.Count}}</td>
                <td>{{.Positions}}</td>
            </tr>
            {{end}}
        </tbody>
    </table>
</div>
```

**Route:**
```
GET /draft/{id}/stats/franchise
```

**Handler:**
```go
func (h *DraftHandler) GetFranchiseStats(w http.ResponseWriter, r *http.Request) {
    // Query picks JOIN players GROUP BY players.team
    // Count and list positions for each NFL team
    // Render stats_franchise template
}
```

### 2.2 Players Drafted by Position View

**Add to Phase 5:**

Shows all drafted players organized by position (like Excel "Players Drafted" sheet).

**Template: `templates/draft/drafted_by_position.templ`**

```html
<div class="drafted-by-position">
    <div class="position-summary">
        <span>QB: {{.QBCount}}/{{.TotalQB}}</span>
        <span>RB: {{.RBCount}}/{{.TotalRB}}</span>
        <span>WR: {{.WRCount}}/{{.TotalWR}}</span>
        <!-- ... etc -->
    </div>
    
    <div class="position-columns">
        <div class="position-column">
            <h3>Quarterbacks ({{.QBCount}})</h3>
            {{range .QBDrafted}}
                <div>{{.PlayerName}} - {{.TeamName}}</div>
            {{end}}
        </div>
        <!-- Repeat for each position -->
    </div>
</div>
```

### 2.3 Big Board View (Separate from Draft Board)

**Add to Phase 3:**

The Big Board is a **compact grid view** showing team columns and round rows, different from the main draft board.

**Template: `templates/draft/big_board.templ`**

```html
<div class="big-board">
    <table class="big-board-grid">
        <thead>
            <tr>
                <th>Round</th>
                {{range .Teams}}
                <th>
                    {{.TeamName}}<br>
                    <small>{{.OwnerName}}</small>
                </th>
                {{end}}
            </tr>
        </thead>
        <tbody>
            {{range .Rounds}}
            <tr>
                <td>Round {{.Number}}</td>
                {{range .Picks}}
                <td class="pick-cell">
                    {{if .Player}}
                        <strong>{{.Player.Name}}</strong><br>
                        <small>{{.Player.Position}} - {{.Player.Team}}</small>
                    {{else}}
                        <span class="empty">-</span>
                    {{end}}
                </td>
                {{end}}
            </tr>
            {{end}}
        </tbody>
    </table>
</div>
```

**Route:**
```
GET /draft/{id}/big-board
```

### 2.4 Draft Pause/Resume

**Add to Phase 2:**

**UI Elements:**
- "Pause Draft" button (when status = active)
- "Resume Draft" button (when status = paused)
- Paused indicator banner

**Routes:**
```
POST /draft/{id}/pause
POST /draft/{id}/resume
```

**Handler:**
```go
func (h *DraftHandler) PauseDraft(w http.ResponseWriter, r *http.Request) {
    // Validate commissioner permission
    // Update draft.status = 'paused'
    // Log to audit_log
    // Return updated UI
}

func (h *DraftHandler) ResumeDraft(w http.ResponseWriter, r *http.Request) {
    // Validate commissioner permission
    // Update draft.status = 'active'
    // Log to audit_log
    // Return updated UI
}
```

### 2.5 Player Watchlist/Queue

**Add to Phase 4:**

**UI:**
- "Add to Queue" button on each player row
- Queue sidebar showing queued players
- Drag to reorder queue
- Quick-draft from queue

**Routes:**
```
POST   /draft/{id}/queue         # Add player to queue
GET    /draft/{id}/queue         # Get queue
DELETE /draft/{id}/queue/{id}    # Remove from queue
PUT    /draft/{id}/queue/reorder # Reorder queue
```

**Template Component:**
```html
<div id="player-queue">
    <h3>Your Queue</h3>
    {{range .QueuedPlayers}}
    <div class="queue-item" draggable="true">
        <span>{{.Player.Name}} ({{.Player.Position}})</span>
        <button hx-delete="/draft/{{$.DraftID}}/queue/{{.ID}}"
                hx-target="#player-queue">Remove</button>
    </div>
    {{end}}
</div>
```

### 2.6 Custom Player Addition

**Add to Phase 6:**

Allow commissioners to add players not in the database (rookies, late-season additions).

**Form:**
```html
<form hx-post="/players/custom" 
      hx-target="#players-list" 
      hx-swap="afterbegin">
    <h3>Add Custom Player</h3>
    <input name="name" placeholder="Player Name" required>
    <input name="team" placeholder="Team (e.g., KCC)" required>
    <select name="position" required>
        <option value="QB">QB</option>
        <option value="RB">RB</option>
        <option value="WR">WR</option>
        <option value="TE">TE</option>
        <option value="K">K</option>
        <option value="D/ST">D/ST</option>
    </select>
    <input type="number" name="bye_week" placeholder="Bye Week" min="1" max="18">
    <button type="submit">Add Player</button>
</form>
```

**Handler:**
```go
func (h *PlayerHandler) AddCustomPlayer(w http.ResponseWriter, r *http.Request) {
    // Parse form data
    // Set is_custom = true
    // Insert into players table
    // Return player row template
}
```

### 2.7 Trade Pick Support

**Add to Phase 4:**

Allow commissioners to mark picks as traded and reassign to different teams.

**UI:**
- Right-click pick to edit
- "Trade This Pick" option
- Modal to select new team
- "TRADED" badge on pick

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
    "notes": "3rd round pick swap"
}
```

**Handler:**
```go
func (h *DraftHandler) TradePick(w http.ResponseWriter, r *http.Request) {
    // Validate commissioner permission
    // Update pick.team_id
    // Set pick.is_traded = true
    // Log to audit_log
    // Return updated board
}
```

---

## 3. VALIDATION & ERROR HANDLING

### 3.1 Comprehensive Validation Layer

**Add to Phase 1:**

Create a validation package with all 19 rules from spec.

**File: `internal/validation/draft_validation.go`**

```go
package validation

import "errors"

var (
    ErrInvalidLeagueSize    = errors.New("invalid league size. Must be 8, 10, 12, or 14")
    ErrInvalidScoringFormat = errors.New("invalid scoring format")
    ErrInvalidDraftType     = errors.New("invalid draft type")
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
)

func ValidateDraft(draft *Draft) error {
    if draft.Name == "" {
        return ErrDraftNameRequired
    }
    if draft.NumTeams != 8 && draft.NumTeams != 10 && 
       draft.NumTeams != 12 && draft.NumTeams != 14 {
        return ErrInvalidLeagueSize
    }
    // ... more validations
    return nil
}

func ValidateTeam(team *Team, existingTeams []Team) error {
    if team.TeamName == "" {
        return ErrTeamNameRequired
    }
    if len(team.TeamName) > 50 {
        return ErrTeamNameTooLong
    }
    for _, t := range existingTeams {
        if t.TeamName == team.TeamName {
            return ErrDuplicateTeamName
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
    currentTeam, err := CalculateCurrentTeam(pick.OverallPick, draft.NumTeams, teams)
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

### 3.2 Client-Side Validation

**Add to Phase 3:**

Use HTMX attributes and HTML5 validation.

```html
<form hx-post="/draft/{id}/teams"
      hx-validate="true">
    <input name="team_name" 
           required 
           minlength="1" 
           maxlength="50"
           placeholder="Team Name">
    <input name="owner_name" 
           maxlength="50"
           placeholder="Owner Name (optional)">
    <button type="submit">Add Team</button>
</form>
```

### 3.3 Error Display Component

**Template: `templates/components/error.templ`**

```html
<div class="error-message" role="alert">
    <svg class="error-icon"><!-- X icon --></svg>
    <span>{{.ErrorMessage}}</span>
    <button class="dismiss">×</button>
</div>
```

**Handler Pattern:**
```go
func (h *Handler) SomeAction(w http.ResponseWriter, r *http.Request) {
    err := doSomething()
    if err != nil {
        // Return error template instead of success
        component.Error(err.Error()).Render(r.Context(), w)
        w.WriteHeader(http.StatusBadRequest)
        return
    }
    // ... success path
}
```

---

## 4. ADDITIONAL ROUTES

### 4.1 Complete Route List (Updated)

**Add to Phase 2-6:**

```go
// Draft Management
GET    /                           # Home - list all drafts
GET    /draft/new                  # New draft form
POST   /draft/create               # Create draft
GET    /draft/{id}/setup           # Edit draft settings
POST   /draft/{id}/update          # Update settings
POST   /draft/{id}/start           # Start draft
POST   /draft/{id}/pause           # Pause draft
POST   /draft/{id}/resume          # Resume draft
POST   /draft/{id}/complete        # Complete draft
DELETE /draft/{id}                 # Delete draft

// Draft Views
GET    /draft/{id}                 # Main draft board
GET    /draft/{id}/big-board       # Big board grid view
GET    /draft/{id}/board           # Draft board (if separate)

// Teams
GET    /draft/{id}/teams           # List teams
POST   /draft/{id}/teams           # Add team
PUT    /teams/{id}                 # Update team
DELETE /teams/{id}                 # Delete team
GET    /teams/{id}                 # Team detail page
GET    /teams/{id}/roster          # Team roster

// Picks
POST   /draft/{id}/pick            # Make a pick
POST   /draft/{id}/undo            # Undo last pick
POST   /draft/{id}/trade           # Trade a pick
GET    /draft/{id}/current         # Current pick info

// Players
GET    /draft/{id}/players         # Available players (filtered)
GET    /draft/{id}/players/search  # Search players
GET    /players/{id}               # Player details
POST   /players/custom             # Add custom player

// Statistics & Analytics
GET    /draft/{id}/stats                  # Overall stats
GET    /draft/{id}/stats/teams            # Stats by team
GET    /draft/{id}/stats/franchise        # Stats by NFL team
GET    /draft/{id}/stats/position         # Stats by position
GET    /draft/{id}/drafted-players        # Players drafted view
GET    /draft/{id}/bye-weeks              # Bye week analysis
GET    /draft/{id}/value-picks            # Value picks report

// Player Queue (Watchlist)
GET    /draft/{id}/queue                  # Get queue
POST   /draft/{id}/queue                  # Add to queue
DELETE /draft/{id}/queue/{id}             # Remove from queue
PUT    /draft/{id}/queue/reorder          # Reorder queue

// Export
GET    /draft/{id}/export/csv             # CSV export
GET    /draft/{id}/export/pdf             # PDF export
GET    /draft/{id}/export/json            # JSON export

// Real-time Updates
GET    /draft/{id}/stream                 # SSE endpoint
```

---

## 5. PERFORMANCE OPTIMIZATIONS

### 5.1 Available Players Query Optimization

**Add to Phase 1:**

```go
// Optimized query with proper indexing
const queryAvailablePlayers = `
    SELECT p.* 
    FROM players p
    WHERE p.id NOT IN (
        SELECT player_id 
        FROM picks 
        WHERE draft_id = ?
    )
    AND (? = '' OR p.position IN (?))
    AND (? = '' OR p.name LIKE ? OR p.team LIKE ?)
    ORDER BY 
        CASE 
            WHEN ? = 'Redraft' THEN 
                CASE ? 
                    WHEN 'PPR' THEN p.ppr_rank
                    WHEN 'Half-PPR' THEN p.half_ppr_rank
                    ELSE p.std_rank
                END
            ELSE p.dynasty_rank
        END
    LIMIT 100
`

// Add to database setup
CREATE INDEX idx_picks_player_draft ON picks(player_id, draft_id);
CREATE INDEX idx_players_name ON players(name);
CREATE INDEX idx_players_rank ON players(ppr_rank, half_ppr_rank, std_rank);
```

### 5.2 Draft Board Pagination

**Add to Phase 6:**

For large drafts (420+ picks), implement pagination.

```html
<div class="board-pagination">
    <button hx-get="/draft/{{.DraftID}}/board?rounds=1-10"
            hx-target="#board-content">
        Rounds 1-10
    </button>
    <button hx-get="/draft/{{.DraftID}}/board?rounds=11-20"
            hx-target="#board-content">
        Rounds 11-20
    </button>
    <button hx-get="/draft/{{.DraftID}}/board?rounds=21-30"
            hx-target="#board-content">
        Rounds 21-30
    </button>
</div>

<div id="board-content">
    <!-- Board rows for selected rounds -->
</div>
```

### 5.3 Caching Strategy

**Add to Phase 6:**

```go
// Use an in-memory cache for frequently accessed data
type DraftCache struct {
    availablePlayers map[int][]*Player // draft_id -> players
    currentPick      map[int]*PickInfo  // draft_id -> pick info
    mutex            sync.RWMutex
}

func (c *DraftCache) InvalidateAvailablePlayers(draftID int) {
    c.mutex.Lock()
    defer c.mutex.Unlock()
    delete(c.availablePlayers, draftID)
}

// Invalidate cache after every pick
func (r *PickRepository) Create(pick *Pick) error {
    err := r.db.Create(pick)
    if err == nil {
        r.cache.InvalidateAvailablePlayers(pick.DraftID)
    }
    return err
}
```

---

## 6. UI/UX ENHANCEMENTS

### 6.1 Loading States

**Add to Phase 6:**

Use HTMX indicators for all async actions.

```html
<!-- Global loading indicator -->
<div id="loading-indicator" class="htmx-indicator">
    <svg class="spinner">...</svg>
    <span>Loading...</span>
</div>

<!-- Per-action loading -->
<button hx-post="/draft/{id}/pick"
        hx-indicator="#loading-indicator">
    Draft Player
</button>
```

**CSS:**
```css
.htmx-indicator {
    display: none;
}
.htmx-request .htmx-indicator {
    display: flex;
}
.htmx-request.htmx-indicator {
    display: flex;
}
```

### 6.2 Success Confirmations

**Add toast notification system:**

```html
<div id="toast-container"></div>

<!-- Success toast template -->
<div class="toast toast-success" hx-swap-oob="afterbegin:#toast-container">
    <svg class="check-icon">...</svg>
    <span>{{.Message}}</span>
</div>
```

**Handler:**
```go
func (h *DraftHandler) MakePick(w http.ResponseWriter, r *http.Request) {
    // ... make pick ...
    
    // Return board update + success toast
    templ.Handler(component.DraftBoard(board), 
                  component.Toast("Player drafted successfully!")).
        ServeHTTP(w, r)
}
```

### 6.3 Print-Friendly Views

**Add to Phase 5:**

```html
<style media="print">
    .no-print { display: none; }
    .big-board { 
        font-size: 10pt;
        page-break-after: always;
    }
    @page {
        size: landscape;
        margin: 0.5in;
    }
</style>

<button class="no-print" onclick="window.print()">
    Print Big Board
</button>
```

### 6.4 Draft URL Sharing

**Add to Phase 2:**

```html
<div class="draft-share">
    <label>Share this draft:</label>
    <input type="text" 
           readonly 
           value="https://yourdomain.com/draft/{{.DraftID}}"
           id="draft-url">
    <button onclick="copyDraftURL()">
        Copy Link
    </button>
</div>

<script>
function copyDraftURL() {
    const input = document.getElementById('draft-url');
    input.select();
    document.execCommand('copy');
    // Show toast
}
</script>
```

---

## 7. TESTING REQUIREMENTS

### 7.1 Unit Tests (Add to All Phases)

**File: `internal/draft/snake_test.go`**

```go
func TestCalculateCurrentTeam(t *testing.T) {
    teams := []Team{
        {ID: 1, DraftPosition: 1},
        {ID: 2, DraftPosition: 2},
        {ID: 3, DraftPosition: 3},
        {ID: 4, DraftPosition: 4},
    }
    
    tests := []struct{
        pick     int
        expected int
    }{
        {1, 1},   // Round 1, Pick 1 -> Team 1
        {4, 4},   // Round 1, Pick 4 -> Team 4
        {5, 4},   // Round 2, Pick 1 -> Team 4 (snake)
        {8, 1},   // Round 2, Pick 4 -> Team 1 (snake)
        {9, 1},   // Round 3, Pick 1 -> Team 1
    }
    
    for _, tt := range tests {
        team, err := CalculateCurrentTeam(tt.pick, 4, teams)
        if err != nil {
            t.Errorf("Pick %d: unexpected error: %v", tt.pick, err)
        }
        if team.DraftPosition != tt.expected {
            t.Errorf("Pick %d: expected position %d, got %d", 
                     tt.pick, tt.expected, team.DraftPosition)
        }
    }
}
```

### 7.2 Integration Tests

**File: `internal/handlers/draft_test.go`**

```go
func TestCompleteDraftFlow(t *testing.T) {
    // 1. Create draft
    // 2. Add teams
    // 3. Start draft
    // 4. Make all picks in correct order
    // 5. Verify draft completion
    // 6. Verify no more picks can be made
}
```

---

## 8. DOCUMENTATION UPDATES

### 8.1 Update Implementation Plan

**Add to each phase:**

✅ **Phase 1 Additions:**
- Draft status enum
- Audit log table
- Player queue table
- Enhanced validation package
- Snake draft formula implementation

✅ **Phase 2 Additions:**
- Pause/resume functionality
- Max rounds configuration
- Draft URL sharing
- Commissioner permissions

✅ **Phase 3 Additions:**
- Big Board separate view
- Stats by Franchise view
- Players Drafted by Position view
- Bye week conflict warnings
- Draft completion detection

✅ **Phase 4 Additions:**
- Player watchlist/queue
- Trade pick recording
- Advanced SSE with multiple event types

✅ **Phase 5 Additions:**
- ADP vs Actual tracking
- Value picks report
- Draft grades
- Print-friendly exports

✅ **Phase 6 Additions:**
- Custom player addition
- All validation error displays
- Performance optimizations
- Complete testing suite

---

## 9. PRIORITY RECOMMENDATIONS

### Must Have (MVP)
1. ✅ Draft status field and pause/resume
2. ✅ Snake draft formula (exact implementation)
3. ✅ Draft completion detection
4. ✅ Big Board view
5. ✅ All validation rules
6. ✅ Basic error handling

### Should Have (Post-MVP)
1. 🔲 Player watchlist/queue
2. 🔲 Stats by Franchise view
3. 🔲 Players Drafted by Position view
4. 🔲 Trade pick support
5. 🔲 Custom player addition
6. 🔲 Audit log

### Nice to Have (Future)
1. 🔲 Draft grades and value picks
2. 🔲 Print-friendly views
3. 🔲 Advanced caching
4. 🔲 Bye week analysis
5. 🔲 Position run tracking

---

## 10. UPDATED TIMELINE

### Week 1: Foundation (Extended)
- ✅ Original Phase 1 tasks
- ✅ Enhanced database schema with all new tables
- ✅ Validation package with all 19 rules
- ✅ Snake draft formula
- ✅ Draft completion logic

### Week 2: Setup & Configuration
- ✅ Original Phase 2 tasks
- ✅ Pause/resume functionality
- ✅ Draft URL sharing
- ✅ Max rounds configuration

### Week 3: Core Board
- ✅ Original Phase 3 tasks
- ✅ Big Board view (separate)
- ✅ Current pick calculation
- ✅ All validation integrated

### Week 4: Advanced Features
- ✅ Original Phase 4 tasks
- ✅ Player queue/watchlist
- ✅ Trade recording
- ✅ Enhanced SSE

### Week 5: Analytics & Stats
- ✅ Original Phase 5 tasks
- ✅ Stats by Franchise
- ✅ Players Drafted by Position
- ✅ Value picks analysis

### Week 6: Polish & Launch
- ✅ Original Phase 6 tasks
- ✅ Custom player addition
- ✅ All error displays
- ✅ Performance optimization
- ✅ Complete testing

**Total: 6-7 weeks** (1 week buffer added)

---

## Summary

By addressing these 61 gaps, the implementation plan will fully match the specification document and deliver all features present in the original Excel workbook plus modern web capabilities. The updated plan is now comprehensive and production-ready.

**Next Steps:**
1. Review and approve gap analysis
2. Update main implementation plan document
3. Begin Phase 1 with enhanced schema
4. Implement in iterative sprints

