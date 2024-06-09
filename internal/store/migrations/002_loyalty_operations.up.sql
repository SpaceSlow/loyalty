CREATE TYPE loyalty_operation_status AS ENUM ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED');

CREATE TABLE loyalty_operations(
    id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    user_id INTEGER NOT NULL,
    order_number BIGINT UNIQUE NOT NULL,
    status loyalty_operation_status NOT NULL DEFAULT 'NEW',
    accrual DOUBLE PRECISION NOT NULL CHECK (accrual >= 0) DEFAULT 0,
    withdrawal DOUBLE PRECISION NOT NULL CHECK (withdrawal >= 0) DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    withdrawn_at TIMESTAMP WITH TIME ZONE,

    FOREIGN KEY (user_id) REFERENCES users (id)
);
