# Player Data Seeding Guide

## CSV Format

Create a CSV file with the following columns (header row required):

```csv
name,team,position,bye_week,dynasty_rank,sf_rank,std_rank,half_ppr_rank,ppr_rank
Ja'Marr Chase,CIN,WR,10,,,1,1,1
Patrick Mahomes II,KCC,QB,7,,,20,20,20
Bijan Robinson,ATL,RB,5,2,2,2,2,2
```

### Column Descriptions

- **name** (required): Player full name
- **team** (required): NFL team abbreviation (e.g., CIN, KCC, SFO)
- **position** (required): One of: QB, RB, WR, TE, K, D/ST, DL, LB, DB
- **bye_week** (optional): Bye week (1-18)
- **dynasty_rank** (optional): Dynasty league ranking
- **sf_rank** (optional): Superflex ranking
- **std_rank** (optional): Standard scoring ranking
- **half_ppr_rank** (optional): Half-PPR ranking
- **ppr_rank** (optional): PPR ranking

### Ranking Notes

- Lower numbers = higher rank (1 is best)
- Empty values are allowed for optional fields
- Rankings determine player order in available players list

## Importing Player Data

### Method 1: Using the Seed Command

```bash
# Build the seed tool
go build -o seed ./cmd/seed/main.go

# Import from CSV
./seed -file players.csv

# Or run directly
go run ./cmd/seed/main.go -file players.csv
```

### Method 2: Using Justfile

Add this to your justfile (if you want):

```bash
# Seed player data from CSV
seed file:
    @go run ./cmd/seed/main.go -file {{file}}
```

Then run:
```bash
just seed players.csv
```

## Sample CSV File

You can create a CSV file with your player data. Here's a minimal example:

```csv
name,team,position,bye_week,std_rank,half_ppr_rank,ppr_rank
Ja'Marr Chase,CIN,WR,10,1,1,1
Saquon Barkley,PHI,RB,9,2,2,2
Justin Jefferson,MIN,WR,6,3,3,3
```

## Getting Player Data

### Option 1: Export from Excel

If you have the original Excel file:
1. Open the `HELP_PlayerDB` sheet
2. Export as CSV
3. Ensure columns match the expected format
4. Run the seed command

### Option 2: Manual Entry

Create a CSV file manually with your top players. You can add more players later using the custom player feature in the web UI.

### Option 3: Import from FantasyPros

1. Get FantasyPros consensus ADP data
2. Format it as CSV with the required columns
3. Import using the seed command

## Verifying Import

After seeding, you can verify players were imported:

```bash
# Using SQLite shell
sqlite3 draft-board.db "SELECT COUNT(*) FROM players;"
sqlite3 draft-board.db "SELECT name, team, position FROM players LIMIT 10;"
```

## Adding More Players Later

You can add individual players using the web UI:
1. Go to any draft
2. Use the "Add Custom Player" feature (via API endpoint `/players/custom`)

Or import another CSV file - the seed script will add new players without duplicates (based on name matching).

