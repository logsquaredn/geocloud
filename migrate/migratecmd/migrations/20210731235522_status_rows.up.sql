BEGIN;

INSERT INTO status_type(
    job_status
) VALUES (
    'COMPLETED'
);

INSERT INTO status_type(
    job_status
) VALUES (
    'IN PROGRESS'
);

INSERT INTO status_type(
    job_status
) VALUES (
    'WAITING'
);

INSERT INTO status_type(
    job_status
) VALUES (
    'ERROR'
);

COMMIT;
