ALTER TABLE players RENAME COLUMN best_streak TO combo;
ALTER TABLE players RENAME COLUMN notes_missed TO misses;
ALTER TABLE players DROP COLUMN IF EXISTS total_notes;
ALTER TABLE players DROP COLUMN IF EXISTS notes_hit;
ALTER TABLE players DROP COLUMN IF EXISTS avg_multiplier;
