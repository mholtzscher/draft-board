package repository

import (
	"database/sql"
	"fmt"

	"github.com/vibes/draft-board/internal/models"
)

type AuditRepository struct {
	db *sql.DB
}

func NewAuditRepository(db *sql.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

func (r *AuditRepository) Log(draftID int, actionType string, entityID *int, details string) error {
	query := `INSERT INTO audit_log (draft_id, action_type, entity_id, details) VALUES (?, ?, ?, ?)`
	_, err := r.db.Exec(query, draftID, actionType, entityID, details)
	if err != nil {
		return fmt.Errorf("failed to log audit: %w", err)
	}
	return nil
}

func (r *AuditRepository) GetByDraft(draftID int) ([]models.AuditLog, error) {
	query := `SELECT * FROM audit_log WHERE draft_id = ? ORDER BY performed_at DESC`
	rows, err := r.db.Query(query, draftID)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit logs: %w", err)
	}
	defer rows.Close()

	var logs []models.AuditLog
	for rows.Next() {
		var log models.AuditLog
		var entityID sql.NullInt64
		err := rows.Scan(&log.ID, &log.DraftID, &log.ActionType, &entityID, &log.Details, &log.PerformedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}
		if entityID.Valid {
			id := int(entityID.Int64)
			log.EntityID = &id
		}
		logs = append(logs, log)
	}

	return logs, nil
}

