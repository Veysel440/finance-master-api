package ports

type IdempotencyRepo interface {
	Get(userID int64, key, resource string) (int64, bool, error)
	Save(userID int64, key, resource string, resourceID int64) error
}
