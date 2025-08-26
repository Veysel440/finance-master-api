package cron

import (
	"context"
	"log"
	"time"

	"github.com/Veysel440/finance-master-api/internal/services"
)

func StartRatesWarm(ctx context.Context, s *services.RatesService, bases []string, every time.Duration) (stop func()) {
	if s == nil || every <= 0 || len(bases) == 0 {
		return func() {}
	}
	tkr := time.NewTicker(every)
	done := make(chan struct{})

	warm := func() {
		for _, b := range bases {
			if _, _, _, err := s.Latest(b); err != nil {
				log.Printf("rates warm %s: %v", b, err)
			}
		}
	}
	go func() {
		warm()
		for {
			select {
			case <-tkr.C:
				warm()
			case <-ctx.Done():
				close(done)
				return
			}
		}
	}()
	return func() { tkr.Stop(); <-done }
}
