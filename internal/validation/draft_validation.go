package validation

import "github.com/vibes/draft-board/internal/models"

func ValidateDraft(draft *models.Draft) error {
	if draft.Name == "" {
		return ErrDraftNameRequired
	}
	if draft.NumTeams < 2 || draft.NumTeams > 14 {
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

