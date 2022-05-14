INSERT INTO customer (
    customer_id
) VALUES (
    $1
) ON CONFLICT DO NOTHING;
