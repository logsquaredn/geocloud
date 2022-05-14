INSERT INTO customer (
    customer_id,
    customer_name
) VALUES (
    $1,
    $2
) ON CONFLICT DO NOTHING RETURNING customer_id, customer_name;
