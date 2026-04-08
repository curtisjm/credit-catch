-- 002_oauth_and_statement_day.down.sql

ALTER TABLE user_cards DROP COLUMN statement_close_day;
DROP TABLE IF EXISTS oauth_accounts;
ALTER TABLE users ALTER COLUMN password SET NOT NULL;
