CREATE TABLE IF NOT EXISTS sink (
    sink_id VARCHAR (64) PRIMARY KEY,
    sink_address VARCHAR(256) NOT NULL,
    job_id VARCHAR (64) NOT NULL REFERENCES job(job_id)
);
