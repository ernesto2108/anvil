CREATE TABLE IF NOT EXISTS run_projects (
    run_id  TEXT NOT NULL,
    project TEXT NOT NULL,
    PRIMARY KEY (run_id, project)
);
