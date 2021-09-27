BEGIN;

INSERT INTO status_type(
    job_status
) VALUES (
    'waiting'
);

INSERT INTO status_type(
    job_status
) VALUES (
    'inprogress'
);

INSERT INTO status_type(
    job_status
) VALUES (
    'complete'
);

INSERT INTO status_type(
    job_status
) VALUES (
    'error'
);

COMMIT;
