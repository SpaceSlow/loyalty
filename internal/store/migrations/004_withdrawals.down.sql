ALTER TRIGGER update_balance ON accruals RENAME TO update_accrual_sum;
ALTER FUNCTION update_balance_accrual() RENAME TO update_balance;

DROP TRIGGER update_balance ON withdrawals;
DROP FUNCTION update_balance_withdrawal;
DROP TABLE withdrawals;
