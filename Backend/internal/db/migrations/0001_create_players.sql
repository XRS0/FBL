DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_tables WHERE tablename = 'players') THEN
        CREATE TABLE players (
            id SERIAL PRIMARY KEY,
            name VARCHAR(255) NOT NULL,
            height INTEGER NOT NULL,
            weight INTEGER NOT NULL,
            position VARCHAR(50) NOT NULL,
            chat_id BIGINT UNIQUE NOT NULL,
            contact VARCHAR(255),
            team_id INTEGER REFERENCES teams(id) ON DELETE SET NULL,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );
    END IF;
END $$;
