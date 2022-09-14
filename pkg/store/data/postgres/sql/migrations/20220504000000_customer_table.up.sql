CREATE TABLE IF NOT EXISTS customer ( 
    customer_id VARCHAR (64) PRIMARY KEY,
    api_key VARCHAR (64) NOT NULL,
    email VARCHAR (512) NOT NULL
);
