CREATE TABLE IF NOT EXISTS teams (
    id UUID PRIMARY KEY,
    team_name VARCHAR(100) UNIQUE NOT NULL
);


CREATE INDEX idx_teams_team_name ON teams(team_name);