CREATE TABLE IF NOT EXISTS job ( 
    job_id VARCHAR (64) PRIMARY KEY,
    customer_id VARCHAR (64) NOT NULL REFERENCES customer(customer_id),
    input_id VARCHAR (64) NOT NULL REFERENCES storage(storage_id),
    output_id VARCHAR (64) NOT NULL REFERENCES storage(storage_id),
    task_type VARCHAR (32) NOT NULL REFERENCES task(task_type),
    job_status VARCHAR (32) NOT NULL REFERENCES status_type(job_status) DEFAULT 'waiting',
    job_error VARCHAR (512),
    start_time TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    end_time TIMESTAMP WITH TIME ZONE,
    job_args TEXT[]
);
