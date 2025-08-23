package mysql

import "github.com/jmoiron/sqlx"

type AuditRepo struct{ db *sqlx.DB }

func NewAuditRepo(db *sqlx.DB) *AuditRepo { return &AuditRepo{db: db} }

func (r *AuditRepo) Insert(userID int64, action, entity string, entityID *int64, details string) error {
	_, err := r.db.Exec(`INSERT INTO audit_logs(user_id, action, entity, entity_id, details) VALUES (?,?,?,?,?)`,
		userID, action, entity, entityID, details)
	return err
}
