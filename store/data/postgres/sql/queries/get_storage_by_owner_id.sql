SELECT storage_id, storage_status, owner_id, storage_name, last_used, create_time FROM storage WHERE owner_id = $1 ORDER BY create_time OFFSET $2 LIMIT $3;
