CREATE TABLE players (
    id SERIAL PRIMARY KEY,
    chat_id BIGINT UNIQUE NOT NULL,  -- Telegram ID игрока
    team_id INTEGER,                 -- ID команды, в которой находится игрок
    position VARCHAR(50),            -- Позиция игрока
    height FLOAT,                    -- Рост игрока
    weight FLOAT,                    -- Вес игрока
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE SET NULL
);
