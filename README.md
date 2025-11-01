# Fantasy Draft Board

A modern web application for managing offline fantasy football drafts.

## Docker Deployment

### Quick Start with Docker Compose

```bash
# Build and start
docker-compose up -d

# View logs
docker-compose logs -f

# Stop
docker-compose down
```

The application will be available at http://localhost:8080

### Seed Player Data in Docker

```bash
# Copy CSV file and seed
docker cp data/sample-players.csv draft-board:/tmp/players.csv
docker exec draft-board /app/seed -file /tmp/players.csv
```

See `docs/DOCKER.md` for detailed Docker deployment instructions.

## Quick Start

1. Build the server:
```bash
go build -o draft-board ./cmd/server/main.go
```

2. Seed player data (optional but recommended):
```bash
# Use sample data
just seed-sample

# Or import your own CSV
just seed path/to/your/players.csv
```

3. Run the server:
```bash
./draft-board
```

4. Open http://localhost:8080 in your browser

## Seeding Player Data

The application needs player data to function. You can import players from a CSV file.

### Quick Start with Sample Data

```bash
just seed-sample
```

This will import 20 sample players for testing.

### Import Your Own CSV

1. Create a CSV file with this format:
```csv
name,team,position,bye_week,std_rank,half_ppr_rank,ppr_rank,dynasty_rank
Ja'Marr Chase,CIN,WR,10,1,1,1,1
Patrick Mahomes II,KCC,QB,7,20,20,20,20
```

2. Import it:
```bash
just seed your-players.csv
```

See `docs/SEEDING.md` for detailed CSV format and options.

### CSV Format

Required columns: `name`, `team`, `position`
Optional columns: `bye_week`, `dynasty_rank`, `sf_rank`, `std_rank`, `half_ppr_rank`, `ppr_rank`

## Features

- Create and manage drafts
- Support for 8, 10, 12, and 14 team leagues
- Multiple scoring formats (Standard, Half-PPR, PPR)
- Dynasty and Redraft rankings
- Snake draft pattern automation
- Real-time draft board updates
- Player queue/watchlist
- Export functionality
- Comprehensive statistics
- Beautiful Tokyo Night dark theme

## Database

The application uses SQLite and will create a database file (`draft-board.db`) automatically on first run.

## Environment Variables

- `PORT` - Server port (default: 8080)
- `DB_PATH` - Database file path (default: ./draft-board.db)

## Common Commands

```bash
just run          # Start the server
just build        # Build the binary
just seed-sample  # Import sample player data
just seed file.csv # Import players from CSV
just db-shell     # Open SQLite shell
just db-reset     # Reset database (delete and recreate)
```

See `justfile` for all available commands.

