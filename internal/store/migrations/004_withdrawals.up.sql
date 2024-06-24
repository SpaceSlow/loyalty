CREATE TABLE withdrawals(
    id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    user_id INTEGER NOT NULL,
    order_number BIGINT UNIQUE NOT NULL,
    sum DOUBLE PRECISION NOT NULL CHECK (sum >= 0),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    FOREIGN KEY (user_id) REFERENCES users (id)
);

CREATE FUNCTION update_balance_withdrawal()
    RETURNS trigger
    LANGUAGE plpgsql AS
$$
BEGIN
    UPDATE balances
    SET current=current-NEW.sum, withdrawn=withdrawn+NEW.sum
    WHERE balances.user_id=NEW.user_id;
    RETURN NULL;
END
$$;

CREATE TRIGGER update_balance
    AFTER INSERT ON withdrawals
    FOR EACH ROW
EXECUTE FUNCTION update_balance_withdrawal();

ALTER FUNCTION update_balance() RENAME TO update_balance_accrual;
ALTER TRIGGER update_accrual_sum ON accruals RENAME TO update_balance;
