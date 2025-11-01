package models

import "time"

type AuditLog struct {
	ID          int       `db:"id"`
	DraftID     int       `db:"draft_id"`
	ActionType  string    `db:"action_type"`
	EntityID    *int      `db:"entity_id"`
	Details     string    `db:"details"`
	PerformedAt time.Time `db:"performed_at"`
}

type PositionSetting struct {
	ID       int    `db:"id"`
	DraftID int    `db:"draft_id"`
	Position string `db:"position"`
	Enabled  bool   `db:"enabled"`
}

