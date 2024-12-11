DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'uni_match_statistics_match_id'
    ) THEN
        ALTER TABLE match_statistics DROP CONSTRAINT uni_match_statistics_match_id;
    END IF;
END $$;

ALTER TABLE match_statistics ADD CONSTRAINT uni_match_statistics_match_id UNIQUE (match_id);

CREATE TABLE IF NOT EXISTS match_statistics (
    id SERIAL PRIMARY KEY,
    match_id INT UNIQUE NOT NULL REFERENCES matches(id) ON DELETE CASCADE,
    team_id1 INT NOT NULL REFERENCES teams(id),
    team_id2 INT NOT NULL REFERENCES teams(id),
    team1_score INT DEFAULT 0,
    team2_score INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

