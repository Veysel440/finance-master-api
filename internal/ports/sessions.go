package ports

import "time"

type Session struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"userId"`
	UA         string    `json:"ua"`
	IP         string    `json:"ip"`
	CreatedAt  time.Time `json:"createdAt"`
	LastUsedAt time.Time `json:"lastUsedAt"`
	ExpiresAt  time.Time `json:"expiresAt"`
}

type SessionRepo interface {
	StoreRefreshMeta(userID int64, refreshHash, ua, ip string, exp, now time.Time) error
	ValidateRefresh(userID int64, refreshHash, ua, ip string, now time.Time, bindUA, bindIP bool) (bool, error)
	RotateRefreshMeta(userID int64, oldHash, newHash, ua, ip string, exp, now time.Time) error

	ListSessions(userID int64) ([]Session, error)
	RevokeSession(userID, sessionID int64) error
}

type LoginGuardRepo interface {
	IncLoginFail(email, ip string, at time.Time, window time.Duration) (fails int, until *time.Time, err error)
	ResetLoginFail(email, ip string) error
	LockUser(userID int64, until time.Time) error
	GetUserLock(userID int64) (*time.Time, error)
}
