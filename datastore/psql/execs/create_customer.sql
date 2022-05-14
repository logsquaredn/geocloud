INSERT INTO customer (
    customer_id
) VALUES (
    $1
) ON CONFLICT DO NOTHING RETURNING customer_id;
