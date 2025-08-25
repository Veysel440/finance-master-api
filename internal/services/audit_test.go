package services

import (
	"testing"

	"github.com/Veysel440/finance-master-api/internal/ports"
)

type fakeAuditRepo struct {
	last struct {
		uid      int64
		action   string
		entity   string
		entityID *int64
		details  string
	}
}

func (f *fakeAuditRepo) Insert(uid int64, action, entity string, entityID *int64, details string) error {
	f.last.uid, f.last.action, f.last.entity, f.last.entityID, f.last.details = uid, action, entity, entityID, details
	return nil
}

var _ ports.AuditRepo = (*fakeAuditRepo)(nil)

func TestAudit_Log_SerializesDetails(t *testing.T) {
	r := &fakeAuditRepo{}
	a := &AuditService{Repo: r}

	id := int64(42)
	a.Log(7, "tx.create", "transaction", &id, map[string]any{"amount": 10, "currency": "TRY"})

	if r.last.uid != 7 || r.last.action != "tx.create" || r.last.entity != "transaction" {
		t.Fatalf("unexpected fields: %+v", r.last)
	}
	if r.last.entityID == nil || *r.last.entityID != id {
		t.Fatalf("entity id not set")
	}
	if r.last.details == "" || r.last.details[0] != '{' {
		t.Fatalf("details not json")
	}
}

func TestAudit_NoRepo_NoPanic(t *testing.T) {
	var a *AuditService
	a.Log(1, "x", "y", nil, nil)
	(&AuditService{}).Log(1, "x", "y", nil, nil)
}
