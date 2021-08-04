CREATE TABLE IF NOT EXISTS task (
    task_type VARCHAR (32) PRIMARY KEY,
    task_params TEXT[],
    task_queue_name VARCHAR (512) NOT NULL,
    task_ref VARCHAR (512) NOT NULL
);
