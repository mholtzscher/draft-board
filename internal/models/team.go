package models

type Team struct {
	ID            int    `db:"id"`
	DraftID       int    `db:"draft_id"`
	TeamName      string `db:"team_name"`
	OwnerName     string `db:"owner_name"`
	DraftPosition int    `db:"draft_position"`
}

