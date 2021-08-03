DELETE FROM status_type WHERE status_type.job_status IN (
    'COMPLETED',
    'IN PROGRESS',
    'WAITING',
    'ERROR'
);
