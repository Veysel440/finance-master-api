package mysql

import (
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
	err := r.db.Get(&u, `SELECT id, name, email, pass_hash FROM users WHERE email = ? LIMIT 1`, email)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *AuthRepo) StoreRefresh(userID int64, refreshHash string, expires time.Time) error {
	_, err := r.db.Exec(`INSERT INTO sessions (user_id, refresh_hash, expires_at) VALUES (?,?,?)`,
		userID, refreshHash, expires)
	return err
}

func (r *AuthRepo) InvalidateRefresh(userID int64, refreshHash string) error {
	_, err := r.db.Exec(`DELETE FROM sessions WHERE user_id=? AND refresh_hash=?`, userID, refreshHash)
	return err
}
