package snake

import (
	"testing"
)

func TestCalculateRound(t *testing.T) {
	tests := []struct {
		name       string
		pickNumber int
		numTeams   int
		want       int
	}{
		{
			name:       "first pick of round 1",
			pickNumber: 1,
			numTeams:   12,
			want:       1,
		},
		{
			name:       "last pick of round 1",
			pickNumber: 12,
			numTeams:   12,
			want:       1,
		},
		{
			name:       "first pick of round 2",
			pickNumber: 13,
			numTeams:   12,
			want:       2,
		},
		{
			name:       "middle pick",
			pickNumber: 25,
			numTeams:   12,
			want:       3,
		},
		{
			name:       "10 team league round 1",
			pickNumber: 5,
			numTeams:   10,
			want:       1,
		},
		{
			name:       "10 team league round 2",
			pickNumber: 15,
			numTeams:   10,
			want:       2,
		},
		{
			name:       "invalid numTeams zero",
			pickNumber: 5,
			numTeams:   0,
			want:       0,
		},
		{
			name:       "invalid numTeams negative",
			pickNumber: 5,
			numTeams:   -1,
			want:       0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateRound(tt.pickNumber, tt.numTeams)
			if got != tt.want {
				t.Errorf("CalculateRound(%d, %d) = %d, want %d",
					tt.pickNumber, tt.numTeams, got, tt.want)
			}
		})
	}
}

func TestCalculateCurrentTeam(t *testing.T) {
	// Create a standard 12-team league
	teams12 := make([]Team, 12)
	for i := 0; i < 12; i++ {
		teams12[i] = Team{
			ID:            i + 1,
			DraftPosition: i + 1,
		}
	}

	// Create a 10-team league
	teams10 := make([]Team, 10)
	for i := 0; i < 10; i++ {
		teams10[i] = Team{
			ID:            i + 1,
			DraftPosition: i + 1,
		}
	}

	tests := []struct {
		name          string
		pickNumber    int
		numTeams      int
		teams         []Team
		wantTeamPos   int // Expected draft position
		wantErr       bool
		wantErrString string
	}{
		// Round 1 (odd round - normal order)
		{
			name:        "pick 1 - first team",
			pickNumber:  1,
			numTeams:    12,
			teams:       teams12,
			wantTeamPos: 1,
			wantErr:     false,
		},
		{
			name:        "pick 6 - middle of round 1",
			pickNumber:  6,
			numTeams:    12,
			teams:       teams12,
			wantTeamPos: 6,
			wantErr:     false,
		},
		{
			name:        "pick 12 - last of round 1",
			pickNumber:  12,
			numTeams:    12,
			teams:       teams12,
			wantTeamPos: 12,
			wantErr:     false,
		},
		// Round 2 (even round - reverse order)
		{
			name:        "pick 13 - first of round 2 (snake back)",
			pickNumber:  13,
			numTeams:    12,
			teams:       teams12,
			wantTeamPos: 12, // Team 12 picks again
			wantErr:     false,
		},
		{
			name:        "pick 18 - middle of round 2",
			pickNumber:  18,
			numTeams:    12,
			teams:       teams12,
			wantTeamPos: 7, // Reverse order: 12, 11, 10, 9, 8, 7
			wantErr:     false,
		},
		{
			name:        "pick 24 - last of round 2",
			pickNumber:  24,
			numTeams:    12,
			teams:       teams12,
			wantTeamPos: 1, // Team 1 picks again
			wantErr:     false,
		},
		// Round 3 (odd round - normal order again)
		{
			name:        "pick 25 - first of round 3",
			pickNumber:  25,
			numTeams:    12,
			teams:       teams12,
			wantTeamPos: 1,
			wantErr:     false,
		},
		{
			name:        "pick 36 - last of round 3",
			pickNumber:  36,
			numTeams:    12,
			teams:       teams12,
			wantTeamPos: 12,
			wantErr:     false,
		},
		// Round 4 (even round - reverse order)
		{
			name:        "pick 37 - first of round 4",
			pickNumber:  37,
			numTeams:    12,
			teams:       teams12,
			wantTeamPos: 12,
			wantErr:     false,
		},
		// 10-team league tests
		{
			name:        "10 team - pick 1",
			pickNumber:  1,
			numTeams:    10,
			teams:       teams10,
			wantTeamPos: 1,
			wantErr:     false,
		},
		{
			name:        "10 team - pick 10 (end of round 1)",
			pickNumber:  10,
			numTeams:    10,
			teams:       teams10,
			wantTeamPos: 10,
			wantErr:     false,
		},
		{
			name:        "10 team - pick 11 (start of round 2)",
			pickNumber:  11,
			numTeams:    10,
			teams:       teams10,
			wantTeamPos: 10, // Snake back
			wantErr:     false,
		},
		{
			name:        "10 team - pick 20 (end of round 2)",
			pickNumber:  20,
			numTeams:    10,
			teams:       teams10,
			wantTeamPos: 1,
			wantErr:     false,
		},
		// Error cases
		{
			name:          "invalid pick number - zero",
			pickNumber:    0,
			numTeams:      12,
			teams:         teams12,
			wantErr:       true,
			wantErrString: "invalid pick number",
		},
		{
			name:          "invalid pick number - negative",
			pickNumber:    -5,
			numTeams:      12,
			teams:         teams12,
			wantErr:       true,
			wantErrString: "invalid pick number",
		},
		{
			name:          "invalid numTeams - zero",
			pickNumber:    5,
			numTeams:      0,
			teams:         teams12,
			wantErr:       true,
			wantErrString: "invalid number of teams",
		},
		{
			name:          "invalid numTeams - negative",
			pickNumber:    5,
			numTeams:      -3,
			teams:         teams12,
			wantErr:       true,
			wantErrString: "invalid number of teams",
		},
		{
			name:          "team not found in list",
			pickNumber:    1,
			numTeams:      12,
			teams:         []Team{}, // Empty team list
			wantErr:       true,
			wantErrString: "no team found for draft position",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CalculateCurrentTeam(tt.pickNumber, tt.numTeams, tt.teams)

			// Check error cases
			if tt.wantErr {
				if err == nil {
					t.Errorf("CalculateCurrentTeam() expected error but got nil")
					return
				}
				if tt.wantErrString != "" && err.Error() != tt.wantErrString {
					// Check if error message contains expected string
					if len(err.Error()) < len(tt.wantErrString) ||
						err.Error()[:len(tt.wantErrString)] != tt.wantErrString {
						t.Errorf("CalculateCurrentTeam() error = %v, want error containing %v",
							err, tt.wantErrString)
					}
				}
				return
			}

			// Check success cases
			if err != nil {
				t.Errorf("CalculateCurrentTeam() unexpected error = %v", err)
				return
			}

			if got == nil {
				t.Errorf("CalculateCurrentTeam() returned nil team")
				return
			}

			if got.DraftPosition != tt.wantTeamPos {
				t.Errorf("CalculateCurrentTeam() got team with draft position %d, want %d",
					got.DraftPosition, tt.wantTeamPos)
			}
		})
	}
}

