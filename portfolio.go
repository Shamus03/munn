package munn

import (
	"fmt"
	"time"
)

type Portfolio struct {
	Accounts              []*Account
	ScheduledTransactions []*ScheduledTransaction
	ManualAdjustments     []*ManualAdjustment
}

func (p *Portfolio) Project(from, to time.Time) {
	// Set up initial balances
	for _, adj := range p.ManualAdjustments {
		if adj.Time.Before(from) {
			adj.Apply(adj.Time)
		}
	}
	for _, trans := range p.ScheduledTransactions {
		trans.lastApplied = from
	}

	// Always print first period
	changed := true
	for now := from; now.Before(to); now = now.AddDate(0, 0, 1) {
		for _, adj := range p.ManualAdjustments {
			if adj.Apply(now) {
				changed = true
			}
		}

		for _, trans := range p.ScheduledTransactions {
			if trans.Apply(now) {
				changed = true
			}
		}

		if changed {
			for _, acc := range p.Accounts {
				fmt.Printf("%s\t%s\t%.2f\n", now.Format("2006-01-02"), acc.Name, acc.Balance)
			}
		}
		changed = false
	}
}

func (p *Portfolio) NewAccount(name string) *Account {
	a := &Account{
		Name:    name,
		Balance: 0,
	}
	p.Accounts = append(p.Accounts, a)
	return a
}

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

func (p *Portfolio) NewScheduledTransaction(from, to *Account, desc string, f Frequency, amt float32) *ScheduledTransaction {
	t := &ScheduledTransaction{
		Description: desc,
		Frequency:   f,
		FromAccount: from,
		ToAccount:   to,
		Amount:      amt,
	}
	p.ScheduledTransactions = append(p.ScheduledTransactions, t)
	return t
}

type ManualAdjustment struct {
	Account *Account
	Time    time.Time
	Balance float32
	applied bool
}

func (a *ManualAdjustment) Apply(now time.Time) bool {
	if a.applied || now.Before(a.Time) {
		return false
	}

	a.applied = true
	a.Account.Balance = a.Balance
	return true
}

type Frequency interface {
	Next(time.Time) time.Time
}

func Weekly(day time.Weekday) Frequency {
	return weeklyFrequency{day}
}

type weeklyFrequency struct {
	weekday time.Weekday
}

func (f weeklyFrequency) Next(t time.Time) time.Time {
	n := t.AddDate(0, 0, 7)
	for n.Weekday() != f.weekday {
		n = n.AddDate(0, 0, -1)
	}
	return n
}

type monthlyFrequency struct {
	day int
}

func Monthly(day int) Frequency {
	return monthlyFrequency{day}
}

func (f monthlyFrequency) Next(t time.Time) time.Time {
	year, month, _ := t.AddDate(0, 1, 0).Date()
	return time.Date(year, month, f.day, 0, 0, 0, 0, time.Local)
}

type ScheduledTransaction struct {
	Description string
	Frequency   Frequency
	FromAccount *Account
	ToAccount   *Account
	Amount      float32
	lastApplied time.Time
}

func (s *ScheduledTransaction) Apply(now time.Time) bool {
	next := s.Frequency.Next(s.lastApplied)
	if now.Before(next) {
		return false
	}

	// If either account is nil, the transaction represents money in/out of the overall portfolio
	if s.FromAccount != nil {
		s.FromAccount.Balance -= s.Amount
	}
	if s.ToAccount != nil {
		s.ToAccount.Balance += s.Amount
	}
	s.lastApplied = now

	fmt.Printf("Applied %s, %s\n", now.Format("2006-01-02"), s.Description)
	return true
}

type Account struct {
	Name    string
	Balance float32
}
