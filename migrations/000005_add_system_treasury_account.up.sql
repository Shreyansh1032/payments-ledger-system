ALTER TABLE accounts ADD COLUMN is_system BOOLEAN NOT NULL DEFAULT false;

ALTER TABLE accounts DROP CONSTRAINT accounts_balance_check;
ALTER TABLE accounts ADD CONSTRAINT accounts_balance_check CHECK (balance >= 0 OR is_system = true);

INSERT INTO users (username, first_name, last_name, password_hash)
VALUES ('system_treasury', 'System', 'Treasury', 'no-login-system-account');

INSERT INTO accounts (user_id, balance, is_system)
SELECT id, 0, true FROM users WHERE username = 'system_treasury';
