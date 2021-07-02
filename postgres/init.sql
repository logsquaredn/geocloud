CREATE TABLE job ( 
    job_id VARCHAR (36) PRIMARY KEY,
    job_type VARCHAR (32),
    job_status VARCHAR (32),
    job_error VARCHAR (128)
);

INSERT INTO job(job_id, job_type, job_status)
VALUES('7aff60d2-8a7e-4bc8-ae8d-6e9a98c98e0c', 'removeBadGeometry', 'COMPLETED');

INSERT INTO job(job_id, job_type, job_status) 
VALUES('7aff60d2-8a7e-4bc8-ae8d-6e9a98c98e0d', 'buffer', 'IN PROGRESS');

INSERT INTO job(job_id, job_type, job_status)
VALUES('7aff60d2-8a7e-4bc8-ae8d-6e9a98c98e0e', 'filter', 'COMPLETED');

INSERT INTO job(job_id, job_type, job_status, job_error)
VALUES('7aff60d2-8a7e-4bc8-ae8d-6e9a98c98e0f', 'reproject', 'ERROR', 'invalid geojson');
