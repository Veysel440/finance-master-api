package services

import (
	"encoding/json"

	"github.com/Veysel440/finance-master-api/internal/obs"
	"github.com/Veysel440/finance-master-api/internal/ports"
)

type AuditService struct{ Repo ports.AuditRepo }

func (a *AuditService) Log(uid int64, action, entity string, entityID *int64, details map[string]any) {
	if a == nil || a.Repo == nil {
		return
	}
	var d string
	if details != nil {
		safe := obs.MaskPIIMap(details)
		b, _ := json.Marshal(safe)
		d = string(b)
	}
	_ = a.Repo.Insert(uid, action, entity, entityID, d)
}
