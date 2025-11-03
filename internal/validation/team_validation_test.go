package validation

import (
	"testing"

	"github.com/vibes/draft-board/internal/models"
)

func TestValidateTeam(t *testing.T) {
	existingTeams := []models.Team{
		{ID: 1, TeamName: "Team Alpha", DraftPosition: 1},
		{ID: 2, TeamName: "Team Beta", DraftPosition: 2},
		{ID: 3, TeamName: "Team Gamma", DraftPosition: 3},
	}

	tests := []struct {
		name          string
		team          *models.Team
		existingTeams []models.Team
		numTeams      int
		wantErr       error
	}{
		{
			name: "valid new team",
			team: &models.Team{
				ID:            4,
				TeamName:      "Team Delta",
				DraftPosition: 4,
			},
			existingTeams: existingTeams,
			numTeams:      12,
			wantErr:       nil,
		},
		{
			name: "valid team - position 1",
			team: &models.Team{
				ID:            5,
				TeamName:      "Team Echo",
				DraftPosition: 1,
			},
			existingTeams: []models.Team{},
			numTeams:      12,
			wantErr:       nil,
		},
		{
			name: "valid team - max position",
			team: &models.Team{
				ID:            5,
				TeamName:      "Team Echo",
				DraftPosition: 12,
			},
			existingTeams: []models.Team{},
			numTeams:      12,
			wantErr:       nil,
		},
		{
			name: "valid update - same team name",
			team: &models.Team{
				ID:            1,
				TeamName:      "Team Alpha", // Same name as existing
				DraftPosition: 5,
			},
			existingTeams: existingTeams,
			numTeams:      12,
			wantErr:       nil,
		},
		{
			name: "valid update - same draft position",
			team: &models.Team{
				ID:            1,
				TeamName:      "Updated Alpha",
				DraftPosition: 1, // Same position as existing
			},
			existingTeams: existingTeams,
			numTeams:      12,
			wantErr:       nil,
		},
		{
			name: "invalid - empty team name",
			team: &models.Team{
				ID:            4,
				TeamName:      "",
				DraftPosition: 4,
			},
			existingTeams: existingTeams,
			numTeams:      12,
			wantErr:       ErrTeamNameRequired,
		},
		{
			name: "invalid - team name too long",
			team: &models.Team{
				ID:            4,
				TeamName:      "This is a very long team name that exceeds the maximum allowed length of fifty characters",
				DraftPosition: 4,
			},
			existingTeams: existingTeams,
			numTeams:      12,
			wantErr:       ErrTeamNameTooLong,
		},
		{
			name: "valid - team name exactly 50 characters",
			team: &models.Team{
				ID:            4,
				TeamName:      "12345678901234567890123456789012345678901234567890", // Exactly 50
				DraftPosition: 4,
			},
			existingTeams: existingTeams,
			numTeams:      12,
			wantErr:       nil,
		},
		{
			name: "invalid - team name 51 characters",
			team: &models.Team{
				ID:            4,
				TeamName:      "123456789012345678901234567890123456789012345678901", // 51
				DraftPosition: 4,
			},
			existingTeams: existingTeams,
			numTeams:      12,
			wantErr:       ErrTeamNameTooLong,
		},
		{
			name: "invalid - draft position too low",
			team: &models.Team{
				ID:            4,
				TeamName:      "Team Delta",
				DraftPosition: 0,
			},
			existingTeams: existingTeams,
			numTeams:      12,
			wantErr:       ErrInvalidDraftPosition,
		},
		{
			name: "invalid - draft position negative",
			team: &models.Team{
				ID:            4,
				TeamName:      "Team Delta",
				DraftPosition: -1,
			},
			existingTeams: existingTeams,
			numTeams:      12,
			wantErr:       ErrInvalidDraftPosition,
		},
		{
			name: "invalid - draft position too high",
			team: &models.Team{
				ID:            4,
				TeamName:      "Team Delta",
				DraftPosition: 13,
			},
			existingTeams: existingTeams,
			numTeams:      12,
			wantErr:       ErrInvalidDraftPosition,
		},
		{
			name: "invalid - duplicate team name",
			team: &models.Team{
				ID:            4,
				TeamName:      "Team Alpha", // Already exists
				DraftPosition: 4,
			},
			existingTeams: existingTeams,
			numTeams:      12,
			wantErr:       ErrDuplicateTeamName,
		},
		{
			name: "invalid - duplicate draft position",
			team: &models.Team{
				ID:            4,
				TeamName:      "Team Delta",
				DraftPosition: 1, // Already taken by Team Alpha
			},
			existingTeams: existingTeams,
			numTeams:      12,
			wantErr:       ErrDuplicateDraftPos,
		},
		{
			name: "invalid - both duplicate name and position",
			team: &models.Team{
				ID:            4,
				TeamName:      "Team Alpha",
				DraftPosition: 1,
			},
			existingTeams: existingTeams,
			numTeams:      12,
			wantErr:       ErrDuplicateTeamName, // First error encountered
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTeam(tt.team, tt.existingTeams, tt.numTeams)
			if err != tt.wantErr {
				t.Errorf("ValidateTeam() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateTeamRosterCount(t *testing.T) {
	tests := []struct {
		name      string
		teamCount int
		numTeams  int
		wantErr   error
	}{
		{
			name:      "valid - correct team count",
			teamCount: 12,
			numTeams:  12,
			wantErr:   nil,
		},
		{
			name:      "valid - minimum teams",
			teamCount: 2,
			numTeams:  2,
			wantErr:   nil,
		},
		{
			name:      "valid - maximum teams",
			teamCount: 14,
			numTeams:  14,
			wantErr:   nil,
		},
		{
			name:      "invalid - too few teams",
			teamCount: 11,
			numTeams:  12,
			wantErr:   ErrIncompleteTeamRoster,
		},
		{
			name:      "invalid - too many teams",
			teamCount: 13,
			numTeams:  12,
			wantErr:   ErrIncompleteTeamRoster,
		},
		{
			name:      "invalid - zero teams",
			teamCount: 0,
			numTeams:  12,
			wantErr:   ErrIncompleteTeamRoster,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTeamRosterCount(tt.teamCount, tt.numTeams)
			if err != tt.wantErr {
				t.Errorf("ValidateTeamRosterCount() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
