package validation

import (
	"testing"

	"github.com/vibes/draft-board/internal/models"
)

func TestValidateDraft(t *testing.T) {
	tests := []struct {
		name    string
		draft   *models.Draft
		wantErr error
	}{
		{
			name: "valid draft - standard scoring",
			draft: &models.Draft{
				Name:          "My League",
				NumTeams:      12,
				ScoringFormat: "Standard",
				DraftType:     "Redraft",
			},
			wantErr: nil,
		},
		{
			name: "valid draft - PPR scoring",
			draft: &models.Draft{
				Name:          "PPR League",
				NumTeams:      10,
				ScoringFormat: "PPR",
				DraftType:     "Redraft",
			},
			wantErr: nil,
		},
		{
			name: "valid draft - Half-PPR scoring",
			draft: &models.Draft{
				Name:          "Half PPR League",
				NumTeams:      8,
				ScoringFormat: "Half-PPR",
				DraftType:     "Dynasty",
			},
			wantErr: nil,
		},
		{
			name: "valid draft - Dynasty type",
			draft: &models.Draft{
				Name:          "Dynasty League",
				NumTeams:      14,
				ScoringFormat: "PPR",
				DraftType:     "Dynasty",
			},
			wantErr: nil,
		},
		{
			name: "valid draft - minimum teams",
			draft: &models.Draft{
				Name:          "Small League",
				NumTeams:      2,
				ScoringFormat: "Standard",
				DraftType:     "Redraft",
			},
			wantErr: nil,
		},
		{
			name: "valid draft - maximum teams",
			draft: &models.Draft{
				Name:          "Large League",
				NumTeams:      14,
				ScoringFormat: "Standard",
				DraftType:     "Redraft",
			},
			wantErr: nil,
		},
		{
			name: "invalid - no name",
			draft: &models.Draft{
				Name:          "",
				NumTeams:      12,
				ScoringFormat: "Standard",
				DraftType:     "Redraft",
			},
			wantErr: ErrDraftNameRequired,
		},
		{
			name: "invalid - too few teams",
			draft: &models.Draft{
				Name:          "My League",
				NumTeams:      1,
				ScoringFormat: "Standard",
				DraftType:     "Redraft",
			},
			wantErr: ErrInvalidLeagueSize,
		},
		{
			name: "invalid - too many teams",
			draft: &models.Draft{
				Name:          "My League",
				NumTeams:      15,
				ScoringFormat: "Standard",
				DraftType:     "Redraft",
			},
			wantErr: ErrInvalidLeagueSize,
		},
		{
			name: "invalid - zero teams",
			draft: &models.Draft{
				Name:          "My League",
				NumTeams:      0,
				ScoringFormat: "Standard",
				DraftType:     "Redraft",
			},
			wantErr: ErrInvalidLeagueSize,
		},
		{
			name: "invalid - negative teams",
			draft: &models.Draft{
				Name:          "My League",
				NumTeams:      -5,
				ScoringFormat: "Standard",
				DraftType:     "Redraft",
			},
			wantErr: ErrInvalidLeagueSize,
		},
		{
			name: "invalid - unknown scoring format",
			draft: &models.Draft{
				Name:          "My League",
				NumTeams:      12,
				ScoringFormat: "Unknown",
				DraftType:     "Redraft",
			},
			wantErr: ErrInvalidScoringFormat,
		},
		{
			name: "invalid - empty scoring format",
			draft: &models.Draft{
				Name:          "My League",
				NumTeams:      12,
				ScoringFormat: "",
				DraftType:     "Redraft",
			},
			wantErr: ErrInvalidScoringFormat,
		},
		{
			name: "invalid - case sensitive scoring format",
			draft: &models.Draft{
				Name:          "My League",
				NumTeams:      12,
				ScoringFormat: "standard", // lowercase
				DraftType:     "Redraft",
			},
			wantErr: ErrInvalidScoringFormat,
		},
		{
			name: "invalid - unknown draft type",
			draft: &models.Draft{
				Name:          "My League",
				NumTeams:      12,
				ScoringFormat: "Standard",
				DraftType:     "Unknown",
			},
			wantErr: ErrInvalidDraftType,
		},
		{
			name: "invalid - empty draft type",
			draft: &models.Draft{
				Name:          "My League",
				NumTeams:      12,
				ScoringFormat: "Standard",
				DraftType:     "",
			},
			wantErr: ErrInvalidDraftType,
		},
		{
			name: "invalid - case sensitive draft type",
			draft: &models.Draft{
				Name:          "My League",
				NumTeams:      12,
				ScoringFormat: "Standard",
				DraftType:     "redraft", // lowercase
			},
			wantErr: ErrInvalidDraftType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDraft(tt.draft)
			if err != tt.wantErr {
				t.Errorf("ValidateDraft() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
