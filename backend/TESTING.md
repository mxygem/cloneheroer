# Testing the Clone Hero Score Tracker

## Quick Start

1. **Set up your environment variables:**

   Create a `.env` file in the `backend/` directory (or use the provided `.env.example`):

   ```bash
   cd backend
   cp .env.example .env
   # Edit .env with your settings
   ```

   Minimum required variables:
   ```bash
   export WATCH_DIR=/path/to/your/screenshots
   export DATABASE_URL="postgres://user:password@localhost:5432/cloneheroer?sslmode=disable"
   ```

2. **Set up PostgreSQL database:**

   ```bash
   # Create database
   createdb cloneheroer
   
   # Or using psql:
   psql -U postgres -c "CREATE DATABASE cloneheroer;"
   ```

3. **Run the service:**

   Option A: Using the run script
   ```bash
   ./run.sh
   ```

   Option B: Manually
   ```bash
   export WATCH_DIR=/path/to/screenshots
   export DATABASE_URL="postgres://user:pass@localhost/cloneheroer?sslmode=disable"
   go run cmd/server/main.go
   ```

## Testing with Sample Images

The repository includes sample screenshots in `testdata/screenshots/`. To test:

1. **Copy a sample image to your watch directory:**
   ```bash
   cp ../testdata/screenshots/clonehero-Adolescents-20251019101128.png "$WATCH_DIR/"
   ```

2. **Watch the logs** - You should see:
   ```
   processing new file: /path/to/screenshot.png
   parsing image: /path/to/screenshot.png
   creating score for: Adolescents - [Song Name]
   successfully created score with ID: 1
   ```

3. **Check the database:**
   ```bash
   psql -U postgres -d cloneheroer -c "SELECT * FROM scores ORDER BY created_at DESC LIMIT 5;"
   ```

4. **Query the API:**
   ```bash
   curl http://localhost:3000/scores?limit=5
   ```

## Troubleshooting

### "failed to connect to database"
- Verify PostgreSQL is running: `pg_isready`
- Check your `DATABASE_URL` connection string
- Ensure the database exists: `psql -l | grep cloneheroer`

### "failed to create parser: Tesseract not found"
- Install Tesseract OCR (see README.md)
- Verify installation: `tesseract --version`

### "no timestamp found in filename"
- This is a warning, not an error
- The service will use the file's modification time instead
- Ensure screenshots follow the naming pattern: `*YYYYMMDDHHmmss*.png`

### Parser not extracting data correctly
- The parser uses heuristic-based region extraction
- You may need to adjust coordinates in `internal/parser/parser.go` based on your screenshot format
- Check the extracted text by adding debug logging

### Images not being processed
- Verify `WATCH_DIR` exists and is readable
- Check file permissions
- Ensure images are PNG, JPEG, or WebP format
- Check logs for any error messages

## Manual Testing Endpoints

Once the service is running, test the API:

```bash
# Health check
curl http://localhost:3000/health

# List scores
curl http://localhost:3000/scores?limit=10

# List scores with pagination
curl http://localhost:3000/scores?limit=5&offset=0
```

## Expected Database Schema

After running migrations, you should have:

- `artists` - Artist names
- `songs` - Song information linked to artists
- `scores` - Score records with player data as JSONB
- `players` - Individual player performance data

Check the schema:
```bash
psql -U postgres -d cloneheroer -c "\dt"
```


