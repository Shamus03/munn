package munn

import (
	"sort"
	"time"
)

// ProjectionRecord is a record in a projection.
type ProjectionRecord struct {
	Time        time.Time
	AccountName string
	Balance     float32
}

// Project a portfolio's balances for a period of time.
func (p *Portfolio) Project(years int) []ProjectionRecord {
	// Apply all manual adjustments to get past data
	manGrp := make(map[time.Time][]*ManualAdjustment)
	var manTimes []time.Time
	for _, adj := range p.ManualAdjustments {
		if _, ok := manGrp[adj.Time]; !ok {
			manTimes = append(manTimes, adj.Time)
		}
		manGrp[adj.Time] = append(manGrp[adj.Time], adj)
	}
	sort.Sort(sortTime(manTimes))

	from := manTimes[0]
	to := from.AddDate(years, 0, 0)

	var recs []ProjectionRecord
	now := from

	recordedTimes := make(map[time.Time]bool)
	recordAccounts := func() {
		// Only record records for a given time once
		if recordedTimes[now] {
			return
		}
		recordedTimes[now] = true
		for _, acc := range p.Accounts {
			recs = append(recs, ProjectionRecord{
				Time:        now,
				AccountName: acc.Name,
				Balance:     acc.Balance,
			})
		}

		if p.RetirementPlan != nil && p.RetirementPlan.retireDate == nil {
			if p.TotalBalance() > p.RetirementPlan.BalanceNeeded(now) {
				rd := now
				p.RetirementPlan.retireDate = &rd
			}
		}
	}

	for _, t := range manTimes {
		now = t

		// Hacky way to ensure one-time transactions don't get applied if they fall within the manual adjustment period
		for _, trans := range p.Transactions {
			trans.Apply(now)
		}

		var changed bool
		for _, adj := range manGrp[now] {
			if adj.Apply(now) {
				changed = true
			}
		}

		if changed {
			recordAccounts()
		}
	}

	for ; !now.After(to); now = now.AddDate(0, 0, 1) {
		var changed bool

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
			recordAccounts()
		}
	}
	return recs
}

type sortTime []time.Time

func (t sortTime) Len() int           { return len(t) }
func (t sortTime) Less(i, j int) bool { return t[i].Before(t[j]) }
func (t sortTime) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
