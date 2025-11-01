package repository

import (
	"database/sql"
	"fmt"

	"github.com/vibes/draft-board/internal/models"
)

type TeamRepository struct {
	db *sql.DB
}

func NewTeamRepository(db *sql.DB) *TeamRepository {
	return &TeamRepository{db: db}
}

func (r *TeamRepository) Create(team *models.Team) error {
	query := `INSERT INTO teams (draft_id, team_name, owner_name, draft_position) VALUES (?, ?, ?, ?)`
	result, err := r.db.Exec(query, team.DraftID, team.TeamName, team.OwnerName, team.DraftPosition)
	if err != nil {
		return fmt.Errorf("failed to create team: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	team.ID = int(id)
	return nil
}

func (r *TeamRepository) GetByID(id int) (*models.Team, error) {
	query := `SELECT * FROM teams WHERE id = ?`
	team := &models.Team{}
	err := r.db.QueryRow(query, id).Scan(&team.ID, &team.DraftID, &team.TeamName, &team.OwnerName, &team.DraftPosition)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("team not found")
		}
		return nil, fmt.Errorf("failed to get team: %w", err)
	}
	return team, nil
}

func (r *TeamRepository) GetByDraft(draftID int) ([]models.Team, error) {
	query := `SELECT * FROM teams WHERE draft_id = ? ORDER BY draft_position`
	rows, err := r.db.Query(query, draftID)
	if err != nil {
		return nil, fmt.Errorf("failed to get teams: %w", err)
	}
	defer rows.Close()

	var teams []models.Team
	for rows.Next() {
		var team models.Team
		err := rows.Scan(&team.ID, &team.DraftID, &team.TeamName, &team.OwnerName, &team.DraftPosition)
		if err != nil {
			return nil, fmt.Errorf("failed to scan team: %w", err)
		}
		teams = append(teams, team)
	}

	return teams, nil
}

func (r *TeamRepository) Update(team *models.Team) error {
	query := `UPDATE teams SET team_name = ?, owner_name = ?, draft_position = ? WHERE id = ?`
	_, err := r.db.Exec(query, team.TeamName, team.OwnerName, team.DraftPosition, team.ID)
	if err != nil {
		return fmt.Errorf("failed to update team: %w", err)
	}
	return nil
}

func (r *TeamRepository) Delete(id int) error {
	query := `DELETE FROM teams WHERE id = ?`
	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete team: %w", err)
	}
	return nil
}

func (r *TeamRepository) CountByDraft(draftID int) (int, error) {
	query := `SELECT COUNT(*) FROM teams WHERE draft_id = ?`
	var count int
	err := r.db.QueryRow(query, draftID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count teams: %w", err)
	}
	return count, nil
}

