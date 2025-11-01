package validation

import (
	"github.com/vibes/draft-board/internal/models"
	"github.com/vibes/draft-board/internal/snake"
)

func ValidatePick(pick *models.Pick, draft *models.Draft, teams []models.Team, pickCount int) error {
	// Validate pick is sequential
	if pick.OverallPick != pickCount+1 {
		return ErrInvalidPickNumber
	}

	// Validate correct team's turn
	snakeTeams := make([]snake.Team, len(teams))
	for i, t := range teams {
		snakeTeams[i] = snake.Team{
			ID:            t.ID,
			DraftPosition: t.DraftPosition,
		}
	}

	currentTeam, err := snake.CalculateCurrentTeam(pick.OverallPick, draft.NumTeams, snakeTeams)
	if err != nil || currentTeam.ID != pick.TeamID {
		return ErrNotTeamTurn
	}

	// Validate draft is active
	if !draft.CanMakePicks() {
		return ErrDraftNotActive
	}

	return nil
}

func ValidatePlayerNotDrafted(playerID int, draftedPlayerIDs []int) error {
	for _, draftedID := range draftedPlayerIDs {
		if draftedID == playerID {
			return ErrPlayerAlreadyDrafted
		}
	}
	return nil
}

func ValidateSearchQuery(query string) error {
	if len(query) > 50 {
		return ErrSearchQueryTooLong
	}
	return nil
}

func ValidatePosition(position string) error {
	validPositions := map[string]bool{
		"QB": true, "RB": true, "WR": true, "TE": true,
		"K": true, "D/ST": true, "DL": true, "LB": true, "DB": true,
	}
	if !validPositions[position] {
		return ErrInvalidPosition
	}
	return nil
}

