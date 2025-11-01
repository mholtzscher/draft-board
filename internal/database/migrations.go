package database

import (
	"database/sql"
	"fmt"
)

func RunMigrations(db *sql.DB) error {
	migrations := []string{
		createPlayersTable,
		createDraftsTable,
		createTeamsTable,
		createPicksTable,
		createPositionSettingsTable,
		createDraftQueueTable,
		createAuditLogTable,
		createIndexes,
	}

	for i, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("migration %d failed: %w", i+1, err)
		}
	}

	return nil
}

const createPlayersTable = `
CREATE TABLE IF NOT EXISTS players (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    team TEXT NOT NULL,
    position TEXT NOT NULL CHECK(position IN ('QB', 'RB', 'WR', 'TE', 'K', 'D/ST', 'DL', 'LB', 'DB')),
    bye_week INTEGER CHECK(bye_week BETWEEN 1 AND 18),
    dynasty_rank INTEGER,
    sf_rank INTEGER,
    std_rank INTEGER,
    half_ppr_rank INTEGER,
    ppr_rank INTEGER,
    is_custom BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
`

const createDraftsTable = `
CREATE TABLE IF NOT EXISTS drafts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    num_teams INTEGER NOT NULL CHECK(num_teams BETWEEN 2 AND 14),
    scoring_format TEXT NOT NULL CHECK(scoring_format IN ('Standard', 'Half-PPR', 'PPR')),
    draft_type TEXT NOT NULL CHECK(draft_type IN ('Redraft', 'Dynasty')),
    qb_setting TEXT DEFAULT '1QB',
    snake_draft BOOLEAN DEFAULT TRUE,
    status TEXT DEFAULT 'setup' CHECK(status IN ('setup', 'active', 'paused', 'completed')),
    max_rounds INTEGER DEFAULT 16,
    commissioner_id TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed BOOLEAN DEFAULT FALSE
);
`

const createTeamsTable = `
CREATE TABLE IF NOT EXISTS teams (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    draft_id INTEGER NOT NULL,
    team_name TEXT NOT NULL,
    owner_name TEXT,
    draft_position INTEGER NOT NULL,
    FOREIGN KEY (draft_id) REFERENCES drafts(id) ON DELETE CASCADE,
    UNIQUE(draft_id, draft_position),
    UNIQUE(draft_id, team_name)
);
`

const createPicksTable = `
CREATE TABLE IF NOT EXISTS picks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    draft_id INTEGER NOT NULL,
    team_id INTEGER NOT NULL,
    player_id INTEGER NOT NULL,
    round INTEGER NOT NULL,
    overall_pick INTEGER NOT NULL,
    is_traded BOOLEAN DEFAULT FALSE,
    adp_rank INTEGER,
    picked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (draft_id) REFERENCES drafts(id) ON DELETE CASCADE,
    FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE,
    FOREIGN KEY (player_id) REFERENCES players(id),
    UNIQUE(draft_id, player_id)
);
`

const createPositionSettingsTable = `
CREATE TABLE IF NOT EXISTS position_settings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    draft_id INTEGER NOT NULL,
    position TEXT NOT NULL,
    enabled BOOLEAN DEFAULT TRUE,
    FOREIGN KEY (draft_id) REFERENCES drafts(id) ON DELETE CASCADE,
    UNIQUE(draft_id, position)
);
`

const createDraftQueueTable = `
CREATE TABLE IF NOT EXISTS draft_queue (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    draft_id INTEGER NOT NULL,
    team_id INTEGER NOT NULL,
    player_id INTEGER NOT NULL,
    queue_order INTEGER NOT NULL,
    added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (draft_id) REFERENCES drafts(id) ON DELETE CASCADE,
    FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE,
    FOREIGN KEY (player_id) REFERENCES players(id),
    UNIQUE(draft_id, team_id, player_id)
);
`

const createAuditLogTable = `
CREATE TABLE IF NOT EXISTS audit_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    draft_id INTEGER NOT NULL,
    action_type TEXT NOT NULL CHECK(action_type IN ('pick', 'undo', 'trade', 'pause', 'resume', 'complete', 'start')),
    entity_id INTEGER,
    details TEXT,
    performed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (draft_id) REFERENCES drafts(id) ON DELETE CASCADE
);
`

const createIndexes = `
CREATE INDEX IF NOT EXISTS idx_players_position ON players(position);
CREATE INDEX IF NOT EXISTS idx_players_team ON players(team);
CREATE INDEX IF NOT EXISTS idx_players_name ON players(name);
CREATE INDEX IF NOT EXISTS idx_players_rank ON players(ppr_rank, half_ppr_rank, std_rank, dynasty_rank);
CREATE INDEX IF NOT EXISTS idx_picks_draft ON picks(draft_id);
CREATE INDEX IF NOT EXISTS idx_picks_team ON picks(team_id);
CREATE INDEX IF NOT EXISTS idx_picks_player_draft ON picks(player_id, draft_id);
CREATE INDEX IF NOT EXISTS idx_teams_draft ON teams(draft_id);
CREATE INDEX IF NOT EXISTS idx_draft_queue_draft ON draft_queue(draft_id);
CREATE INDEX IF NOT EXISTS idx_audit_log_draft ON audit_log(draft_id);
CREATE INDEX IF NOT EXISTS idx_drafts_status ON drafts(status);
`

