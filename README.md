# cloneheroer

A service for various clone hero related tasks.

## Initial Prompt - Score Tracking

Please help me create a service in golang that tracks scores from screenshots of scores achieved after playing a song in a game called Clone Hero.

The service should:

1. Watch the given directory for new images (directory is provided by flag or config file)
2. When a new image appears, parse the following information from the image:
   - Artist name - first text in top left of image
   - Song name - second text in top left of image
   - Charter - third text in top left of image
   - Total Score - number at top of image in center
   - Stars Achieved - Below total score in center
   - Players and their data - main data in image, up to 4 players
       - For each player:
          - Player name
          - Instrument - guitar, rhythm, co-op, bass, pro drums, drums, keys, also may say "no part"; will need to be parsed from image of instrument icons that can be found in `img/instrum-icons.png`
          - Difficulty
          - Score
          - Best Streak
          - Accuracy
          - Misses
          - Rank - if played in versus mode
   - Date and time - parsed from image name. image name is in the format of `20251212052231` which is yyyyMMddHHmmss
3. Store the score in a postgres database
   - The database schema should be created with the following tables:
      - scores (id, artist, song_id(foreign key to songs table), charter, total_score, stars_achieved, players, created_at)
      - songs (id, name, artist_id(foreign key to artists table), charters, created_at)
      - artists (id, name, created_at)
      - players (id, name, instrument, difficulty, score, combo, accuracy, misses, rank, created_at)
4. Provide a REST API to query the scores
5. The service should be able to run on any machine that has a modern Linux distribution installed.

## Implementation Status

The service has been implemented with the following components:

### Completed Features

1. **Database Schema** - PostgreSQL migrations for artists, songs, scores, and players tables
2. **Database Repository** - CRUD operations for all entities, including score creation
3. **REST API** - Echo-based HTTP server with endpoints for:
   - `GET /scores` - List scores with pagination
   - `PATCH /artists/:id` - Update artist
   - `PATCH /songs/:id` - Update song
   - `PATCH /scores/:id` - Update score
   - `PATCH /players/:id` - Update player
   - `GET /health` - Health check
4. **File Watcher** - Monitors `WATCH_DIR` for new image files (PNG, JPEG, WebP)
5. **Image Parser** - OCR-based extraction of score data from screenshots using Tesseract
6. **Integration** - File watcher → Image parser → Database insertion pipeline

### Setup Instructions

#### Prerequisites

1. **Go 1.22+** - Install from https://go.dev/dl/
2. **PostgreSQL** - Database server running and accessible
3. **Tesseract OCR** - Required for image parsing
   ```bash
   # Ubuntu/Debian
   sudo apt-get install tesseract-ocr
   
   # macOS
   brew install tesseract
   ```

#### Configuration

The service uses environment variables for configuration:

- `WATCH_DIR` (required) - Directory to watch for new screenshot images
- `DATABASE_URL` (required) - PostgreSQL connection string (e.g., `postgres://user:pass@localhost/dbname?sslmode=disable`)
- `PORT` (optional, default: 3000) - HTTP server port
- `MIGRATE_ON_START` (optional, default: true) - Run database migrations on startup
- `PROCESSED_DIR` (optional) - Directory to move successfully processed images
- `FAILED_DIR` (optional) - Directory to move images that failed to process

#### Running the Service

1. Install dependencies:
   ```bash
   cd backend
   go mod download
   ```

2. Set environment variables:
   ```bash
   export WATCH_DIR=/path/to/screenshots
   export DATABASE_URL="postgres://user:password@localhost/cloneheroer?sslmode=disable"
   ```

3. Run the service:
   ```bash
   go run cmd/server/main.go
   ```

The service will:
- Run database migrations on startup
- Start watching the `WATCH_DIR` for new images
- Process existing images in the directory
- Start the HTTP API server
- Process new images as they appear

### Notes

- **Instrument Detection**: Currently uses OCR text parsing. For more accurate instrument detection, template matching with `img/instrum-icons.png` should be implemented (see TODO in `parser.go`).
- **Image Processing**: The parser uses heuristic-based region extraction. You may need to adjust the region coordinates in `parser.go` based on your screenshot format.
- **Error Handling**: Failed images are moved to `FAILED_DIR` if configured, otherwise they remain in `WATCH_DIR`.

### Frontend

A basic Next.js frontend is included in the `frontend/` directory. To run it:

```bash
cd frontend
npm install
npm run dev
```

The frontend displays scores from the API at `http://localhost:3000` (default).

Thanks!
