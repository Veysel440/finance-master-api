package mysql

import (
	"context"
	"database/sql"
	"time"

	"github.com/Veysel440/finance-master-api/internal/ports"
	"github.com/jmoiron/sqlx"
)

type SessionRepo struct{ db *sqlx.DB }

func NewSessionRepo(db *sqlx.DB) *SessionRepo { return &SessionRepo{db: db} }

func (r *SessionRepo) StoreRefreshMeta(userID int64, hash, ua, ip string, exp, now time.Time) error {
	_, err := r.db.ExecContext(context.Background(), `
		INSERT INTO sessions (user_id, refresh_hash, ua, ip, created_at, last_used_at, expires_at)
		VALUES (?,?,?,?,?,?,?)
		ON DUPLICATE KEY UPDATE ua=VALUES(ua), ip=VALUES(ip), last_used_at=VALUES(last_used_at), expires_at=VALUES(expires_at)
	`, userID, hash, ua, ip, now.UTC(), now.UTC(), exp.UTC())
	return err
}

func (r *SessionRepo) ValidateRefresh(userID int64, hash, ua, ip string, now time.Time, bindUA, bindIP bool) (bool, error) {
	var s ports.Session
	err := r.db.Get(&s, `
		SELECT id, user_id, ua, ip, created_at, last_used_at, expires_at
		FROM sessions WHERE user_id=? AND refresh_hash=? LIMIT 1
	`, userID, hash)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if now.After(s.ExpiresAt) {
		return false, nil
	}
	if bindUA && s.UA != ua {
		return false, nil
	}
	if bindIP && s.IP != ip {
		return false, nil
	}
	_, _ = r.db.ExecContext(context.Background(),
		`UPDATE sessions SET last_used_at=? WHERE id=?`, now.UTC(), s.ID)
	return true, nil
}

func (r *SessionRepo) RotateRefreshMeta(userID int64, oldHash, newHash, ua, ip string, exp, now time.Time) error {
	// Var olan kaydı yeni hash ile güncelle
	_, err := r.db.ExecContext(context.Background(), `
		UPDATE sessions SET refresh_hash=?, ua=?, ip=?, last_used_at=?, expires_at=?
		WHERE user_id=? AND refresh_hash=?`,
		newHash, ua, ip, now.UTC(), exp.UTC(), userID, oldHash)
	return err
}

func (r *SessionRepo) ListSessions(userID int64) ([]ports.Session, error) {
	var rows []ports.Session
	err := r.db.Select(&rows, `
		SELECT id, user_id, ua, ip, created_at, last_used_at, expires_at
		FROM sessions WHERE user_id=? ORDER BY last_used_at DESC
	`, userID)
	return rows, err
}

func (r *SessionRepo) RevokeSession(userID, sessionID int64) error {
	_, err := r.db.ExecContext(context.Background(),
		`DELETE FROM sessions WHERE id=? AND user_id=?`, sessionID, userID)
	return err
}

var _ ports.SessionRepo = (*SessionRepo)(nil)
