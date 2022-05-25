SELECT s.storage_id, s.customer_id, s.storage_name, s.last_used, s.create_time FROM storage s INNER JOIN job j ON s.storage_id = j.output_id WHERE j.job_id = $1;
