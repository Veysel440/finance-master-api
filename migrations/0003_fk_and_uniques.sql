-- +goose Up
-- Indeksler
ALTER TABLE wallets    ADD INDEX idx_wallet_user (user_id);
ALTER TABLE categories ADD INDEX idx_cat_user (user_id);
ALTER TABLE transactions
    ADD INDEX idx_tx_wallet (wallet_id),
    ADD INDEX idx_tx_category (category_id),
    ADD INDEX idx_tx_user_occ (user_id, occurred_at);

-- Benzersiz kısıtlar (kullanıcı içinde aynı ad tek olsun)
ALTER TABLE wallets    ADD UNIQUE KEY uniq_wallet_user_name (user_id, name);
ALTER TABLE categories ADD UNIQUE KEY uniq_cat_user_type_name (user_id, type, name);

-- Dış anahtarlar
ALTER TABLE transactions
    ADD CONSTRAINT fk_t_wallet   FOREIGN KEY (wallet_id)   REFERENCES wallets(id)    ON UPDATE CASCADE ON DELETE RESTRICT,
    ADD CONSTRAINT fk_t_category FOREIGN KEY (category_id) REFERENCES categories(id) ON UPDATE CASCADE ON DELETE RESTRICT;

-- +goose Down
ALTER TABLE transactions
    DROP FOREIGN KEY fk_t_wallet,
    DROP FOREIGN KEY fk_t_category,
    DROP INDEX idx_tx_wallet,
    DROP INDEX idx_tx_category,
    DROP INDEX idx_tx_user_occ;
ALTER TABLE wallets    DROP INDEX idx_wallet_user, DROP INDEX uniq_wallet_user_name;
ALTER TABLE categories DROP INDEX idx_cat_user,    DROP INDEX uniq_cat_user_type_name;