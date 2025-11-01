package draft

import (
	"testing"

	"github.com/vibes/draft-board/internal/models"
)

func TestCheckDraftCompletion(t *testing.T) {
	tests := []struct {
		name      string
		draft     *models.Draft
		pickCount int
		want      bool
	}{
		{
			name: "draft not complete - picks remaining",
			draft: &models.Draft{
				NumTeams:  12,
				MaxRounds: 15,
				Status:    "active",
			},
			pickCount: 100,
			want:      false,
		},
		{
			name: "draft complete - exact pick count",
			draft: &models.Draft{
				NumTeams:  12,
				MaxRounds: 15,
				Status:    "active",
			},
			pickCount: 180, // 12 * 15
			want:      true,
		},
		{
			name: "draft complete - exceeded pick count",
			draft: &models.Draft{
				NumTeams:  12,
				MaxRounds: 15,
				Status:    "active",
			},
			pickCount: 200,
			want:      true,
		},
		{
			name: "draft not complete - status completed but max rounds not met",
			draft: &models.Draft{
				NumTeams:  12,
				MaxRounds: 15,
				Status:    "completed",
			},
			pickCount: 50,
			want:      false, // MaxRounds check happens first
		},
		{
			name: "draft not complete - no max rounds set",
			draft: &models.Draft{
				NumTeams:  10,
				MaxRounds: 0,
				Status:    "active",
			},
			pickCount: 100,
			want:      false,
		},
		{
			name: "draft complete - status completed, no max rounds",
			draft: &models.Draft{
				NumTeams:  10,
				MaxRounds: 0,
				Status:    "completed",
			},
			pickCount: 50,
			want:      true,
		},
		{
			name: "10 team league - complete",
			draft: &models.Draft{
				NumTeams:  10,
				MaxRounds: 16,
				Status:    "active",
			},
			pickCount: 160,
			want:      true,
		},
		{
			name: "10 team league - not complete",
			draft: &models.Draft{
				NumTeams:  10,
				MaxRounds: 16,
				Status:    "active",
			},
			pickCount: 159,
			want:      false,
		},
		{
			name: "zero picks - not complete",
			draft: &models.Draft{
				NumTeams:  12,
				MaxRounds: 15,
				Status:    "active",
			},
			pickCount: 0,
			want:      false,
		},
		{
			name: "paused draft with all picks made",
			draft: &models.Draft{
				NumTeams:  8,
				MaxRounds: 10,
				Status:    "paused",
			},
			pickCount: 80,
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CheckDraftCompletion(tt.draft, tt.pickCount)
			if got != tt.want {
				t.Errorf("CheckDraftCompletion() = %v, want %v", got, tt.want)
			}
		})
	}
}
