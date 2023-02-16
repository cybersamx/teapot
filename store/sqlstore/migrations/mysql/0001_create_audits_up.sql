# mysql/0001_create_audits_up.sql

CREATE TABLE IF NOT EXISTS audits
(
  request_id     VARCHAR(36) PRIMARY KEY,
  created_at     BIGINT,
  client_agent   VARCHAR(192),
  client_address VARCHAR(64),
  status_code    INT,
  error          VARCHAR(256),
  event          VARCHAR(512),
  KEY idx_request_id (request_id)
);
