CREATE TABLE IF NOT EXISTS job ( 
    job_id VARCHAR (64) PRIMARY KEY,
    task_type VARCHAR (32) NOT NULL REFERENCES task(task_type),
    job_status VARCHAR (32) NOT NULL REFERENCES status_type(job_status),
    job_error VARCHAR (128)
);
