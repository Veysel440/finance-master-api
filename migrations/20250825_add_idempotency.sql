-- +goose Up
CREATE TABLE IF NOT EXISTS idempotency_keys (
                                                user_id     BIGINT      NOT NULL,
                                                idem_key    VARCHAR(128) NOT NULL,
                                                resource    VARCHAR(64)  NOT NULL,
                                                resource_id BIGINT       NOT NULL,
                                                created_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
                                                PRIMARY KEY (user_id, idem_key),
                                                KEY idx_idem_resource (user_id, resource, resource_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- +goose Down
DROP TABLE IF EXISTS idempotency_keys;