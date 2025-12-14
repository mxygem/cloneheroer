package db

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repo provides database operations using pgxpool.
type Repo struct {
	pool *pgxpool.Pool
}

// NewRepo creates a Repo wrapping the provided pgx pool.
func NewRepo(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool}
}

// Score represents a stored score row.
type Score struct {
	ID           int64                  `json:"id"`
	SongID       *int64                 `json:"song_id,omitempty"`
	Artist       string                 `json:"artist"`
	Charter      *string                `json:"charter,omitempty"`
	TotalScore   *int64                 `json:"total_score,omitempty"`
	StarsAchieved *int                  `json:"stars_achieved,omitempty"`
	Players      map[string]any         `json:"players,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
}

// ListScores returns paginated scores.
func (r *Repo) ListScores(ctx context.Context, limit, offset int32) ([]Score, error) {
	rows, err := r.pool.Query(ctx, `
        SELECT id, song_id, artist, charter, total_score, stars_achieved, players, created_at
        FROM scores
        ORDER BY created_at DESC
        LIMIT $1 OFFSET $2
    `, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Score
	for rows.Next() {
		var s Score
		var playersData map[string]any
		var songID *int64
		var charter *string
		var totalScore *int64
		var stars *int
		if err := rows.Scan(
			&s.ID,
			&songID,
			&s.Artist,
			&charter,
			&totalScore,
			&stars,
			&playersData,
			&s.CreatedAt,
		); err != nil {
			return nil, err
		}
		s.SongID = songID
		s.Charter = charter
		s.TotalScore = totalScore
		s.StarsAchieved = stars
		s.Players = playersData
		out = append(out, s)
	}
	return out, nil
}

// Artist represents an artist row.
type Artist struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// ListArtists returns paginated artists.
func (r *Repo) ListArtists(ctx context.Context, limit, offset int32) ([]Artist, error) {
	rows, err := r.pool.Query(ctx, `
        SELECT id, name, created_at
        FROM artists
        ORDER BY name ASC
        LIMIT $1 OFFSET $2
    `, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Artist
	for rows.Next() {
		var a Artist
		if err := rows.Scan(&a.ID, &a.Name, &a.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, nil
}

// Song represents a song row.
type Song struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	ArtistID  *int64    `json:"artist_id,omitempty"`
	Charters  []string  `json:"charters"`
	CreatedAt time.Time `json:"created_at"`
}

// ListSongs returns paginated songs.
func (r *Repo) ListSongs(ctx context.Context, limit, offset int32) ([]Song, error) {
	rows, err := r.pool.Query(ctx, `
        SELECT id, name, artist_id, charters, created_at
        FROM songs
        ORDER BY name ASC
        LIMIT $1 OFFSET $2
    `, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Song
	for rows.Next() {
		var s Song
		var artistID *int64
		if err := rows.Scan(&s.ID, &s.Name, &artistID, &s.Charters, &s.CreatedAt); err != nil {
			return nil, err
		}
		s.ArtistID = artistID
		out = append(out, s)
	}
	return out, nil
}

// UpdateArtist partially updates an artist.
func (r *Repo) UpdateArtist(ctx context.Context, id int64, name *string) error {
	if name == nil {
		return errors.New("no fields to update")
	}
	_, err := r.pool.Exec(ctx, `UPDATE artists SET name = $1 WHERE id = $2`, name, id)
	return err
}

// UpdateSong partially updates a song. Charters replaces the slice if provided.
func (r *Repo) UpdateSong(ctx context.Context, id int64, name *string, artistID *int64, charters []string) error {
	if name == nil && artistID == nil && charters == nil {
		return errors.New("no fields to update")
	}

	if charters != nil {
		_, err := r.pool.Exec(ctx, `UPDATE songs SET charters = $1 WHERE id = $2`, charters, id)
		if err != nil {
			return err
		}
	}
	if name != nil {
		if _, err := r.pool.Exec(ctx, `UPDATE songs SET name = $1 WHERE id = $2`, name, id); err != nil {
			return err
		}
	}
	if artistID != nil {
		if _, err := r.pool.Exec(ctx, `UPDATE songs SET artist_id = $1 WHERE id = $2`, artistID, id); err != nil {
			return err
		}
	}
	return nil
}

// UpdateScore updates score fields.
func (r *Repo) UpdateScore(ctx context.Context, id int64, totalScore *int64, stars *int, charter *string) error {
	if totalScore == nil && stars == nil && charter == nil {
		return errors.New("no fields to update")
	}
	if totalScore != nil {
		if _, err := r.pool.Exec(ctx, `UPDATE scores SET total_score = $1 WHERE id = $2`, totalScore, id); err != nil {
			return err
		}
	}
	if stars != nil {
		if _, err := r.pool.Exec(ctx, `UPDATE scores SET stars_achieved = $1 WHERE id = $2`, stars, id); err != nil {
			return err
		}
	}
	if charter != nil {
		if _, err := r.pool.Exec(ctx, `UPDATE scores SET charter = $1 WHERE id = $2`, charter, id); err != nil {
			return err
		}
	}
	return nil
}

// UpdatePlayer updates player stats for manual corrections.
func (r *Repo) UpdatePlayer(ctx context.Context, id int64, name *string, instrument *string, difficulty *string, score *int64, combo *int, accuracy *float64, misses *int, rank *int) error {
	if name == nil && instrument == nil && difficulty == nil && score == nil && combo == nil && accuracy == nil && misses == nil && rank == nil {
		return errors.New("no fields to update")
	}

	if name != nil {
		if _, err := r.pool.Exec(ctx, `UPDATE players SET name = $1 WHERE id = $2`, name, id); err != nil {
			return err
		}
	}
	if instrument != nil {
		if _, err := r.pool.Exec(ctx, `UPDATE players SET instrument = $1 WHERE id = $2`, instrument, id); err != nil {
			return err
		}
	}
	if difficulty != nil {
		if _, err := r.pool.Exec(ctx, `UPDATE players SET difficulty = $1 WHERE id = $2`, difficulty, id); err != nil {
			return err
		}
	}
	if score != nil {
		if _, err := r.pool.Exec(ctx, `UPDATE players SET score = $1 WHERE id = $2`, score, id); err != nil {
			return err
		}
	}
	if combo != nil {
		if _, err := r.pool.Exec(ctx, `UPDATE players SET combo = $1 WHERE id = $2`, combo, id); err != nil {
			return err
		}
	}
	if accuracy != nil {
		if _, err := r.pool.Exec(ctx, `UPDATE players SET accuracy = $1 WHERE id = $2`, accuracy, id); err != nil {
			return err
		}
	}
	if misses != nil {
		if _, err := r.pool.Exec(ctx, `UPDATE players SET misses = $1 WHERE id = $2`, misses, id); err != nil {
			return err
		}
	}
	if rank != nil {
		if _, err := r.pool.Exec(ctx, `UPDATE players SET rank = $1 WHERE id = $2`, rank, id); err != nil {
			return err
		}
	}
	return nil
}

// Player represents a player in a score.
type Player struct {
	Name       string   `json:"name"`
	Instrument *string  `json:"instrument,omitempty"`
	Difficulty *string  `json:"difficulty,omitempty"`
	Score      *int64   `json:"score,omitempty"`
	Combo      *int     `json:"combo,omitempty"`
	Accuracy   *float64 `json:"accuracy,omitempty"`
	Misses     *int     `json:"misses,omitempty"`
	Rank       *int     `json:"rank,omitempty"`
}

// CreateScoreData holds all data needed to create a score.
type CreateScoreData struct {
	Artist       string
	SongName     string
	Charter      *string
	TotalScore   *int64
	StarsAchieved *int
	Players      []Player
	CreatedAt    time.Time
}

// CreateScore creates a new score with artist, song, and players.
// It handles creating or finding the artist and song, then creates the score and players.
func (r *Repo) CreateScore(ctx context.Context, data CreateScoreData) (int64, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	// Get or create artist
	var artistID int64
	err = tx.QueryRow(ctx, `
		INSERT INTO artists (name) VALUES ($1)
		ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
		RETURNING id
	`, data.Artist).Scan(&artistID)
	if err != nil {
		return 0, err
	}

	// Get or create song
	var songID int64
	var charters []string
	if data.Charter != nil && *data.Charter != "" {
		charters = []string{*data.Charter}
	} else {
		charters = []string{}
	}
	
	err = tx.QueryRow(ctx, `
		INSERT INTO songs (name, artist_id, charters)
		VALUES ($1, $2, $3)
		ON CONFLICT (name, artist_id) DO UPDATE SET name = EXCLUDED.name
		RETURNING id
	`, data.SongName, artistID, charters).Scan(&songID)
	if err != nil {
		return 0, err
	}

	// Update charters array if charter is provided (add if not already present)
	if data.Charter != nil && *data.Charter != "" {
		_, err = tx.Exec(ctx, `
			UPDATE songs SET charters = array_append(charters, $1)
			WHERE id = $2 AND NOT ($1 = ANY(charters))
		`, *data.Charter, songID)
		if err != nil {
			return 0, err
		}
	}

	// Create score
	var scoreID int64
	err = tx.QueryRow(ctx, `
		INSERT INTO scores (song_id, artist, charter, total_score, stars_achieved, players, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`, songID, data.Artist, data.Charter, data.TotalScore, data.StarsAchieved, nil, data.CreatedAt).Scan(&scoreID)
	if err != nil {
		return 0, err
	}

	// Create players
	for _, p := range data.Players {
		_, err = tx.Exec(ctx, `
			INSERT INTO players (score_id, name, instrument, difficulty, score, combo, accuracy, misses, rank, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`, scoreID, p.Name, p.Instrument, p.Difficulty, p.Score, p.Combo, p.Accuracy, p.Misses, p.Rank, data.CreatedAt)
		if err != nil {
			return 0, err
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return 0, err
	}

	return scoreID, nil
}

