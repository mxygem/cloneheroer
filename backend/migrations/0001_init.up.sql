CREATE TABLE IF NOT EXISTS artists (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS songs (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    artist_id INTEGER REFERENCES artists(id) ON DELETE CASCADE,
    charters TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[],
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(name, artist_id)
);

CREATE TABLE IF NOT EXISTS scores (
    id SERIAL PRIMARY KEY,
    song_id INTEGER REFERENCES songs(id) ON DELETE CASCADE,
    artist TEXT NOT NULL,
    charter TEXT,
    total_score BIGINT,
    stars_achieved INTEGER,
    players JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS players (
    id SERIAL PRIMARY KEY,
    score_id INTEGER REFERENCES scores(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    instrument TEXT,
    difficulty TEXT,
    score BIGINT,
    combo INTEGER,
    accuracy NUMERIC,
    misses INTEGER,
    rank INTEGER,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_scores_song_id ON scores(song_id);
CREATE INDEX IF NOT EXISTS idx_scores_created_at ON scores(created_at);
CREATE INDEX IF NOT EXISTS idx_players_score_id ON players(score_id);
CREATE INDEX IF NOT EXISTS idx_songs_artist_id ON songs(artist_id);

