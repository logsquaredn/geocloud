create table task (
    task_type VARCHAR (32),
    task_ref VARCHAR (512)
);

INSERT INTO task(task_type, task_ref) 
VALUES('buffer', 'logsquaredn/geocloud@sha256:6e0526e05dfbc9b09e37fbbe1f0f54c50d21dd19f78d549a84b857a6dae1f3bf');

INSERT INTO task(task_type, task_ref) 
VALUES('filter', 'logsquaredn/geocloud@sha256:b512a4292f2882aedffc679cfb8436eaa143d0fdb1bb66d72c40d9813e07d3e5');

INSERT INTO task(task_type, task_ref) 
VALUES('reproject', 'logsquaredn/geocloud@sha256:9394ad5f1e6c60c0c2d05c5e034705a08187e8815962bd2dfeb5b9ab00653412');

INSERT INTO task(task_type, task_ref) 
VALUES('removeBadGeometry', 'logsquaredn/geocloud@sha256:ac569317103eb0bb44dab3c2a7c9d5a865077c83128d206d3242cf78be6f9290');


CREATE TABLE job ( 
    job_id VARCHAR (36) PRIMARY KEY,
    task_type VARCHAR (32) NOT NULL REFERENCES task (task_type),
    job_status VARCHAR (32) NOT NULL,
    job_error VARCHAR (128)
);

INSERT INTO job(job_id, task_type, job_status)
VALUES('7aff60d2-8a7e-4bc8-ae8d-6e9a98c98e0c', 'removeBadGeometry', 'COMPLETED');

INSERT INTO job(job_id, task_type, job_status) 
VALUES('7aff60d2-8a7e-4bc8-ae8d-6e9a98c98e0d', 'buffer', 'IN PROGRESS');

INSERT INTO job(job_id, task_type, job_status)
VALUES('7aff60d2-8a7e-4bc8-ae8d-6e9a98c98e0e', 'filter', 'COMPLETED');

INSERT INTO job(job_id, task_type, job_status, job_error)
VALUES('7aff60d2-8a7e-4bc8-ae8d-6e9a98c98e0f', 'reproject', 'ERROR', 'invalid geojson');
