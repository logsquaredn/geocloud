UPDATE storage SET storage_status = $2, last_used = $3 WHERE storage_id = $1 RETURNING storage_id, storage_status, owner_id, storage_name, last_used, create_time;
