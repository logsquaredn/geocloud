CREATE TYPE job_status AS ENUM ('waiting', 'inprogress', 'complete', 'error');

CREATE TABLE IF NOT EXISTS job ( 
    job_id VARCHAR (64) PRIMARY KEY,
    owner_id VARCHAR (64) NOT NULL,
    input_id VARCHAR (64) NOT NULL REFERENCES storage(storage_id),
    output_id VARCHAR (64) REFERENCES storage(storage_id),
    job_status JOB_STATUS NOT NULL DEFAULT 'waiting',
    job_error VARCHAR (512),
    start_time TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    end_time TIMESTAMP WITH TIME ZONE
);
