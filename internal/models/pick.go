package models

import "time"

type Pick struct {
	ID          int       `db:"id"`
	DraftID     int       `db:"draft_id"`
	TeamID      int       `db:"team_id"`
	PlayerID    int       `db:"player_id"`
	Round       int       `db:"round"`
	OverallPick int       `db:"overall_pick"`
	IsTraded    bool      `db:"is_traded"`
	ADPRank     *int      `db:"adp_rank"`
	PickedAt    time.Time `db:"picked_at"`
}

