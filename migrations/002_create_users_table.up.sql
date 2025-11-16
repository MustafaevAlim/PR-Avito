CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    username VARCHAR(100) NOT NULL,
    team_name VARCHAR(100) NOT NULL,
    is_active BOOLEAN NOT NULL,

    FOREIGN KEY (team_name) REFERENCES teams(team_name) ON DELETE SET NULL
);

