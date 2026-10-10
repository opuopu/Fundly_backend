-- 000001_init_schema.down.sql
-- Revert initial schema

DROP TABLE IF EXISTS user_terms_acknowledgments;
DROP TABLE IF EXISTS terms_and_conditions;
DROP TABLE IF EXISTS feedbacks;
DROP TABLE IF EXISTS user_configs;
DROP TABLE IF EXISTS attachments;
DROP TABLE IF EXISTS budgets;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS accounts;
DROP TABLE IF EXISTS users;
