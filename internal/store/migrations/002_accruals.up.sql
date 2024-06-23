CREATE TYPE accrual_status AS ENUM ('PROCESSING', 'INVALID', 'PROCESSED');

CREATE TABLE accruals(
    id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    user_id INTEGER NOT NULL,
    order_number BIGINT UNIQUE NOT NULL,
    status accrual_status NOT NULL DEFAULT 'PROCESSING',
    sum DOUBLE PRECISION NOT NULL CHECK (sum >= 0) DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    FOREIGN KEY (user_id) REFERENCES users (id)
);

CREATE VIEW unprocessed_orders_view AS SELECT order_number FROM accruals WHERE status='PROCESSING';
