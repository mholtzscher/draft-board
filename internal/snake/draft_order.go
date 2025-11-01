package snake

import (
	"errors"
	"fmt"
	"math"
)

type Team struct {
	ID            int
	DraftPosition int
}

func CalculateCurrentTeam(pickNumber int, numTeams int, teams []Team) (*Team, error) {
	if pickNumber < 1 {
		return nil, errors.New("invalid pick number")
	}

	if numTeams <= 0 {
		return nil, errors.New("invalid number of teams")
	}

	round := int(math.Ceil(float64(pickNumber) / float64(numTeams)))
	positionInRound := ((pickNumber - 1) % numTeams) + 1

	var draftPosition int
	if round%2 == 1 {
		// Odd rounds (1, 3, 5, ...): normal order (1, 2, 3, ...)
		draftPosition = positionInRound
	} else {
		// Even rounds (2, 4, 6, ...): reverse order (N, N-1, N-2, ...)
		draftPosition = numTeams - positionInRound + 1
	}

	for _, team := range teams {
		if team.DraftPosition == draftPosition {
			return &team, nil
		}
	}

	return nil, fmt.Errorf("no team found for draft position %d", draftPosition)
}

func CalculateRound(pickNumber int, numTeams int) int {
	if numTeams <= 0 {
		return 0
	}
	return int(math.Ceil(float64(pickNumber) / float64(numTeams)))
}

