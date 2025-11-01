# Docker Deployment Guide

## Quick Start

### Using Docker Compose (Recommended)

1. **Build and start the application:**
   ```bash
   docker-compose up -d
   ```

2. **View logs:**
   ```bash
   docker-compose logs -f
   ```

3. **Stop the application:**
   ```bash
   docker-compose down
   ```

4. **Rebuild after code changes:**
   ```bash
   docker-compose up -d --build
   ```

### Using Docker directly

1. **Build the image:**
   ```bash
   docker build -t draft-board .
   ```

2. **Run the container:**
   ```bash
   docker run -d \
     --name draft-board \
     -p 8080:8080 \
     -v draft-board-data:/app/data \
     draft-board
   ```

## Configuration

### Environment Variables

- `PORT` - Server port (default: 8080)
- `DB_PATH` - Database file path (default: /app/data/draft-board.db)

### Volumes

- `draft-board-data` - Persistent storage for the SQLite database
- `./data:/app/data/imports:ro` - Read-only mount for CSV imports (optional)

## Seeding Player Data

### Option 1: Copy CSV into container

```bash
# Copy CSV file into container
docker cp data/sample-players.csv draft-board:/tmp/players.csv

# Run seed command (using the seed binary included in the image)
docker exec draft-board /app/seed -file /tmp/players.csv
```

### Option 2: Mount data directory

If you mount the data directory (as shown in docker-compose.yml):

```bash
# Place CSV file in ./data directory
cp your-players.csv data/

# Run seed command
docker exec draft-board /app/seed -file /app/data/imports/your-players.csv
```

### Option 3: Seed before starting container

```bash
# Build locally first
go build -o seed ./cmd/seed/main.go

# Seed your database file
./seed -file your-players.csv

# Copy seeded database to volume
docker volume create draft-board-data
docker run --rm -v draft-board-data:/data -v $(pwd):/backup alpine cp /backup/draft-board.db /data/draft-board.db
```

## Production Deployment

### Recommended Setup

1. **Use a reverse proxy** (nginx, Traefik, etc.) for SSL termination
2. **Set up regular backups** of the `draft-board-data` volume
3. **Configure resource limits** in docker-compose.yml:
   ```yaml
   deploy:
     resources:
       limits:
         cpus: '1'
         memory: 512M
   ```

### Example nginx configuration

```nginx
server {
    listen 80;
    server_name draft-board.example.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## Troubleshooting

### Check container logs
```bash
docker-compose logs draft-board
```

### Access container shell
```bash
docker exec -it draft-board sh
```

### Check database
```bash
docker exec -it draft-board sqlite3 /app/data/draft-board.db "SELECT COUNT(*) FROM players;"
```

### Backup database
```bash
docker exec draft-board sqlite3 /app/data/draft-board.db ".backup /app/data/backup.db"
docker cp draft-board:/app/data/backup.db ./backup.db
```

### Restore database
```bash
docker cp backup.db draft-board:/app/data/restup.db
docker exec draft-board sqlite3 /app/data/draft-board.db ".restore /app/data/restup.db"
```

