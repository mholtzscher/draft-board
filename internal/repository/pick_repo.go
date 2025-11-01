package repository

import (
	"database/sql"
	"fmt"

	"github.com/vibes/draft-board/internal/models"
)

type PickRepository struct {
	db *sql.DB
}

func NewPickRepository(db *sql.DB) *PickRepository {
	return &PickRepository{db: db}
}

func (r *PickRepository) Create(pick *models.Pick) error {
	query := `
		INSERT INTO picks (draft_id, team_id, player_id, round, overall_pick, is_traded, adp_rank)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.Exec(query, pick.DraftID, pick.TeamID, pick.PlayerID, pick.Round,
		pick.OverallPick, pick.IsTraded, pick.ADPRank)
	if err != nil {
		return fmt.Errorf("failed to create pick: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	pick.ID = int(id)
	return nil
}

func (r *PickRepository) GetByID(id int) (*models.Pick, error) {
	query := `SELECT * FROM picks WHERE id = ?`
	pick := &models.Pick{}
	err := r.db.QueryRow(query, id).Scan(
		&pick.ID, &pick.DraftID, &pick.TeamID, &pick.PlayerID, &pick.Round,
		&pick.OverallPick, &pick.IsTraded, &pick.ADPRank, &pick.PickedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("pick not found")
		}
		return nil, fmt.Errorf("failed to get pick: %w", err)
	}
	return pick, nil
}

func (r *PickRepository) GetByDraft(draftID int) ([]models.Pick, error) {
	query := `SELECT * FROM picks WHERE draft_id = ? ORDER BY overall_pick`
	rows, err := r.db.Query(query, draftID)
	if err != nil {
		return nil, fmt.Errorf("failed to get picks: %w", err)
	}
	defer rows.Close()

	var picks []models.Pick
	for rows.Next() {
		var pick models.Pick
		err := rows.Scan(
			&pick.ID, &pick.DraftID, &pick.TeamID, &pick.PlayerID, &pick.Round,
			&pick.OverallPick, &pick.IsTraded, &pick.ADPRank, &pick.PickedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan pick: %w", err)
		}
		picks = append(picks, pick)
	}

	return picks, nil
}

func (r *PickRepository) GetLast(draftID int) (*models.Pick, error) {
	query := `SELECT * FROM picks WHERE draft_id = ? ORDER BY overall_pick DESC LIMIT 1`
	pick := &models.Pick{}
	err := r.db.QueryRow(query, draftID).Scan(
		&pick.ID, &pick.DraftID, &pick.TeamID, &pick.PlayerID, &pick.Round,
		&pick.OverallPick, &pick.IsTraded, &pick.ADPRank, &pick.PickedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get last pick: %w", err)
	}
	return pick, nil
}

func (r *PickRepository) Delete(id int) error {
	query := `DELETE FROM picks WHERE id = ?`
	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete pick: %w", err)
	}
	return nil
}

func (r *PickRepository) CountByDraft(draftID int) (int, error) {
	query := `SELECT COUNT(*) FROM picks WHERE draft_id = ?`
	var count int
	err := r.db.QueryRow(query, draftID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count picks: %w", err)
	}
	return count, nil
}

func (r *PickRepository) GetDraftedPlayerIDs(draftID int) ([]int, error) {
	query := `SELECT player_id FROM picks WHERE draft_id = ?`
	rows, err := r.db.Query(query, draftID)
	if err != nil {
		return nil, fmt.Errorf("failed to get drafted player IDs: %w", err)
	}
	defer rows.Close()

	var playerIDs []int
	for rows.Next() {
		var playerID int
		if err := rows.Scan(&playerID); err != nil {
			return nil, fmt.Errorf("failed to scan player ID: %w", err)
		}
		playerIDs = append(playerIDs, playerID)
	}

	return playerIDs, nil
}

func (r *PickRepository) Update(pick *models.Pick) error {
	query := `
		UPDATE picks 
		SET team_id = ?, is_traded = ?, adp_rank = ?
		WHERE id = ?
	`
	_, err := r.db.Exec(query, pick.TeamID, pick.IsTraded, pick.ADPRank, pick.ID)
	if err != nil {
		return fmt.Errorf("failed to update pick: %w", err)
	}
	return nil
}

