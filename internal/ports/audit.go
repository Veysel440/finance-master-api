package ports

type AuditRepo interface {
	Insert(userID int64, action, entity string, entityID *int64, details string) error
}
