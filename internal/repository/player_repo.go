package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/vibes/draft-board/internal/models"
)

type PlayerRepository struct {
	db *sql.DB
}

func NewPlayerRepository(db *sql.DB) *PlayerRepository {
	return &PlayerRepository{db: db}
}

func (r *PlayerRepository) GetByID(id int) (*models.Player, error) {
	query := `SELECT * FROM players WHERE id = ?`
	player := &models.Player{}
	err := r.db.QueryRow(query, id).Scan(
		&player.ID, &player.Name, &player.Team, &player.Position, &player.ByeWeek,
		&player.DynastyRank, &player.SFRank, &player.StdRank, &player.HalfPPRRank,
		&player.PPRRank, &player.IsCustom, &player.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("player not found")
		}
		return nil, fmt.Errorf("failed to get player: %w", err)
	}
	return player, nil
}

func (r *PlayerRepository) GetAvailable(draftID int, filters PlayerFilters) ([]*models.Player, error) {
	query := `
		SELECT p.* FROM players p
		WHERE p.id NOT IN (SELECT player_id FROM picks WHERE draft_id = ?)
	`
	args := []interface{}{draftID}

	if len(filters.Positions) > 0 {
		placeholders := make([]string, len(filters.Positions))
		for i, pos := range filters.Positions {
			placeholders[i] = "?"
			args = append(args, pos)
		}
		query += fmt.Sprintf(" AND p.position IN (%s)", strings.Join(placeholders, ","))
	}

	if filters.Search != "" {
		query += " AND (p.name LIKE ? OR p.team LIKE ?)"
		searchPattern := "%" + filters.Search + "%"
		args = append(args, searchPattern, searchPattern)
	}

	// Order by rank based on draft type and scoring format
	// SQLite doesn't support NULLS LAST, so we use COALESCE to handle NULLs
	if filters.DraftType == "Dynasty" {
		query += " ORDER BY COALESCE(p.dynasty_rank, 9999) ASC"
	} else {
		switch filters.ScoringFormat {
		case "PPR":
			query += " ORDER BY COALESCE(p.ppr_rank, 9999) ASC"
		case "Half-PPR":
			query += " ORDER BY COALESCE(p.half_ppr_rank, 9999) ASC"
		default:
			query += " ORDER BY COALESCE(p.std_rank, 9999) ASC"
		}
	}

	if filters.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", filters.Limit)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get available players: %w", err)
	}
	defer rows.Close()

	var players []*models.Player
	for rows.Next() {
		player := &models.Player{}
		err := rows.Scan(
			&player.ID, &player.Name, &player.Team, &player.Position, &player.ByeWeek,
			&player.DynastyRank, &player.SFRank, &player.StdRank, &player.HalfPPRRank,
			&player.PPRRank, &player.IsCustom, &player.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan player: %w", err)
		}
		players = append(players, player)
	}

	return players, nil
}

func (r *PlayerRepository) Create(player *models.Player) error {
	query := `
		INSERT INTO players (name, team, position, bye_week, dynasty_rank, sf_rank, std_rank, half_ppr_rank, ppr_rank, is_custom)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.Exec(query, player.Name, player.Team, player.Position, player.ByeWeek,
		player.DynastyRank, player.SFRank, player.StdRank, player.HalfPPRRank, player.PPRRank, player.IsCustom)
	if err != nil {
		return fmt.Errorf("failed to create player: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	player.ID = int(id)
	return nil
}

type PlayerFilters struct {
	Positions    []string
	Search       string
	DraftType    string
	ScoringFormat string
	Limit        int
}

