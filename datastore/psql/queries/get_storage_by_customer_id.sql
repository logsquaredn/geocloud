SELECT storage_id, customer_id, storage_name, last_used FROM storage WHERE customer_id = $1 ORDER BY last_used OFFSET $2 LIMIT $3;
