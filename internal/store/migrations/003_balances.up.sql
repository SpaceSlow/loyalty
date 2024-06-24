BEGIN;
CREATE TABLE balances(
    id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    user_id INTEGER NOT NULL,
    current DOUBLE PRECISION NOT NULL CHECK (current >= 0) DEFAULT 0,
    withdrawn DOUBLE PRECISION NOT NULL CHECK (withdrawn >= 0) DEFAULT 0,

    FOREIGN KEY (user_id) REFERENCES users (id)
);

CREATE FUNCTION update_balance()
    RETURNS trigger
    LANGUAGE plpgsql AS
$$
BEGIN
    UPDATE balances
    SET current=current+NEW.sum
    WHERE balances.user_id=NEW.user_id;
    RETURN NULL;
END
$$;

CREATE TRIGGER update_accrual_sum
    AFTER UPDATE OF sum ON accruals
    FOR EACH ROW
EXECUTE FUNCTION update_balance();

CREATE FUNCTION create_balance()
    RETURNS trigger
    LANGUAGE plpgsql AS
$$
BEGIN
    INSERT INTO balances (user_id) VALUES (NEW.id);
    RETURN NULL;
END
$$;

CREATE TRIGGER create_balance
    AFTER INSERT ON users
    FOR EACH ROW
    EXECUTE PROCEDURE create_balance();

INSERT INTO balances (user_id) SELECT id FROM users;
UPDATE balances SET current=(SELECT SUM(accruals.sum)
                             FROM accruals
                             WHERE balances.user_id=accruals.user_id)
                FROM accruals WHERE balances.user_id=accruals.user_id;
COMMIT;
