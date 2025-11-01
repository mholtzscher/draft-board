package repository

import (
	"database/sql"
	"fmt"

	"github.com/vibes/draft-board/internal/models"
)

type DraftRepository struct {
	db *sql.DB
}

func NewDraftRepository(db *sql.DB) *DraftRepository {
	return &DraftRepository{db: db}
}

func (r *DraftRepository) Create(draft *models.Draft) error {
	query := `
		INSERT INTO drafts (name, num_teams, scoring_format, draft_type, qb_setting, snake_draft, status, max_rounds, commissioner_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.Exec(query, draft.Name, draft.NumTeams, draft.ScoringFormat, draft.DraftType,
		draft.QBSetting, draft.SnakeDraft, draft.Status, draft.MaxRounds, draft.CommissionerID)
	if err != nil {
		return fmt.Errorf("failed to create draft: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	draft.ID = int(id)
	return nil
}

func (r *DraftRepository) GetByID(id int) (*models.Draft, error) {
	query := `SELECT * FROM drafts WHERE id = ?`
	draft := &models.Draft{}
	err := r.db.QueryRow(query, id).Scan(
		&draft.ID, &draft.Name, &draft.NumTeams, &draft.ScoringFormat,
		&draft.DraftType, &draft.QBSetting, &draft.SnakeDraft, &draft.Status,
		&draft.MaxRounds, &draft.CommissionerID, &draft.CreatedAt, &draft.Completed,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("draft not found")
		}
		return nil, fmt.Errorf("failed to get draft: %w", err)
	}
	return draft, nil
}

func (r *DraftRepository) Update(draft *models.Draft) error {
	query := `
		UPDATE drafts 
		SET name = ?, num_teams = ?, scoring_format = ?, draft_type = ?, 
		    qb_setting = ?, snake_draft = ?, status = ?, max_rounds = ?, 
		    commissioner_id = ?, completed = ?
		WHERE id = ?
	`
	_, err := r.db.Exec(query, draft.Name, draft.NumTeams, draft.ScoringFormat, draft.DraftType,
		draft.QBSetting, draft.SnakeDraft, draft.Status, draft.MaxRounds,
		draft.CommissionerID, draft.Completed, draft.ID)
	if err != nil {
		return fmt.Errorf("failed to update draft: %w", err)
	}
	return nil
}

func (r *DraftRepository) Delete(id int) error {
	query := `DELETE FROM drafts WHERE id = ?`
	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete draft: %w", err)
	}
	return nil
}

func (r *DraftRepository) List() ([]*models.Draft, error) {
	query := `SELECT * FROM drafts ORDER BY created_at DESC`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list drafts: %w", err)
	}
	defer rows.Close()

	var drafts []*models.Draft
	for rows.Next() {
		draft := &models.Draft{}
		err := rows.Scan(
			&draft.ID, &draft.Name, &draft.NumTeams, &draft.ScoringFormat,
			&draft.DraftType, &draft.QBSetting, &draft.SnakeDraft, &draft.Status,
			&draft.MaxRounds, &draft.CommissionerID, &draft.CreatedAt, &draft.Completed,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan draft: %w", err)
		}
		drafts = append(drafts, draft)
	}

	return drafts, nil
}

