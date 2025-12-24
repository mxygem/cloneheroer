ALTER TABLE players RENAME COLUMN combo TO best_streak;
ALTER TABLE players RENAME COLUMN misses TO notes_missed;
ALTER TABLE players ADD COLUMN total_notes INTEGER;
ALTER TABLE players ADD COLUMN notes_hit INTEGER;
ALTER TABLE players ADD COLUMN avg_multiplier NUMERIC;
ALTER TABLE players ADD COLUMN overhits INTEGER;
