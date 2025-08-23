-- migrations/0004_tx_indexes.sql
-- +goose Up
ALTER TABLE transactions
    ADD INDEX idx_tx_user_date_type (user_id, occurred_at, type);
ALTER TABLE transactions
    ADD FULLTEXT INDEX ftx_tx_note (note);

-- +goose Down
ALTER TABLE transactions
    DROP INDEX idx_tx_user_date_type,
    DROP INDEX ftx_tx_note;