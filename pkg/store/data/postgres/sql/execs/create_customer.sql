INSERT INTO customer (
    customer_id,
    api_key,
    email
) VALUES (
    $1,
    $2,
    $3
) ON CONFLICT DO NOTHING;
