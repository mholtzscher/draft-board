package validation

import (
	"testing"

	"github.com/vibes/draft-board/internal/models"
)

func TestValidatePick(t *testing.T) {
	teams := []models.Team{
		{ID: 1, DraftPosition: 1},
		{ID: 2, DraftPosition: 2},
		{ID: 3, DraftPosition: 3},
		{ID: 4, DraftPosition: 4},
	}

	tests := []struct {
		name      string
		pick      *models.Pick
		draft     *models.Draft
		teams     []models.Team
		pickCount int
		wantErr   error
	}{
		{
			name: "valid first pick",
			pick: &models.Pick{
				TeamID:      1,
				OverallPick: 1,
			},
			draft: &models.Draft{
				NumTeams: 4,
				Status:   "active",
			},
			teams:     teams,
			pickCount: 0,
			wantErr:   nil,
		},
		{
			name: "valid second pick",
			pick: &models.Pick{
				TeamID:      2,
				OverallPick: 2,
			},
			draft: &models.Draft{
				NumTeams: 4,
				Status:   "active",
			},
			teams:     teams,
			pickCount: 1,
			wantErr:   nil,
		},
		{
			name: "valid pick - round 2 snake back",
			pick: &models.Pick{
				TeamID:      4, // Team 4 picks first in round 2 (snake)
				OverallPick: 5,
			},
			draft: &models.Draft{
				NumTeams: 4,
				Status:   "active",
			},
			teams:     teams,
			pickCount: 4,
			wantErr:   nil,
		},
		{
			name: "valid pick - last of round 1",
			pick: &models.Pick{
				TeamID:      4,
				OverallPick: 4,
			},
			draft: &models.Draft{
				NumTeams: 4,
				Status:   "active",
			},
			teams:     teams,
			pickCount: 3,
			wantErr:   nil,
		},
		{
			name: "invalid - non-sequential pick number",
			pick: &models.Pick{
				TeamID:      1,
				OverallPick: 5, // Should be 1
			},
			draft: &models.Draft{
				NumTeams: 4,
				Status:   "active",
			},
			teams:     teams,
			pickCount: 0,
			wantErr:   ErrInvalidPickNumber,
		},
		{
			name: "invalid - skipping pick number",
			pick: &models.Pick{
				TeamID:      3,
				OverallPick: 4, // Should be 3
			},
			draft: &models.Draft{
				NumTeams: 4,
				Status:   "active",
			},
			teams:     teams,
			pickCount: 2,
			wantErr:   ErrInvalidPickNumber,
		},
		{
			name: "invalid - wrong team's turn",
			pick: &models.Pick{
				TeamID:      3, // Should be team 1
				OverallPick: 1,
			},
			draft: &models.Draft{
				NumTeams: 4,
				Status:   "active",
			},
			teams:     teams,
			pickCount: 0,
			wantErr:   ErrNotTeamTurn,
		},
		{
			name: "invalid - draft not active (paused)",
			pick: &models.Pick{
				TeamID:      1,
				OverallPick: 1,
			},
			draft: &models.Draft{
				NumTeams: 4,
				Status:   "paused",
			},
			teams:     teams,
			pickCount: 0,
			wantErr:   ErrDraftNotActive,
		},
		{
			name: "invalid - draft not active (completed)",
			pick: &models.Pick{
				TeamID:      1,
				OverallPick: 1,
			},
			draft: &models.Draft{
				NumTeams: 4,
				Status:   "completed",
			},
			teams:     teams,
			pickCount: 0,
			wantErr:   ErrDraftNotActive,
		},
		{
			name: "invalid - draft not active (setup)",
			pick: &models.Pick{
				TeamID:      1,
				OverallPick: 1,
			},
			draft: &models.Draft{
				NumTeams: 4,
				Status:   "setup",
			},
			teams:     teams,
			pickCount: 0,
			wantErr:   ErrDraftNotActive,
		},
		{
			name: "valid - 12 team league pick",
			pick: &models.Pick{
				TeamID:      1,
				OverallPick: 1,
			},
			draft: &models.Draft{
				NumTeams: 12,
				Status:   "active",
			},
			teams: []models.Team{
				{ID: 1, DraftPosition: 1},
				{ID: 2, DraftPosition: 2},
				{ID: 3, DraftPosition: 3},
				{ID: 4, DraftPosition: 4},
				{ID: 5, DraftPosition: 5},
				{ID: 6, DraftPosition: 6},
				{ID: 7, DraftPosition: 7},
				{ID: 8, DraftPosition: 8},
				{ID: 9, DraftPosition: 9},
				{ID: 10, DraftPosition: 10},
				{ID: 11, DraftPosition: 11},
				{ID: 12, DraftPosition: 12},
			},
			pickCount: 0,
			wantErr:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePick(tt.pick, tt.draft, tt.teams, tt.pickCount)
			if err != tt.wantErr {
				t.Errorf("ValidatePick() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidatePlayerNotDrafted(t *testing.T) {
	draftedPlayers := []int{10, 25, 42, 100, 150}

	tests := []struct {
		name              string
		playerID          int
		draftedPlayerIDs  []int
		wantErr           error
	}{
		{
			name:             "valid - player not drafted",
			playerID:         5,
			draftedPlayerIDs: draftedPlayers,
			wantErr:          nil,
		},
		{
			name:             "valid - empty drafted list",
			playerID:         5,
			draftedPlayerIDs: []int{},
			wantErr:          nil,
		},
		{
			name:             "invalid - player already drafted",
			playerID:         25,
			draftedPlayerIDs: draftedPlayers,
			wantErr:          ErrPlayerAlreadyDrafted,
		},
		{
			name:             "invalid - player at start of list",
			playerID:         10,
			draftedPlayerIDs: draftedPlayers,
			wantErr:          ErrPlayerAlreadyDrafted,
		},
		{
			name:             "invalid - player at end of list",
			playerID:         150,
			draftedPlayerIDs: draftedPlayers,
			wantErr:          ErrPlayerAlreadyDrafted,
		},
		{
			name:             "valid - similar but different ID",
			playerID:         101,
			draftedPlayerIDs: draftedPlayers,
			wantErr:          nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePlayerNotDrafted(tt.playerID, tt.draftedPlayerIDs)
			if err != tt.wantErr {
				t.Errorf("ValidatePlayerNotDrafted() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateSearchQuery(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		wantErr error
	}{
		{
			name:    "valid - short query",
			query:   "Tom Brady",
			wantErr: nil,
		},
		{
			name:    "valid - empty query",
			query:   "",
			wantErr: nil,
		},
		{
			name:    "valid - exactly 50 characters",
			query:   "12345678901234567890123456789012345678901234567890",
			wantErr: nil,
		},
		{
			name:    "invalid - 51 characters",
			query:   "123456789012345678901234567890123456789012345678901",
			wantErr: ErrSearchQueryTooLong,
		},
		{
			name:    "invalid - very long query",
			query:   "This is a very long search query that exceeds the maximum allowed length and should return an error",
			wantErr: ErrSearchQueryTooLong,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSearchQuery(tt.query)
			if err != tt.wantErr {
				t.Errorf("ValidateSearchQuery() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidatePosition(t *testing.T) {
	tests := []struct {
		name     string
		position string
		wantErr  error
	}{
		{
			name:     "valid - QB",
			position: "QB",
			wantErr:  nil,
		},
		{
			name:     "valid - RB",
			position: "RB",
			wantErr:  nil,
		},
		{
			name:     "valid - WR",
			position: "WR",
			wantErr:  nil,
		},
		{
			name:     "valid - TE",
			position: "TE",
			wantErr:  nil,
		},
		{
			name:     "valid - K",
			position: "K",
			wantErr:  nil,
		},
		{
			name:     "valid - D/ST",
			position: "D/ST",
			wantErr:  nil,
		},
		{
			name:     "valid - DL",
			position: "DL",
			wantErr:  nil,
		},
		{
			name:     "valid - LB",
			position: "LB",
			wantErr:  nil,
		},
		{
			name:     "valid - DB",
			position: "DB",
			wantErr:  nil,
		},
		{
			name:     "invalid - lowercase",
			position: "qb",
			wantErr:  ErrInvalidPosition,
		},
		{
			name:     "invalid - unknown position",
			position: "CENTER",
			wantErr:  ErrInvalidPosition,
		},
		{
			name:     "invalid - empty",
			position: "",
			wantErr:  ErrInvalidPosition,
		},
		{
			name:     "invalid - partial match",
			position: "Q",
			wantErr:  ErrInvalidPosition,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePosition(tt.position)
			if err != tt.wantErr {
				t.Errorf("ValidatePosition() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
