SELECT storage_id, storage_status, namespace, storage_name, last_used, create_time 
FROM storage
WHERE last_used < $1;
