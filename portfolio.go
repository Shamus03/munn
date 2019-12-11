package munn

import (
	"fmt"
	"time"
)

// Portfolio represents a person's financial portfolio.
type Portfolio struct {
	Accounts          []*Account
	Transactions      []*Transaction
	ManualAdjustments []*ManualAdjustment
}

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

// NewAccount adds a new account to the portfolio.
func (p *Portfolio) NewAccount(name string) *Account {
	a := &Account{
		Name:    name,
		Balance: 0,
	}
	p.Accounts = append(p.Accounts, a)
	return a
}

// NewManualAdjustment adds a new manual adjustment to the portfolio.
// It should be used to set the initial balance for an account or to log significant intended changes in the value of an account.
func (p *Portfolio) NewManualAdjustment(acc *Account, t time.Time, balance float32) {
	m := &ManualAdjustment{
		Account: acc,
		Time:    t,
		Balance: balance,
	}
	var i int
	for i = 0; i < len(p.ManualAdjustments); i++ {
		if p.ManualAdjustments[i].Time.After(t) {
			break
		}
	}
	newArr := append(p.ManualAdjustments, nil)
	copy(newArr[i+1:], newArr[i:])
	newArr[i] = m

	p.ManualAdjustments = newArr
}

// NewTransaction adds a new transaction to the portfolio.
func (p *Portfolio) NewTransaction(from, to *Account, desc string, s Schedule, amt float32) *Transaction {
	t := &Transaction{
		Description: desc,
		Schedule:    s,
		FromAccount: from,
		ToAccount:   to,
		Amount:      amt,
	}
	p.Transactions = append(p.Transactions, t)
	return t
}

// ManualAdjustment is a single manual adjustment made on an account.
type ManualAdjustment struct {
	Account *Account
	Time    time.Time
	Balance float32
	applied bool
}

// Apply the manual adjustment.
func (a *ManualAdjustment) Apply(now time.Time) bool {
	if a.applied || now.Before(a.Time) {
		return false
	}

	a.applied = true
	a.Account.Balance = a.Balance
	return true
}

// Schedule determines the next time for a transaction to be applied, based on the last time it was applied.
type Schedule interface {
	ShouldApply(time.Time) bool
}

// Weekly schedule will run weekly on the given weekday.
func Weekly(day time.Weekday) Schedule {
	return &weeklySchedule{
		weekday: day,
	}
}

type weeklySchedule struct {
	weekday     time.Weekday
	lastApplied time.Time
}

func (s *weeklySchedule) ShouldApply(t time.Time) bool {
	n := s.lastApplied.AddDate(0, 0, 7)
	for n.Weekday() != s.weekday {
		n = n.AddDate(0, 0, -1)
	}
	if t.Before(n) {
		return false
	}
	s.lastApplied = t
	return true
}

type monthlySchedule struct {
	day         int
	lastApplied time.Time
}

// Monthly schedule will run monthly on the given day of the month.
func Monthly(day int) Schedule {
	return &monthlySchedule{
		day: day,
	}
}

func (s *monthlySchedule) ShouldApply(t time.Time) bool {
	year, month, _ := s.lastApplied.AddDate(0, 1, 0).Date()
	if t.Before(time.Date(year, month, s.day, 0, 0, 0, 0, time.Local)) {
		return false
	}
	s.lastApplied = t
	return true
}

type onceSchedule struct {
	time    time.Time
	applied bool
}

// Once schedule will run once at the given time.
func Once(t time.Time) Schedule {
	return &onceSchedule{
		time: t,
	}
}

func (s *onceSchedule) ShouldApply(t time.Time) bool {
	if s.applied || t.Before(s.time) {
		return false
	}
	s.applied = true
	return true
}

// Transaction is a transaction from one account to another.
// It may have a schedule to repeat the transaction on some interval.
// If FromAccount or ToAccount is nil, this transaction represents money in/out of the portfolio (payments, income, etc.).
// Otherwise it is a transfer between two accounts in the portfolio.
type Transaction struct {
	Description string
	Schedule    Schedule
	FromAccount *Account
	ToAccount   *Account
	Amount      float32
}

// Apply the transaction.
func (t *Transaction) Apply(now time.Time) bool {
	if !t.Schedule.ShouldApply(now) {
		return false
	}

	// If either account is nil, the transaction represents money in/out of the overall portfolio
	if t.FromAccount != nil {
		t.FromAccount.Balance -= t.Amount
	}
	if t.ToAccount != nil {
		t.ToAccount.Balance += t.Amount
	}

	fmt.Printf("Applied %s, %s\n", now.Format("2006-01-02"), t.Description)
	return true
}

// Account is a named account with a balance.
// An account may also have an annual interest rate which is applied monthly.
type Account struct {
	Name               string
	Balance            float32
	AnnualInterestRate float32
	interestSchedule   Schedule
}

// GainInterest adds interest to the account.
func (a *Account) GainInterest(now time.Time) bool {
	if a.interestSchedule == nil {
		a.interestSchedule = Monthly(1)
	}

	if !a.interestSchedule.ShouldApply(now) {
		return false
	}

	monthlyInterest := a.AnnualInterestRate / 12
	a.Balance = a.Balance * (1 + monthlyInterest)

	return true
}
