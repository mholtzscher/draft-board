package repository

import (
	"database/sql"
	"fmt"

	"github.com/vibes/draft-board/internal/models"
)

type QueueRepository struct {
	db *sql.DB
}

func NewQueueRepository(db *sql.DB) *QueueRepository {
	return &QueueRepository{db: db}
}

func (r *QueueRepository) Create(queueItem *models.QueueItem) error {
	query := `INSERT INTO draft_queue (draft_id, team_id, player_id, queue_order) VALUES (?, ?, ?, ?)`
	result, err := r.db.Exec(query, queueItem.DraftID, queueItem.TeamID, queueItem.PlayerID, queueItem.QueueOrder)
	if err != nil {
		return fmt.Errorf("failed to create queue item: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	queueItem.ID = int(id)
	return nil
}

func (r *QueueRepository) GetByTeam(draftID, teamID int) ([]models.QueueItem, error) {
	query := `SELECT * FROM draft_queue WHERE draft_id = ? AND team_id = ? ORDER BY queue_order`
	rows, err := r.db.Query(query, draftID, teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to get queue items: %w", err)
	}
	defer rows.Close()

	var items []models.QueueItem
	for rows.Next() {
		var item models.QueueItem
		err := rows.Scan(&item.ID, &item.DraftID, &item.TeamID, &item.PlayerID, &item.QueueOrder, &item.AddedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan queue item: %w", err)
		}
		items = append(items, item)
	}

	return items, nil
}

func (r *QueueRepository) Delete(id int) error {
	query := `DELETE FROM draft_queue WHERE id = ?`
	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete queue item: %w", err)
	}
	return nil
}

func (r *QueueRepository) GetMaxOrder(draftID, teamID int) (int, error) {
	query := `SELECT COALESCE(MAX(queue_order), 0) FROM draft_queue WHERE draft_id = ? AND team_id = ?`
	var maxOrder int
	err := r.db.QueryRow(query, draftID, teamID).Scan(&maxOrder)
	if err != nil {
		return 0, fmt.Errorf("failed to get max order: %w", err)
	}
	return maxOrder, nil
}

func (r *QueueRepository) Reorder(draftID, teamID int, playerIDs []int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	for i, playerID := range playerIDs {
		query := `UPDATE draft_queue SET queue_order = ? WHERE draft_id = ? AND team_id = ? AND player_id = ?`
		_, err := tx.Exec(query, i+1, draftID, teamID, playerID)
		if err != nil {
			return fmt.Errorf("failed to reorder queue: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

