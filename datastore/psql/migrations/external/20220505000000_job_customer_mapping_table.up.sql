CREATE TABLE IF NOT EXISTS job_customer_mapping ( 
    job_id VARCHAR (64) not null REFERENCES job(job_id) ON DELETE CASCADE,
    customer_id VARCHAR (64) not null REFERENCES customer(customer_id),
    PRIMARY KEY (job_id, customer_id)
);
