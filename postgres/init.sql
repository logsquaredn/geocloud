CREATE TABLE job ( 
    job_id VARCHAR (16) PRIMARY KEY,
    job_type VARCHAR (50),
    job_status VARCHAR (50)
);

INSERT INTO job(job_id, job_type, job_status)
VALUES('1', 'first job', 'completed');

INSERT INTO job(job_id, job_type, job_status) 
VALUES('2', 'second job', 'pending');

INSERT INTO job(job_id, job_type, job_status)
VALUES('3', 'third job', 'completed');
