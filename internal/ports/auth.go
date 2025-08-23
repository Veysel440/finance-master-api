package ports

import "time"

type User struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	PassHash string `json:"-"`
}

type Device struct {
	ID       int64     `json:"id"`
	UserID   int64     `json:"userId"`
	DeviceID string    `json:"deviceId"`
	Name     string    `json:"name"`
	LastSeen time.Time `json:"lastSeen"`
}

type TotpSecret struct {
	UserID      int64      `json:"userId"`
	Secret      string     `json:"-"`
	ConfirmedAt *time.Time `json:"confirmedAt,omitempty"`
}

type AuthRepo interface {
	CreateUser(u *User) error
	FindUserByEmail(email string) (*User, error)

	StoreRefresh(userID int64, refreshHash string, expires time.Time) error
	InvalidateRefresh(userID int64, refreshHash string) error
	HasValidRefresh(userID int64, refreshHash string, now time.Time) (bool, error)
	RotateRefresh(userID int64, oldHash, newHash string, newExp time.Time) error

	UpsertDevice(userID int64, deviceID, name string, seen time.Time) error
	GetTotp(userID int64) (*TotpSecret, error)
	SetTotp(userID int64, secret string) error
	ConfirmTotp(userID int64) error
}
