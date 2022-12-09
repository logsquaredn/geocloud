CREATE TABLE IF NOT EXISTS step (
    step_id VARCHAR (64) PRIMARY KEY,
    job_id VARCHAR (64) NOT NULL REFERENCES job(job_id) ON DELETE CASCADE,
    task_type VARCHAR (32) NOT NULL REFERENCES task(task_type),
    job_args TEXT[]
);
