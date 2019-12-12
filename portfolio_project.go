package munn

import (
	"time"
)

// ProjectionRecord is a record in a projection.
type ProjectionRecord struct {
	Time        time.Time
	AccountName string
	Balance     float32
}

// Project a portfolio's balances for a period of time.
func (p *Portfolio) Project(from, to time.Time) []ProjectionRecord {
	// Set up initial balances
	for _, adj := range p.ManualAdjustments {
		if adj.Time.Before(from) {
			adj.Apply(adj.Time)
		}
	}

	var recs []ProjectionRecord

	// Always show first period
	changed := true
	for now := from; now.Before(to); now = now.AddDate(0, 0, 1) {
		for _, adj := range p.ManualAdjustments {
			if adj.Apply(now) {
				changed = true
			}
		}

		for _, acc := range p.Accounts {
			if acc.GainInterest(now) {
				changed = true
			}
		}

		for _, trans := range p.Transactions {
			if trans.Apply(now) {
				changed = true
			}
		}

		if changed {
			for _, acc := range p.Accounts {
				recs = append(recs, ProjectionRecord{
					Time:        now,
					AccountName: acc.Name,
					Balance:     acc.Balance,
				})
			}
		}
		changed = false
	}
	return recs
}
