package mysql

import (
	"context"
	"errors"

	"github.com/Veysel440/finance-master-api/internal/errs"
	"github.com/Veysel440/finance-master-api/internal/ports"
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type WalletRepo struct{ db *sqlx.DB }

func NewWalletRepo(db *sqlx.DB) *WalletRepo { return &WalletRepo{db: db} }

func (r *WalletRepo) List(userID int64) ([]ports.Wallet, error) {
	var rows []ports.Wallet
	if err := r.db.Select(&rows,
		`SELECT id, user_id, name, currency FROM wallets WHERE user_id=? ORDER BY name`,
		userID,
	); err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *WalletRepo) Create(userID int64, w *ports.Wallet) error {
	res, err := r.db.ExecContext(context.Background(),
		`INSERT INTO wallets(user_id, name, currency) VALUES (?,?,?)`,
		userID, w.Name, w.Currency,
	)
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	w.ID = id
	return nil
}

func (r *WalletRepo) Update(userID int64, w *ports.Wallet) error {
	_, err := r.db.ExecContext(context.Background(),
		`UPDATE wallets SET name=?, currency=? WHERE id=? AND user_id=?`,
		w.Name, w.Currency, w.ID, userID,
	)
	return err
}

func (r *WalletRepo) Delete(userID, id int64) error {
	_, err := r.db.ExecContext(context.Background(),
		`DELETE FROM wallets WHERE id=? AND user_id=?`, id, userID,
	)
	if err != nil {
		var me *mysql.MySQLError
		if errors.As(err, &me) && me.Number == 1451 { // ER_ROW_IS_REFERENCED_2
			return errs.HasTransactions
		}
		return err
	}
	return nil
}
