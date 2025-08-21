package mysql

import (
	"database/sql"
	"time"

	"github.com/Veysel440/finance-master-api/internal/ports"
	"github.com/jmoiron/sqlx"
)

type AuthRepo struct{ db *sqlx.DB }

func NewAuthRepo(db *sqlx.DB) *AuthRepo { return &AuthRepo{db: db} }

func (r *AuthRepo) CreateUser(u *ports.User) error {
	res, err := r.db.Exec(`INSERT INTO users(name,email,pass_hash) VALUES (?,?,?)`, u.Name, u.Email, u.PassHash)
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	u.ID = id
	return nil
}
func (r *AuthRepo) FindUserByEmail(email string) (*ports.User, error) {
	var u ports.User
	err := r.db.Get(&u, `SELECT id,name,email,pass_hash FROM users WHERE email=? LIMIT 1`, email)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *AuthRepo) StoreRefresh(userID int64, refreshHash string, expires time.Time) error {
	_, err := r.db.Exec(`INSERT INTO sessions(user_id, refresh_hash, expires_at) VALUES (?,?,?)`,
		userID, refreshHash, expires)
	return err
}
func (r *AuthRepo) InvalidateRefresh(userID int64, refreshHash string) error {
	_, err := r.db.Exec(`DELETE FROM sessions WHERE user_id=? AND refresh_hash=?`, userID, refreshHash)
	return err
}
func (r *AuthRepo) HasValidRefresh(userID int64, refreshHash string, now time.Time) (bool, error) {
	var n int
	err := r.db.Get(&n, `SELECT COUNT(*) FROM sessions WHERE user_id=? AND refresh_hash=? AND expires_at > ?`,
		userID, refreshHash, now)
	return n > 0, err
}
func (r *AuthRepo) RotateRefresh(userID int64, oldHash, newHash string, newExp time.Time) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()
	if _, err = tx.Exec(`DELETE FROM sessions WHERE user_id=? AND refresh_hash=?`, userID, oldHash); err != nil {
		return err
	}
	_, err = tx.Exec(`INSERT INTO sessions(user_id, refresh_hash, expires_at) VALUES (?,?,?)`, userID, newHash, newExp)
	return err
}

func (r *AuthRepo) UpsertDevice(userID int64, deviceID, name string, seen time.Time) error {
	_, err := r.db.Exec(`
		INSERT INTO user_devices(user_id, device_id, name, last_seen)
		VALUES (?,?,?,?)
		ON DUPLICATE KEY UPDATE name=VALUES(name), last_seen=VALUES(last_seen)`,
		userID, deviceID, name, seen)
	return err
}

func (r *AuthRepo) GetTotp(userID int64) (*ports.TotpSecret, error) {
	var t ports.TotpSecret
	err := r.db.Get(&t, `SELECT user_id, secret, confirmed_at FROM totp_secrets WHERE user_id=? LIMIT 1`, userID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}
func (r *AuthRepo) SetTotp(userID int64, secret string) error {
	_, err := r.db.Exec(`
		INSERT INTO totp_secrets(user_id, secret) VALUES(?,?)
		ON DUPLICATE KEY UPDATE secret=VALUES(secret), confirmed_at=NULL`, userID, secret)
	return err
}
func (r *AuthRepo) ConfirmTotp(userID int64) error {
	_, err := r.db.Exec(`UPDATE totp_secrets SET confirmed_at=NOW() WHERE user_id=?`, userID)
	return err
}
