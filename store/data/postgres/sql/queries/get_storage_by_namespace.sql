SELECT storage_id, storage_status, namespace, storage_name, last_used, create_time FROM storage WHERE namespace = $1 ORDER BY create_time OFFSET $2 LIMIT $3;
