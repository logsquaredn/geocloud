INSERT INTO customer (
    job_id,
    customer_id
) VALUES (
    $1,
    $2
) RETURNING customer_id;