func TestSnakeDraftPattern(t *testing.T) {
	// Test the complete snake pattern for first few rounds
	teams := make([]Team, 4)
	for i := 0; i < 4; i++ {
		teams[i] = Team{
			ID:            i + 1,
			DraftPosition: i + 1,
		}
	}

	// Expected pattern for 4-team snake draft:
	// Round 1: 1, 2, 3, 4
	// Round 2: 4, 3, 2, 1
	// Round 3: 1, 2, 3, 4
	// Round 4: 4, 3, 2, 1
	expectedPattern := []int{
		1, 2, 3, 4, // Round 1
		4, 3, 2, 1, // Round 2
		1, 2, 3, 4, // Round 3
		4, 3, 2, 1, // Round 4
	}

	for pickNum, expectedPos := range expectedPattern {
		team, err := CalculateCurrentTeam(pickNum+1, 4, teams)
		if err != nil {
			t.Errorf("Pick %d: unexpected error: %v", pickNum+1, err)
			continue
		}
		if team.DraftPosition != expectedPos {
			t.Errorf("Pick %d: got position %d, want %d",
				pickNum+1, team.DraftPosition, expectedPos)
		}
	}
}

func TestTeamWithNonSequentialDraftPositions(t *testing.T) {
	// Test with teams that have non-sequential draft positions
	// (e.g., due to draft position trading)
	teams := []Team{
		{ID: 1, DraftPosition: 3},
		{ID: 2, DraftPosition: 1},
		{ID: 3, DraftPosition: 4},
		{ID: 4, DraftPosition: 2},
	}

	tests := []struct {
		pickNumber int
		wantTeamID int
	}{
		{1, 2},  // Draft position 1 -> Team ID 2
		{2, 4},  // Draft position 2 -> Team ID 4
		{3, 1},  // Draft position 3 -> Team ID 1
		{4, 3},  // Draft position 4 -> Team ID 3
		{5, 3},  // Round 2, position 4 (reverse) -> Team ID 3
		{6, 1},  // Round 2, position 3 (reverse) -> Team ID 1
		{7, 4},  // Round 2, position 2 (reverse) -> Team ID 4
		{8, 2},  // Round 2, position 1 (reverse) -> Team ID 2
	}

	for _, tt := range tests {
		t.Run("pick_"+string(rune(tt.pickNumber+'0')), func(t *testing.T) {
			got, err := CalculateCurrentTeam(tt.pickNumber, 4, teams)
			if err != nil {
				t.Errorf("Pick %d: unexpected error: %v", tt.pickNumber, err)
				return
			}
			if got.ID != tt.wantTeamID {
				t.Errorf("Pick %d: got team ID %d, want %d",
					tt.pickNumber, got.ID, tt.wantTeamID)
			}
		})
	}
}
