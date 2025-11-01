package validation

import "github.com/vibes/draft-board/internal/models"

func ValidateTeam(team *models.Team, existingTeams []models.Team, numTeams int) error {
	if team.TeamName == "" {
		return ErrTeamNameRequired
	}
	if len(team.TeamName) > 50 {
		return ErrTeamNameTooLong
	}
	if team.DraftPosition < 1 || team.DraftPosition > numTeams {
		return ErrInvalidDraftPosition
	}
	for _, t := range existingTeams {
		if t.TeamName == team.TeamName && t.ID != team.ID {
			return ErrDuplicateTeamName
		}
		if t.DraftPosition == team.DraftPosition && t.ID != team.ID {
			return ErrDuplicateDraftPos
		}
	}
	return nil
}

func ValidateTeamRosterCount(teamCount int, numTeams int) error {
	if teamCount != numTeams {
		return ErrIncompleteTeamRoster
	}
	return nil
}

