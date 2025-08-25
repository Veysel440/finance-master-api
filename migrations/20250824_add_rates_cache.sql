-- +goose Up
CREATE TABLE IF NOT EXISTS rates_cache (
                                           base      CHAR(3)     NOT NULL PRIMARY KEY,
                                           rate_date DATE        NOT NULL,
                                           rates     JSON        NOT NULL,
                                           saved_at  DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- +goose Down
DROP TABLE IF EXISTS rates_cache;