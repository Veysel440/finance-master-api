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
ON DUPLICATE KEY UPDATE ua=VALUES(ua), ip=VALUES(ip),
  last_used_at=VALUES(last_used_at), expires_at=VALUES(expires_at)
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
	tx, err := r.db.BeginTxx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var sid int64
	err = tx.QueryRowx(`
SELECT id FROM sessions WHERE user_id=? AND refresh_hash=? FOR UPDATE
`, userID, oldHash).Scan(&sid)
	if err != nil {
		return err
	}

	if _, err = tx.Exec(`
INSERT INTO session_rotations (session_id, old_hash, new_hash, ua, ip, rotated_at)
VALUES (?,?,?,?,?,?)
`, sid, oldHash, newHash, ua, ip, now.UTC()); err != nil {
		return err
	}

	if _, err = tx.Exec(`
UPDATE sessions SET refresh_hash=?, ua=?, ip=?, last_used_at=?, expires_at=?
WHERE id=?
`, newHash, ua, ip, now.UTC(), exp.UTC(), sid); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *SessionRepo) ListSessions(userID int64, page, size int) ([]ports.Session, int, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 200 {
		size = 20
	}
	var total int
	if err := r.db.Get(&total, `SELECT COUNT(*) FROM sessions WHERE user_id=?`, userID); err != nil {
		return nil, 0, err
	}
	rows := []ports.Session{}
	err := r.db.Select(&rows, `
		SELECT id,user_id,ua,ip,created_at,last_seen,expires_at
		FROM sessions
		WHERE user_id=?
		ORDER BY last_seen DESC
		LIMIT ? OFFSET ?`, userID, size, (page-1)*size)
	return rows, total, err
}

func (r *SessionRepo) RevokeSession(userID, sessionID int64) error {
	_, err := r.db.Exec(`DELETE FROM sessions WHERE id=? AND user_id=?`, sessionID, userID)
	return err
}
