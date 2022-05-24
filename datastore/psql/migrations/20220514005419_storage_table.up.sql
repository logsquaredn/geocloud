CREATE TABLE IF NOT EXISTS storage ( 
    storage_id VARCHAR (64) PRIMARY KEY,
    customer_id VARCHAR (64) NOT NULL REFERENCES customer(customer_id),
    storage_name VARCHAR (64),
    last_used TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    create_time TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
