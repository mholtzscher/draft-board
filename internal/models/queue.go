package models

import "time"

type QueueItem struct {
	ID         int       `db:"id"`
	DraftID    int       `db:"draft_id"`
	TeamID     int       `db:"team_id"`
	PlayerID   int       `db:"player_id"`
	QueueOrder int       `db:"queue_order"`
	AddedAt    time.Time `db:"added_at"`
}

