package models

import "time"

type Draft struct {
	ID              int       `db:"id"`
	Name            string    `db:"name"`
	NumTeams        int       `db:"num_teams"`
	ScoringFormat   string    `db:"scoring_format"`
	DraftType       string    `db:"draft_type"`
	QBSetting       string    `db:"qb_setting"`
	SnakeDraft      bool      `db:"snake_draft"`
	Status          string    `db:"status"`
	MaxRounds       int       `db:"max_rounds"`
	CommissionerID  string    `db:"commissioner_id"`
	CreatedAt       time.Time `db:"created_at"`
	Completed       bool      `db:"completed"`
}

func (d *Draft) IsActive() bool {
	return d.Status == "active"
}

func (d *Draft) IsCompleted() bool {
	return d.Status == "completed" || d.Completed
}

func (d *Draft) IsPaused() bool {
	return d.Status == "paused"
}

func (d *Draft) CanMakePicks() bool {
	return d.Status == "active"
}

func (d *Draft) CheckDraftCompletion(pickCount int) bool {
	if d.MaxRounds > 0 {
		maxPicks := d.NumTeams * d.MaxRounds
		return pickCount >= maxPicks
	}

	if d.Status == "completed" {
		return true
	}

	return false
}

