package draft

import "github.com/vibes/draft-board/internal/models"

func CheckDraftCompletion(draft *models.Draft, pickCount int) bool {
	if draft.MaxRounds > 0 {
		maxPicks := draft.NumTeams * draft.MaxRounds
		return pickCount >= maxPicks
	}

	if draft.Status == "completed" {
		return true
	}

	return false
}

