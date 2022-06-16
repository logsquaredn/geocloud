CREATE TYPE task_kind AS ENUM ('transformation', 'lookup');

CREATE TABLE IF NOT EXISTS task (
    task_type VARCHAR (32) PRIMARY KEY,
    task_kind TASK_KIND NOT NULL,
    task_params TEXT[],
    task_queue_id VARCHAR (512)
);
