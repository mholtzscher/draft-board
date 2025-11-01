package validation

import "errors"

var (
	ErrInvalidLeagueSize    = errors.New("invalid league size. Must be between 2 and 14 teams")
	ErrInvalidScoringFormat = errors.New("invalid scoring format. Must be Standard, Half-PPR, or PPR")
	ErrInvalidDraftType     = errors.New("invalid draft type. Must be Redraft or Dynasty")
	ErrDraftNameRequired    = errors.New("draft name is required")
	ErrTeamNameRequired     = errors.New("team name is required")
	ErrTeamNameTooLong      = errors.New("team name must be between 1 and 50 characters")
	ErrDuplicateTeamName    = errors.New("team name already exists in this draft")
	ErrInvalidDraftPosition = errors.New("draft position must be between 1 and N")
	ErrDuplicateDraftPos    = errors.New("draft position already assigned")
	ErrIncompleteTeamRoster = errors.New("must have exactly N teams")
	ErrInvalidPlayer        = errors.New("invalid player ID")
	ErrPlayerAlreadyDrafted = errors.New("player has already been drafted")
	ErrInvalidTeam          = errors.New("invalid team ID")
	ErrNotTeamTurn          = errors.New("not this team's turn to pick")
	ErrDraftNotActive       = errors.New("cannot make picks in completed draft")
	ErrInvalidPickNumber    = errors.New("pick number must be sequential")
	ErrSearchQueryTooLong   = errors.New("search query too long (max 50 characters)")
	ErrInvalidPosition      = errors.New("invalid position filter")
	ErrInvalidSortOption     = errors.New("invalid sort option")
)

