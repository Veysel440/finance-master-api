package ports

import "time"

type User struct {
	ID       int64
	Name     string
	Email    string
	PassHash string
}

type Device struct {
	ID       int64
	UserID   int64
	DeviceID string
	Name     string
	LastSeen time.Time
}

type TotpSecret struct {
	UserID      int64
	Secret      string
	ConfirmedAt *time.Time
}

type AuthRepo interface {
	CreateUser(u *User) error
	FindUserByEmail(email string) (*User, error)

	// refresh tokens (hash saklanır)
	StoreRefresh(userID int64, refreshHash string, expires time.Time) error
	InvalidateRefresh(userID int64, refreshHash string) error
	HasValidRefresh(userID int64, refreshHash string, now time.Time) (bool, error)
	RotateRefresh(userID int64, oldHash, newHash string, newExp time.Time) error

	// cihaz ve TOTP
	UpsertDevice(userID int64, deviceID, name string, seen time.Time) error
	GetTotp(userID int64) (*TotpSecret, error)
	SetTotp(userID int64, secret string) error
	ConfirmTotp(userID int64) error
}
