CREATE TABLE match_statistics (
    id SERIAL PRIMARY KEY,
    match_id INT REFERENCES matches(id) ON DELETE CASCADE,
    team_id1 INT REFERENCES teams(id) ON DELETE CASCADE,
    team_id2 INT REFERENCES teams(id) ON DELETE CASCADE,
    team1_score INT DEFAULT 0,
    team2_score INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
