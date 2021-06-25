CREATE TABLE job ( 
    job_id INT PRIMARY KEY,
    job_name VARCHAR (50),
)

INSERT INTO job(job_id, job_name)
VALUES(1, 'first job');

INSERT INTO job(job_id, job_name) 
VALUES(2, 'second job');

INSERT INTO job(job_id, job_name)
VALUES(3, 'third job');
