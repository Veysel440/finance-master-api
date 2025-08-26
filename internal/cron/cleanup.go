package cron

import (
	"context"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
)

func StartCleanup(ctx context.Context, db *sqlx.DB, keepAuditDays, keepSessionDays int, every time.Duration) (stop func()) {
	if db == nil || every <= 0 {
		return func() {}
	}
	tkr := time.NewTicker(every)
	done := make(chan struct{})

	run := func() {
		if keepAuditDays > 0 {
			_, err := db.Exec(`DELETE FROM audit_logs WHERE created_at < UTC_TIMESTAMP() - INTERVAL ? DAY`, keepAuditDays)
			if err != nil {
				log.Println("cleanup audit_logs:", err)
			}
		}
		if keepSessionDays > 0 {
			_, err := db.Exec(`DELETE FROM sessions WHERE expires_at < UTC_TIMESTAMP() - INTERVAL ? DAY`, keepSessionDays)
			if err != nil {
				log.Println("cleanup sessions:", err)
			}
		}
	}

	go func() {
		run()
		for {
			select {
			case <-tkr.C:
				run()
			case <-ctx.Done():
				close(done)
				return
			}
		}
	}()
	return func() { tkr.Stop(); <-done }
}
