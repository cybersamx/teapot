# mysql/0001_create_users_up.sql

CREATE TABLE IF NOT EXISTS users
(
  id            VARCHAR(36)  PRIMARY KEY,
  created_at    BIGINT,
  updated_at    BIGINT,
  deleted_at    BIGINT,
  email         VARCHAR(128) UNIQUE,
  username      VARCHAR(32)  UNIQUE,
  first_name    VARCHAR(64),
  last_name     VARCHAR(64),
  is_verified   BOOLEAN,
  roles         VARCHAR(64),
  password      VARCHAR(128),
  KEY idx_first_name ((lower(first_name))),
  KEY idx_last_name ((lower(last_name)))
);
