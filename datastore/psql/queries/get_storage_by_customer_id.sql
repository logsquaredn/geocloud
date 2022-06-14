SELECT storage_id, customer_id, storage_name, last_used, create_time FROM storage WHERE customer_id = $1 ORDER BY create_time OFFSET $2 LIMIT $3;

