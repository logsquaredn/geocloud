INSERT INTO storage (
    storage_id,
    storage_status,
    owner_id,
    storage_name
) VALUES (
    $1,
    $2,
    $3,
    $4
) RETURNING storage_id, storage_status, owner_id, storage_name, last_used, create_time;
