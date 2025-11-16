CREATE TABLE IF NOT EXISTS pr_reviewers (
    pr_id UUID NOT NULL,
    reviewer_id UUID NOT NULL,
    assigned_at TIMESTAMP WITH TIME ZONE,
    
    PRIMARY KEY (pr_id, reviewer_id),
    FOREIGN KEY (pr_id) REFERENCES prs(id) ON DELETE CASCADE,
    FOREIGN KEY (reviewer_id) REFERENCES users(id) ON DELETE CASCADE
);


CREATE INDEX idx_pr_reviewers_reviewer_id ON pr_reviewers(reviewer_id);