SELECT storage_id, storage_status, customer_id, storage_name, last_used, create_time  FROM storage WHERE customer_id = $1;
