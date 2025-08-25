package services

import (
	"encoding/json"

	"github.com/Veysel440/finance-master-api/internal/ports"
)

type AuditService struct{ Repo ports.AuditRepo }

func (a *AuditService) Log(uid int64, action, entity string, entityID *int64, details map[string]any) {
	if a == nil || a.Repo == nil {
		return
	}
	var payload string
	if details != nil {
		if b, err := json.Marshal(details); err == nil {
			payload = string(b)
		}
	}
	_ = a.Repo.Insert(uid, action, entity, entityID, payload)
}
