CREATE TABLE matches (
    id SERIAL PRIMARY KEY,
    team1_id INT REFERENCES teams(id) ON DELETE CASCADE,
    team2_id INT REFERENCES teams(id) ON DELETE CASCADE,
    date TIMESTAMP NOT NULL,
    location VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);