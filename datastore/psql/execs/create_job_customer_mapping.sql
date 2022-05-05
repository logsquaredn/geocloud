INSERT INTO job_customer_mapping (
    job_id,
    customer_id
) VALUES (
    $1,
    $2
) RETURNING customer_id;
