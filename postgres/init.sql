CREATE TABLE job ( 
    job_id VARCHAR (16) PRIMARY KEY,
    job_type VARCHAR (32),
    job_status VARCHAR (32),
    job_error VARCHAR (128)
);

INSERT INTO job(job_id, job_type, job_status)
VALUES('1', 'removeBadGeometry', 'completed');

INSERT INTO job(job_id, job_type, job_status) 
VALUES('2', 'buffer', 'pending');

INSERT INTO job(job_id, job_type, job_status)
VALUES('3', 'filter', 'completed');

INSERT INTO job(job_id, job_type, job_status)
VALUES('4', 'reproject', 'failed', 'error: invalid geojson');
