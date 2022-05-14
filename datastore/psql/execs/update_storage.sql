UPDATE storage SET last_used = $2 WHERE storage_id = $1 RETURNING storage_id, customer_id, storage_name, last_used;
