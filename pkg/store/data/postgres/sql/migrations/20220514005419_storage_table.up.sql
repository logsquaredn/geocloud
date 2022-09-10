CREATE TYPE storage_status AS ENUM ('final', 'unknown', 'unusable', 'transformable');

CREATE TABLE IF NOT EXISTS storage ( 
    storage_id VARCHAR (64) PRIMARY KEY,
    storage_status STORAGE_STATUS NOT NULL DEFAULT 'unknown',
    customer_id VARCHAR (64) NOT NULL REFERENCES customer(customer_id),
    storage_name VARCHAR (64),
    last_used TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    create_time TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
