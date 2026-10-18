DROP TABLE IF EXISTS email_verifications;
ALTER TABLE users DROP COLUMN email_verified;
ALTER TABLE users DROP COLUMN email;
