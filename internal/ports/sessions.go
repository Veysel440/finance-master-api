package ports

import "time"

type Session struct {
	ID        int64     `db:"id"         json:"id"`
	UserID    int64     `db:"user_id"    json:"userId"`
	UA        string    `db:"ua"         json:"ua"`
	IP        string    `db:"ip"         json:"ip"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
	LastSeen  time.Time `db:"last_seen"  json:"lastSeen"`
	ExpiresAt time.Time `db:"expires_at" json:"expiresAt"`
}

type SessionRepo interface {
	StoreRefreshMeta(userID int64, refreshHash, ua, ip string, exp, now time.Time) error
	ValidateRefresh(userID int64, refreshHash, ua, ip string, now time.Time, bindUA, bindIP bool) (bool, error)
	RotateRefreshMeta(userID int64, oldHash, newHash, ua, ip string, exp, now time.Time) error
	ListSessions(userID int64, page, size int) (rows []Session, total int, err error)
	RevokeSession(userID, sessionID int64) error
}

type LoginGuardRepo interface {
	IncLoginFail(email, ip string, at time.Time, window time.Duration) (fails int, until *time.Time, err error)
	ResetLoginFail(email, ip string) error
	LockUser(userID int64, until time.Time) error
	GetUserLock(userID int64) (*time.Time, error)
}
