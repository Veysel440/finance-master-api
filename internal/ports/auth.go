package ports

import "time"

type User struct {
	ID                    int64
	Name, Email, PassHash string
}
type AuthRepo interface {
	CreateUser(u *User) error
	FindUserByEmail(email string) (*User, error)
	StoreRefresh(userID int64, refreshHash string, expires time.Time) error
	InvalidateRefresh(userID int64, refreshHash string) error
}
