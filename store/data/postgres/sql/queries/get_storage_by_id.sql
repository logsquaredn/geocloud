SELECT storage_id, storage_status, namespace, storage_name, last_used, create_time  FROM storage WHERE storage_id = $1;
