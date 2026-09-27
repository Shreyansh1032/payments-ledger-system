UPDATE accounts SET balance = 0 WHERE is_system = true;
DELETE FROM accounts WHERE is_system = true;
DELETE FROM users WHERE username = 'system_treasury';

ALTER TABLE accounts DROP CONSTRAINT accounts_balance_check;
ALTER TABLE accounts ADD CONSTRAINT accounts_balance_check CHECK (balance >= 0);
ALTER TABLE accounts DROP COLUMN is_system;
