# Fantasy Football Offline Draft Tracker - Specification Document

**Version:** 1.2.2  
**Season:** 2025  
**ADP Date:** August 21, 2025  
**Data Source:** FantasyPros Consensus ADP

---

## Table of Contents

1. [Overview](#overview)
2. [Core Requirements](#core-requirements)
3. [League Configuration](#league-configuration)
4. [Player Database](#player-database)
5. [Draft Functionality](#draft-functionality)
6. [Views and Displays](#views-and-displays)
7. [Statistics and Tracking](#statistics-and-tracking)
8. [Business Rules](#business-rules)
9. [User Workflows](#user-workflows)
10. [Data Validation Rules](#data-validation-rules)
11. [Edge Cases and Special Scenarios](#edge-cases-and-special-scenarios)

---

## 1. Overview

### Purpose
A web-based tool to facilitate offline fantasy football drafts, allowing commissioners to track picks in real-time and share draft progress with all league members.

### Use Cases
- Offline drafts with friends (in-person gatherings)
- Alternative to platform draft tools (ESPN, Yahoo, etc.)
- Mock drafts for practice
- Dynasty and keeper league drafts

### Key Features
- Real-time draft tracking
- Snake draft pattern automation
- Position-based player filtering
- Team roster building
- Statistics by team and position
- Available players list
- Draft board visualization

---

## 2. Core Requirements

### Must Have Features
1. ✅ Support for 8, 10, 12, and 14 team leagues
2. ✅ Snake draft order calculation (automatic reversal on even rounds)
3. ✅ Player selection from comprehensive database (1400+ players)
4. ✅ Real-time draft board updates
5. ✅ Team-by-team roster tracking
6. ✅ Position filtering and search
7. ✅ Available players list (excluding already drafted)
8. ✅ Draft statistics by team
9. ✅ Draft statistics by position
10. ✅ Support for multiple scoring formats
11. ✅ Support for Dynasty and Redraft rankings

### Should Have Features
1. 🔲 Undo last pick functionality
2. 🔲 Export draft results
3. 🔲 Draft history/audit trail
4. 🔲 Player watchlist/queue
5. 🔲 Best available player suggestions

### Nice to Have Features
1. 🔲 Draft timer per pick
2. 🔲 Trade pick support (manual override)
3. 🔲 Keeper league support
4. 🔲 Custom rankings import
5. 🔲 Mobile responsive design

---

## 3. League Configuration

### 3.1 Team Size Options

**Supported League Sizes:**
- 8 teams
- 10 teams
- 12 teams
- 14 teams

**Default:** 10 teams

**Constraint:** Must select one and only one team size

### 3.2 Scoring Formats

**Options:**
1. **Standard (Non-PPR)**
   - No points for receptions
   - Traditional scoring

2. **Half-PPR**
   - 0.5 points per reception
   - Middle ground approach

3. **PPR (Points Per Reception)**
   - 1 point per reception
   - Favors pass-catching backs and slot receivers

**Default:** PPR

**Constraint:** Must select exactly one scoring format

### 3.3 Draft Type

**Options:**
1. **Redraft**
   - One-season league
   - Uses season-specific ADP rankings
   - Standard ADP by scoring format

2. **Dynasty**
   - Multi-year keeper league
   - Uses dynasty-specific rankings
   - Long-term value prioritized

**Default:** Redraft

**Additional Dynasty Setting:**
- **QB Settings:** 1QB or 2QB/Superflex
  - Only applies to Dynasty format
  - Affects quarterback rankings/value

### 3.4 Position Settings

**Available Positions:**

**Offensive Positions (Always Available):**
- QB (Quarterback)
- RB (Running Back)
- WR (Wide Receiver)
- TE (Tight End)
- K (Kicker)
- D/ST (Defense/Special Teams)

**IDP Positions (Optional):**
- DL (Defensive Line)
- LB (Linebacker)
- DB (Defensive Back)

**Default Configuration:**
- QB: Enabled ✓
- RB: Enabled ✓
- WR: Enabled ✓
- TE: Enabled ✓
- K: Disabled ✗
- D/ST: Enabled ✓
- DL: Disabled ✗
- LB: Disabled ✗
- DB: Disabled ✗

**Rules:**
- Position filters affect which players appear in available players list
- Disabled positions can still be drafted (no hard restriction)
- IDP positions have no ranking data (placeholders only)
- Users can manually enable/disable any position

### 3.5 Team Configuration

**Team Information Required:**
1. **Team Name** (required)
   - Display name for the team
   - Example: "Dan's Demons", "Go Long Baby"

2. **Owner Name** (optional)
   - Person managing the team
   - Example: "Dan B", "Steve"

3. **Draft Position** (required)
   - Numeric order (1, 2, 3, ... N teams)
   - Determines snake draft order

**Draft Order Rules:**
- Must have exactly N teams (where N = league size)
- Each team must have unique draft position (1 through N)
- Draft positions cannot be duplicated
- Draft order can be set automatically or manually

**Example 10-Team Draft Order:**
```
Position 1: Go Long Baby (Steve)
Position 2: Fueled By Bourbon (Robert L)
Position 3: Andy Reid's Mustache (Daniel R)
Position 4: Charlie's Angels (Audra)
Position 5: Team 'Cole' Man (Gerry)
Position 6: Jeff's Airmen (Jeff)
Position 7: Bub's Bros (Alan)
Position 8: Memes & Caffeine (Michael)
Position 9: Dan's Demons (Dan B)
Position 10: Robert's Rams (Robert S)
```

---

## 4. Player Database

### 4.1 Player Data Structure

**Total Players:** 1,472

**Player Attributes:**
1. **Player Name** (string, required)
   - Full name including suffixes (Jr., Sr., II, III)
   - Example: "Ja'Marr Chase", "Patrick Mahomes II"

2. **NFL Team** (string, 2-4 char abbreviation)
   - Team abbreviation code
   - Example: "CIN", "KCC", "SFO", "NOS"

3. **Position** (string, enum)
   - One of: QB, RB, WR, TE, K, D/ST, DL, LB, DB
   - Single position per player

4. **Bye Week** (integer, 1-18)
   - Week when team/player is off
   - Used for roster planning

5. **Rankings by Format** (integer, rank order)
   - Dynasty Rank
   - Superflex (SF) Rank
   - Standard Rank
   - Half-PPR Rank
   - PPR Rank

**Ranking Notes:**
- Lower number = higher rank (1 is best)
- Different rankings per scoring format
- Some players may have null/no rank in certain formats
- IDP positions (DL, LB, DB) have no ranking data

### 4.2 Player Count by Position

| Position | Count | Description |
|----------|-------|-------------|
| QB | 81 | Quarterbacks |
| RB | 147 | Running Backs |
| WR | 235 | Wide Receivers |
| TE | 119 | Tight Ends |
| K | 60 | Kickers |
| D/ST | 32 | Defense/Special Teams |
| DL | 353 | Defensive Linemen (IDP) |
| LB | 127 | Linebackers (IDP) |
| DB | 318 | Defensive Backs (IDP) |
| **Total** | **1,472** | |

### 4.3 NFL Team Abbreviations

**Standard 3-Letter Codes:**
- ARI (Arizona Cardinals)
- ATL (Atlanta Falcons)
- BAL (Baltimore Ravens)
- BUF (Buffalo Bills)
- CAR (Carolina Panthers)
- CHI (Chicago Bears)
- CIN (Cincinnati Bengals)
- CLE (Cleveland Browns)
- DAL (Dallas Cowboys)
- DEN (Denver Broncos)
- DET (Detroit Lions)
- GBP (Green Bay Packers)
- HOU (Houston Texans)
- IND (Indianapolis Colts)
- JAC (Jacksonville Jaguars)
- KCC (Kansas City Chiefs)
- LVR (Las Vegas Raiders)
- LAC (Los Angeles Chargers)
- LAR (Los Angeles Rams)
- MIA (Miami Dolphins)
- MIN (Minnesota Vikings)
- NEP (New England Patriots)
- NOS (New Orleans Saints)
- NYG (New York Giants)
- NYJ (New York Jets)
- PHI (Philadelphia Eagles)
- PIT (Pittsburgh Steelers)
- SFO (San Francisco 49ers)
- SEA (Seattle Seahawks)
- TBB (Tampa Bay Buccaneers)
- TEN (Tennessee Titans)
- WAS (Washington Commanders)

### 4.4 Sample Top Players (by PPR Rank)

**Top 20 Overall (PPR):**
1. Ja'Marr Chase (CIN, WR) - Bye 10
2. Bijan Robinson (ATL, RB) - Bye 5
3. Saquon Barkley (PHI, RB) - Bye 9
4. Justin Jefferson (MIN, WR) - Bye 6
5. Jahmyr Gibbs (DET, RB) - Bye 8
6. CeeDee Lamb (DAL, WR) - Bye 10
7. Christian McCaffrey (SFO, RB) - Bye 14
8. Amon-Ra St. Brown (DET, WR) - Bye 8
9. Puka Nacua (LAR, WR) - Bye 8
10. Malik Nabers (NYG, WR) - Bye 14
11. Ashton Jeanty (LVR, RB) - Bye 8
12. Brian Thomas Jr. (JAC, WR) - Bye 8
13. De'Von Achane (MIA, RB) - Bye 12
14. Nico Collins (HOU, WR) - Bye 6
15. Brian Robinson Jr. (WAS, RB) - Bye 11
16. Brock Bowers (LVR, TE) - Bye 8
17. Drake London (ATL, WR) - Bye 5
18. Jonathan Taylor (IND, RB) - Bye 11
19. A.J. Brown (PHI, WR) - Bye 9
20. Josh Allen (BUF, QB) - Bye 7

---

## 5. Draft Functionality

### 5.1 Snake Draft Pattern

**How Snake Draft Works:**

**Round 1 (Forward Order):**
```
Pick 1:  Team 1
Pick 2:  Team 2
Pick 3:  Team 3
...
Pick N:  Team N
```

**Round 2 (Reverse Order):**
```
Pick N+1: Team N
Pick N+2: Team N-1
Pick N+3: Team N-2
...
Pick 2N:  Team 1
```

**Round 3 (Forward Order Again):**
```
Pick 2N+1: Team 1
Pick 2N+2: Team 2
...
```

**Pattern:**
- Odd rounds (1, 3, 5, ...): Draft in ascending order (1 → N)
- Even rounds (2, 4, 6, ...): Draft in descending order (N → 1)

**Example 10-Team Snake Draft (First 3 Rounds):**

| Round | Pick | Team Position | Team Name |
|-------|------|---------------|-----------|
| 1 | 1 | 1 | Team A |
| 1 | 2 | 2 | Team B |
| 1 | 3 | 3 | Team C |
| ... | ... | ... | ... |
| 1 | 10 | 10 | Team J |
| 2 | 11 | 10 | Team J ← |
| 2 | 12 | 9 | Team I ← |
| 2 | 13 | 8 | Team H ← |
| ... | ... | ... | ... |
| 2 | 20 | 1 | Team A ← |
| 3 | 21 | 1 | Team A → |
| 3 | 22 | 2 | Team B → |
| ... | ... | ... | ... |

### 5.2 Pick Mechanics

**Required Information for Each Pick:**
1. Draft ID (which draft this pick belongs to)
2. Player ID (which player was selected)
3. Team ID (which team made the selection)
4. Round Number (calculated automatically)
5. Overall Pick Number (sequential: 1, 2, 3, ...)
6. Timestamp (when the pick was made)

**Validation Rules:**
- ✅ Player can only be drafted once per draft
- ✅ Pick must be made by the team currently on the clock
- ✅ Pick number must be sequential (no skipping)
- ✅ Player must exist in database
- ✅ Team must exist in draft

**Draft State:**
- Current pick number = (total picks already made) + 1
- Current round = ceiling(current_pick / num_teams)
- On the clock team = calculate based on snake pattern

**Snake Draft Team Calculation:**
```
round_number = ceiling(pick_number / num_teams)
position_in_round = ((pick_number - 1) % num_teams) + 1

if round_number is odd:
    team_position = position_in_round
else:
    team_position = num_teams - position_in_round + 1
```

### 5.3 Draft Completion

**Draft is Complete When:**
- All roster spots are filled for all teams
- Commissioner manually ends draft
- Predetermined number of rounds reached

**Typical Draft Lengths:**
- Standard Leagues: 15-16 rounds
- Deep Leagues: 20+ rounds
- Shallow Leagues: 10-12 rounds

**No Hard Limit:**
- System should support unlimited rounds
- Excel version has 26 rounds (Row 27 max in Big Board)
- Webapp should support at least 30 rounds

### 5.4 Undo Functionality

**Requirements:**
- Must be able to undo the last pick made
- Undo removes pick from database
- Player becomes available again
- Pick counter decrements
- Current team on clock reverts to previous team
- Should support multiple undo operations (undo chain)

**Constraints:**
- Cannot undo picks in the middle of draft (only last pick)
- Undo should maintain audit trail (optional)
- Undo should be prominently accessible in UI

---

## 6. Views and Displays

### 6.1 Draft Board View

**Layout:**
- Grid format showing all picks
- One column per team
- One row per round
- Visual indicator of snake pattern

**Column Headers:**
- Team name
- Owner name (optional)
- Current position count (e.g., "2 QB | 5 RB | 4 WR...")

**Cell Content:**
- Player name
- Position badge/indicator
- NFL team
- Pick number (optional)

**Visual Indicators:**
- Current pick highlighted
- Completed picks visible
- Future picks empty/placeholder
- Snake pattern direction arrows

**Sample Layout (10 Teams, First 3 Rounds):**

```
         Team 1    Team 2    Team 3    ...    Team 10
Round 1  Player A  Player B  Player C  ...    Player J
Round 2  Player T  Player S  Player R  ...    Player K (reverse)
Round 3  Player U  Player V  Player W  ...    Player DD
```

### 6.2 Available Players View

**Filter Options:**
1. **Position Checkboxes**
   - QB, RB, WR, TE, K, D/ST, DL, LB, DB
   - Multiple selections allowed
   - Default: All enabled positions checked

2. **Search Box**
   - Filter by player name
   - Filter by team abbreviation
   - Real-time/debounced search

3. **Sort Options**
   - By Rank (ascending - best first)
   - By Name (alphabetical)
   - By Position
   - By Team
   - By Bye Week

**Player List Display:**

**Columns:**
1. Rank (based on selected scoring format)
2. Player Name
3. NFL Team
4. Position
5. Bye Week
6. Action (Draft button/click handler)

**Sample Row:**
```
1  |  Ja'Marr Chase  |  CIN  |  WR  |  10  |  [DRAFT]
```

**Behavior:**
- Only show undrafted players
- Apply position filters
- Apply search filters
- Highlight "Best Available" (top ranked)
- Support quick-draft (click player to draft)

**Additional Features:**
- Player count indicator (e.g., "245 players available")
- Position count indicator (e.g., "QB: 12, RB: 23, WR: 45...")
- Refresh/update button

### 6.3 Team Roster View

**Display Per Team:**

**Summary Header:**
- Team name
- Owner name
- Total picks made
- Current roster composition

**Position Breakdown:**
```
QB: 2
RB: 5
WR: 4
TE: 2
K: 0
D/ST: 1
Total: 14 players
```

**Detailed Roster:**
- List of all drafted players
- Show: Player name, position, NFL team, bye week, round picked
- Group by position or order by pick number
- Highlight positional needs (optional)

**Sample Display:**
```
Team: Dan's Demons (Dan B)
Picks: 14

Quarterbacks (2):
- Josh Allen (BUF) - Round 4, Pick 37
- Baker Mayfield (TBB) - Round 12, Pick 117

Running Backs (5):
- Bijan Robinson (ATL) - Round 1, Pick 9
- Jahmyr Gibbs (DET) - Round 3, Pick 29
...
```

### 6.4 Big Board View

**Alternative Grid Layout:**
- Shows N columns (one per team)
- Shows M rows (one per round)
- Each cell shows player picked
- Compact format for printing/display

**Information Per Cell:**
- Player name
- Position
- NFL team

**Example Format:**
```
Ja'Marr Chase
WR - CIN
```

**Special Features:**
- Color coding by position
- Bye week indicators
- Click for player details
- Export to PDF/image

### 6.5 Statistics Views

**Stats by Team:**

**Table Format:**
| Team | Overall | QB | RB | WR | TE | K | D/ST |
|------|---------|----|----|----|----|---|------|
| Team A | 16 | 2 | 5 | 6 | 2 | 0 | 1 |
| Team B | 16 | 2 | 4 | 7 | 2 | 0 | 1 |
| ... | ... | ... | ... | ... | ... | ... | ... |

**Stats by Position:**

**Players Drafted Counts:**
```
QB: 20 drafted (61 remaining)
RB: 50 drafted (97 remaining)
WR: 60 drafted (175 remaining)
TE: 20 drafted (99 remaining)
K: 0 drafted (60 remaining)
D/ST: 10 drafted (22 remaining)
```

**List of Drafted Players by Position:**
- Group all drafted players by position
- Show in draft order
- Click to see which team drafted

---

## 7. Statistics and Tracking

### 7.1 Team-Level Statistics

**For Each Team Track:**

**Pick Count by Position:**
- Total QB drafted
- Total RB drafted
- Total WR drafted
- Total TE drafted
- Total K drafted
- Total D/ST drafted
- Total IDP drafted (DL, LB, DB)

**Overall Picks:**
- Total players drafted
- Rounds completed
- Picks remaining (if draft not complete)

**Bye Week Distribution:**
- Count of players by bye week
- Identify bye week conflicts
- Bye week coverage analysis

**Positional Needs:**
- Identify weak positions
- Compare to league average
- Suggest next pick positions

**Draft Grade (Optional):**
- Based on ADP vs. actual pick
- Value picks (drafted later than ADP)
- Reaches (drafted earlier than ADP)

### 7.2 League-Level Statistics

**Overall Draft Stats:**
- Total picks made
- Current round
- Picks remaining
- Draft duration (time elapsed)
- Average time per pick

**Position Run Tracking:**
- Identify position runs (e.g., "5 RBs picked in last 10 picks")
- Position scarcity indicators
- Next position likely to be drafted

**Team Franchise Statistics:**

Track picks by NFL team:
```
Kansas City Chiefs: 8 players drafted
Dallas Cowboys: 12 players drafted
Detroit Lions: 15 players drafted
```

**Value Picks Report:**
- Players drafted well below ADP
- Steals of the draft
- Biggest reaches

### 7.3 Player Statistics

**For Each Player (if Drafted):**
- Drafted by which team
- Round picked
- Overall pick number
- ADP vs. Actual pick difference
- Time of pick (timestamp)

**For Available Players:**
- ADP rank
- Position rank
- Time since last drafted (if tracking draft speed)

---

## 8. Business Rules

### 8.1 Draft Rules

**BR-001: One Player Per Pick**
- Each pick must select exactly one player
- Cannot draft multiple players simultaneously

**BR-002: No Duplicate Picks**
- A player can only be drafted once per draft
- System must prevent selecting already-drafted players

**BR-003: Sequential Pick Order**
- Picks must be made in sequential order
- Cannot skip picks
- Cannot go backwards (except via undo)

**BR-004: Snake Draft Pattern**
- Even rounds reverse team order
- Odd rounds follow standard order
- Pattern calculated automatically

**BR-005: Position Filter Display Only**
- Position settings affect UI filters only
- Do not prevent drafting filtered positions
- Players from disabled positions can still be drafted

### 8.2 Team Rules

**BR-006: Unique Team Names**
- Each team must have unique name within draft
- Team names are required

**BR-007: Unique Draft Positions**
- Draft positions 1 through N must be assigned
- Each position used exactly once
- No gaps or duplicates

**BR-008: Team Count Must Match League Size**
- If league is 10 teams, must have exactly 10 teams
- Cannot start draft without all teams configured

### 8.3 Validation Rules

**BR-009: Player Must Exist**
- Selected player must be in database
- Player ID must be valid

**BR-010: Team Must Exist**
- Picking team must be valid team in draft
- Team ID must match draft

**BR-011: Correct Team's Turn**
- Only the team currently on the clock can make a pick
- System prevents out-of-turn picks

**BR-012: Draft Must Be Active**
- Cannot make picks in completed drafts
- Cannot make picks in deleted drafts

### 8.4 Data Integrity Rules

**BR-013: Referential Integrity**
- Picks reference valid players
- Picks reference valid teams
- Teams reference valid drafts

**BR-014: Timestamp Accuracy**
- All picks must have timestamp
- Timestamps must be sequential
- Timestamps cannot be in future

**BR-015: Round Calculation Accuracy**
- Round number must be accurate
- Round = ceiling(pick_number / num_teams)
- Must match snake pattern

### 8.5 UI/UX Rules

**BR-016: Visual Current Pick Indicator**
- Always show which team is on the clock
- Always show current round and pick number
- Update in real-time

**BR-017: Available Players Only**
- Only show undrafted players in available list
- Remove player immediately after drafting
- Update counts in real-time

**BR-018: Filter Persistence**
- Remember user's position filters
- Remember search terms
- Remember sort preferences

---

## 9. User Workflows

### 9.1 Commissioner Setup Workflow

**Step 1: Create New Draft**
- Click "New Draft" button
- Enter draft name
- System generates unique draft ID

**Step 2: Configure Draft Settings**
- Select number of teams (8/10/12/14)
- Select scoring format (Standard/Half-PPR/PPR)
- Select draft type (Redraft/Dynasty)
- If Dynasty, select QB setting (1QB/2QB-SF)

**Step 3: Configure Position Filters**
- Enable/disable positions as needed
- Default: QB, RB, WR, TE, D/ST enabled
- Save position settings

**Step 4: Add Teams**
- For each team:
  - Enter team name (required)
  - Enter owner name (optional)
  - Assign draft position (1-N)
- Ensure all N positions filled
- Save team configuration

**Step 5: Verify and Start Draft**
- Review all settings
- Verify team count matches league size
- Verify all draft positions assigned
- Click "Start Draft"

### 9.2 Making a Pick Workflow

**Step 1: Identify Current Pick**
- System highlights team on the clock
- Shows: "Round X, Pick Y"
- Shows: "Team Name is on the clock"

**Step 2: Search/Filter Players**
- Use position filters to narrow options
- Use search box to find specific player
- Sort by rank, name, or other criteria
- Review available players

**Step 3: Select Player**
- Click on desired player row
- Or click "Draft" button next to player
- System prompts for confirmation (optional)

**Step 4: Confirm Pick**
- System validates pick
- System records pick in database
- System updates draft board
- System updates available players list
- System updates team roster
- System advances to next pick

**Step 5: Continue Draft**
- Next team is now on the clock
- Repeat steps 2-4 until draft complete

### 9.3 Viewing Draft Progress Workflow

**As Participant (Non-Commissioner):**

**Step 1: Access Draft**
- Navigate to draft URL
- Draft loads with current state

**Step 2: View Draft Board**
- See all completed picks
- See which team is on the clock
- See team rosters

**Step 3: Monitor Available Players**
- View available players list
- Apply filters for positions of interest
- Track specific players

**Step 4: View Team Statistics**
- Click on team to see detailed roster
- View position breakdown
- See bye week distribution

**Step 5: Auto-Refresh**
- Board updates automatically when picks made
- No manual refresh required

### 9.4 Undoing a Pick Workflow

**Step 1: Identify Mistake**
- Commissioner notices error
- Wrong player drafted
- Out of order pick

**Step 2: Click Undo**
- Click "Undo Last Pick" button
- System prompts for confirmation

**Step 3: Confirm Undo**
- Click "Confirm"
- System removes last pick from database
- Player becomes available again
- Pick counter decrements
- Previous team back on the clock

**Step 4: Make Correct Pick**
- Proceed with correct pick
- Continue draft normally

### 9.5 Exporting Draft Results Workflow

**Step 1: Complete or Pause Draft**
- Draft can be exported at any time
- All current picks will be included

**Step 2: Choose Export Format**
- CSV (spreadsheet compatible)
- PDF (printable document)
- JSON (for API/integration)

**Step 3: Configure Export Options**
- Include team rosters
- Include statistics
- Include timestamps
- Select layout (grid vs. list)

**Step 4: Download Export**
- Click "Export"
- File downloads to user's device
- Open in appropriate application

---

## 10. Data Validation Rules

### 10.1 Draft Configuration Validation

**DV-001: League Size**
- **Rule:** Must be 8, 10, 12, or 14
- **Error:** "Invalid league size. Must be 8, 10, 12, or 14."

**DV-002: Scoring Format**
- **Rule:** Must be Standard, Half-PPR, or PPR
- **Error:** "Invalid scoring format."

**DV-003: Draft Type**
- **Rule:** Must be Redraft or Dynasty
- **Error:** "Invalid draft type."

**DV-004: Draft Name**
- **Rule:** Required, 1-100 characters
- **Error:** "Draft name is required."

### 10.2 Team Configuration Validation

**DV-005: Team Name Required**
- **Rule:** Team name cannot be empty
- **Error:** "Team name is required."

**DV-006: Team Name Length**
- **Rule:** 1-50 characters
- **Error:** "Team name must be between 1 and 50 characters."

**DV-007: Unique Team Name**
- **Rule:** No duplicate team names in same draft
- **Error:** "Team name already exists in this draft."

**DV-008: Draft Position Range**
- **Rule:** Must be between 1 and N (league size)
- **Error:** "Draft position must be between 1 and [N]."

**DV-009: Unique Draft Position**
- **Rule:** Each position used exactly once
- **Error:** "Draft position [X] is already assigned."

**DV-010: Complete Team Roster**
- **Rule:** Must have exactly N teams for league size N
- **Error:** "Must have exactly [N] teams. Currently have [X]."

### 10.3 Pick Validation

**DV-011: Player Exists**
- **Rule:** Player ID must be valid
- **Error:** "Invalid player ID."

**DV-012: Player Available**
- **Rule:** Player not already drafted in this draft
- **Error:** "Player [Name] has already been drafted."

**DV-013: Team Exists**
- **Rule:** Team ID must be valid for this draft
- **Error:** "Invalid team ID."

**DV-014: Correct Team Turn**
- **Rule:** Only team on the clock can draft
- **Error:** "Not [Team Name]'s turn to pick."

**DV-015: Draft Active**
- **Rule:** Draft must not be completed or deleted
- **Error:** "Cannot make picks in completed draft."

**DV-016: Sequential Pick**
- **Rule:** Pick number must be next sequential number
- **Error:** "Pick number must be [N]."

### 10.4 Search and Filter Validation

**DV-017: Search Length**
- **Rule:** Search query 0-50 characters
- **Error:** "Search query too long."

**DV-018: Valid Position Filter**
- **Rule:** Position must be valid (QB, RB, WR, TE, K, D/ST, DL, LB, DB)
- **Error:** "Invalid position filter."

**DV-019: Valid Sort Option**
- **Rule:** Sort must be valid option (rank, name, position, team, bye)
- **Error:** "Invalid sort option."

---

## 11. Edge Cases and Special Scenarios

### 11.1 Draft Trade Support

**Scenario:** Team A wants to trade picks with Team B

**Current Excel Behavior:**
- Manually edit "Team" column in Draft sheet
- Change team name for affected picks
- Does not update Big Board (visual only)
- Everything else updates automatically

**Webapp Requirements:**
- Support manual override of draft order
- Allow commissioner to assign pick to different team
- Mark pick as "traded" for audit trail
- Update all views to reflect trade

**Implementation Notes:**
- Keep picks in sequential order by number
- Change team_id associated with pick
- Add "is_traded" flag to picks table
- Show trade indicator in UI

### 11.2 Keeper League Support

**Scenario:** League allows keepers from previous season

**Two Approaches:**

**Approach 1: Pre-Draft Keepers**
- Enter keepers as picks in corresponding rounds
- If Team A keeps Player X in Round 5, enter as Round 5 pick
- Draft begins at next available pick

**Approach 2: Delete Kept Players**
- Remove keeper players from database before draft
- Only available players appear in list
- Warning: Can break rankings if not careful

**Webapp Recommendation:**
- Approach 1 is safer
- Add "Keeper" flag to picks
- Allow commissioner to pre-populate keeper picks
- Show "KEEPER" indicator on draft board
- Start live draft after keepers set

### 11.3 Custom Rankings

**Scenario:** User has their own rankings, not FantasyPros

**Current Excel Behavior:**
- Edit HELP_PlayerDB sheet (hidden by default)
- Manually change rank values
- Warning: Can break formulas if done incorrectly

**Webapp Requirements:**
- Import CSV with custom rankings
- Map CSV columns to player IDs
- Update rank values in database
- Preserve original rankings (optional rollback)

**Implementation Notes:**
- Add "custom_rank" column separate from default ranks
- Toggle between "Official ADP" and "Custom Rankings"
- CSV format: player_name, team, position, rank
- Match players by name+team+position (fuzzy match)

### 11.4 Mid-Draft Disconnection

**Scenario:** Browser closes or internet connection drops mid-draft

**Requirements:**
- Draft state persists in database
- Reload page shows current state
- No picks lost
- Can resume from last pick

**Implementation:**
- All state in database, not session
- Draft ID in URL for easy return
- Auto-save after every action
- Show "last updated" timestamp
- Refresh button to sync

### 11.5 Multiple Concurrent Viewers

**Scenario:** Multiple people viewing draft at same time

**Requirements:**
- All viewers see same state
- Updates propagate to all viewers
- No conflicting picks possible

**Implementation Options:**
1. **Polling:** Clients poll server every N seconds
2. **SSE (Server-Sent Events):** Server pushes updates to clients
3. **WebSockets:** Bidirectional real-time communication

**Recommendation:** SSE for simplicity
- Server sends event when pick made
- Clients listen for "pick" events
- Update UI automatically
- Fallback to polling if SSE unavailable

### 11.6 Very Large Drafts

**Scenario:** 14 teams, 30 rounds = 420 picks

**Considerations:**
- Page load time with 420 pick cells
- Database query performance
- UI rendering performance

**Optimizations:**
- Paginate draft board (show 10 rounds at a time)
- Virtual scrolling for player list
- Index database properly
- Cache available players query
- Lazy load team rosters

### 11.7 Interrupted Draft

**Scenario:** Draft paused for hours/days, resume later

**Requirements:**
- Save draft state
- Mark draft as "paused" or "in progress"
- Easy resume from exact point
- No data loss

**Implementation:**
- Draft status field: "setup", "active", "paused", "completed"
- Commissioner can pause/resume
- All data persists across sessions
- Show "paused" indicator in UI
- Button to "Resume Draft"

### 11.8 Draft Completion

**Scenario:** All picks made, draft complete

**Requirements:**
- Detect draft completion
- Mark draft as complete
- Lock draft from further picks
- Show completion message
- Redirect to results page

**Detection Logic:**
```
total_possible_picks = num_teams * num_rounds
if picks_made >= total_possible_picks:
    draft.status = "completed"
```

**OR:**
- Commissioner manually marks complete
- "End Draft" button always available

### 11.9 Accidental Double Click

**Scenario:** User double-clicks player, tries to draft twice

**Prevention:**
- Disable button after first click
- Show loading spinner
- Prevent duplicate API calls
- Return error if player already drafted

**UI Feedback:**
- "Drafting..." message
- Button disabled and greyed
- Success message on completion
- Error message if failed

### 11.10 Player Not Found in Database

**Scenario:** Commissioner wants to draft player not in list (rookie, updated roster)

**Options:**
1. **Manual Entry:**
   - Allow commissioner to add custom player
   - Enter: name, team, position
   - Rank = null or 9999
   
2. **Import Additional Players:**
   - Upload CSV with new players
   - Append to database
   - Available immediately

**Webapp Recommendation:**
- Support manual entry in UI
- Simple form: "Add Custom Player"
- Fields: Name, Team, Position
- Automatically added to database
- Available for drafting

### 11.11 ADP Updates

**Scenario:** ADP data updated mid-season or mid-draft

**Current Excel Behavior:**
- Download new version of spreadsheet
- Manually copy team names and picks
- Start over with new ADP data

**Webapp Recommendation:**
- Allow ADP data refresh without losing draft state
- Keep picks separate from player database
- Update player ranks without affecting completed picks
- Show "ADP updated" notification

---

## 12. Technical Specifications

### 12.1 Performance Requirements

**Response Time:**
- Page load: < 2 seconds
- Pick submission: < 500ms
- Search/filter: < 300ms
- Auto-refresh: Every 2-5 seconds

**Concurrent Users:**
- Support 20+ simultaneous viewers
- Support 1 active drafter at a time
- Handle multiple drafts simultaneously

**Data Volume:**
- 1,500 players
- Up to 420 picks per draft (14 teams × 30 rounds)
- Up to 100 concurrent drafts
- 5-year retention of completed drafts

### 12.2 Browser Compatibility

**Supported Browsers:**
- Chrome 90+
- Firefox 88+
- Safari 14+
- Edge 90+

**Mobile Support:**
- iOS Safari 14+
- Android Chrome 90+
- Responsive design for tablets

### 12.3 Data Retention

**Active Drafts:**
- Keep indefinitely until manually deleted
- Auto-delete drafts abandoned for 1+ year (optional)

**Completed Drafts:**
- Archive after completion
- Keep for historical reference
- Allow export before deletion

**Player Database:**
- Update annually before season
- Keep historical player data (optional)

### 12.4 Security Requirements

**Access Control:**
- Draft creator is commissioner (owner)
- Commissioner can edit settings
- All viewers can see draft (read-only)
- Optional: Password protect draft

**Data Privacy:**
- No personal information required
- Team names and owner names are public within draft
- No email, no signup (optional)

**Input Validation:**
- Sanitize all user inputs
- Prevent SQL injection
- Prevent XSS attacks
- Rate limiting on API endpoints

---

## 13. API Endpoints (If Needed)

### 13.1 Draft Management

```
GET    /api/drafts                  # List all drafts
POST   /api/drafts                  # Create new draft
GET    /api/drafts/:id              # Get draft details
PUT    /api/drafts/:id              # Update draft settings
DELETE /api/drafts/:id              # Delete draft
POST   /api/drafts/:id/start        # Start draft
POST   /api/drafts/:id/pause        # Pause draft
POST   /api/drafts/:id/complete     # Complete draft
```

### 13.2 Team Management

```
GET    /api/drafts/:id/teams        # List teams in draft
POST   /api/drafts/:id/teams        # Add team
PUT    /api/teams/:id               # Update team
DELETE /api/teams/:id               # Delete team
GET    /api/teams/:id/roster        # Get team roster
```

### 13.3 Pick Management

```
GET    /api/drafts/:id/picks        # List all picks
POST   /api/drafts/:id/picks        # Make a pick
DELETE /api/picks/:id               # Undo pick (delete last)
GET    /api/drafts/:id/current      # Get current pick info
```

### 13.4 Player Queries

```
GET    /api/players                 # List all players (with filters)
GET    /api/players/:id             # Get player details
GET    /api/drafts/:id/available    # Get available players
POST   /api/players                 # Add custom player
```

### 13.5 Statistics

```
GET    /api/drafts/:id/stats        # Overall draft stats
GET    /api/drafts/:id/stats/teams  # Stats by team
GET    /api/teams/:id/stats         # Individual team stats
```

### 13.6 Real-time Updates

```
GET    /api/drafts/:id/stream       # SSE endpoint for live updates
```

---

## 14. Success Metrics

### 14.1 Functionality Metrics

- ✅ 100% of picks recorded successfully
- ✅ 0 duplicate players drafted
- ✅ 0 out-of-order picks
- ✅ 100% snake draft accuracy
- ✅ < 1% error rate on player selection

### 14.2 Performance Metrics

- ✅ < 2 second page load
- ✅ < 500ms pick submission
- ✅ < 5 second update propagation
- ✅ 99.9% uptime during draft
- ✅ Support 20+ concurrent viewers

### 14.3 Usability Metrics

- ✅ < 3 clicks to make a pick
- ✅ < 10 seconds to find player (with search)
- ✅ 0 confusion about whose turn it is
- ✅ Mobile accessible
- ✅ Minimal training required

---

## 15. Glossary

**ADP (Average Draft Position):** Statistical average of when a player is typically drafted across many drafts.

**Big Board:** Grid view showing all teams and their picks by round.

**Commissioner:** Person organizing and managing the draft.

**Dynasty League:** Multi-year fantasy league where rosters carry over from season to season.

**IDP (Individual Defensive Player):** Defensive positions (DL, LB, DB) scored individually rather than as team defense.

**Keeper League:** League where teams can retain selected players from previous season.

**Mock Draft:** Practice draft used for preparation.

**PPR (Points Per Reception):** Scoring format awarding points for receptions.

**Redraft:** Single-season fantasy league where teams draft fresh each year.

**Snake Draft:** Draft format where pick order reverses each round.

**Superflex (SF):** Roster position that can be filled by any offensive player, including QB, increasing QB value.

---

## 16. Future Considerations

**Post-MVP Features:**
- Auction draft support (bidding instead of snake)
- Best Ball draft (no weekly management)
- Draft timer with countdown
- Player notes and tags
- Draft room chat
- Commissioner overrides
- Advanced analytics dashboard
- Integration with fantasy platforms
- Mobile native app
- Offline mode
- Multi-language support

---

## Document History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2025-01-XX | Claude | Initial specification based on Excel workbook v1.2.2 |

---

## Appendix A: Excel Workbook Structure

**Sheets in Original Workbook:**
1. HOME - Landing page with info
2. Settings - League configuration
3. Teams & Draft Order - Team setup
4. Draft (10 Teams) - Main draft sheet for 10-team league
5. Draft (8 Teams) - Draft sheet for 8-team league
6. Draft (12 Teams) - Draft sheet for 12-team league
7. Draft (14 Teams) - Draft sheet for 14-team league
8. Big Board - Grid view of all picks
9. Players Drafted - List by position
10. Stats by Team - Position counts per team
11. Stats by Franchise - Players drafted per NFL team
12. HELP_PlayerDB - Full player database (1472 players)
13. HELP_PlayerDB_Filtered - Filtered by enabled positions
14. HELP_PlayersLeft - Available players (not drafted)
15. HELP_HIDDEN - Helper sheet for Big Board formulas
16. FAQ - Frequently asked questions

**Key Formulas:**
- Snake draft team calculation (alternating order)
- Available players filter (exclude drafted)
- Position counts (COUNTIFS by team and position)
- Big Board player display (per team per round)
- Stats aggregation

---

## Appendix B: Sample Data

**Sample Team Configuration (10 Teams):**
1. Go Long Baby (Steve)
2. Fueled By Bourbon (Robert L)
3. Andy Reid's Mustache (Daniel R)
4. Charlie's Angels (Audra)
5. Team 'Cole' Man (Gerry)
6. Jeff's Airmen (Jeff)
7. Bub's Bros (Alan)
8. Memes & Caffeine (Michael)
9. Dan's Demons (Dan B)
10. Robert's Rams (Robert S)

**Sample Draft Picks (First 10 Picks, 10-Team League):**
1. Ja'Marr Chase (CIN, WR) - Team 1
2. Saquon Barkley (PHI, RB) - Team 2
3. Justin Jefferson (MIN, WR) - Team 3
4. Bijan Robinson (ATL, RB) - Team 4
5. Jahmyr Gibbs (DET, RB) - Team 5
6. Puka Nacua (LAR, WR) - Team 6
7. CeeDee Lamb (DAL, WR) - Team 7
8. Derrick Henry (BAL, RB) - Team 8
9. Amon-Ra St. Brown (DET, WR) - Team 9
10. Ashton Jeanty (LVR, RB) - Team 10

**Round 2 (Snake Reversal):**
11. Christian McCaffrey (SFO, RB) - Team 10 (reversed)
12. Nico Collins (HOU, WR) - Team 9 (reversed)
... continues in reverse order

---

**END OF SPECIFICATION DOCUMENT**
