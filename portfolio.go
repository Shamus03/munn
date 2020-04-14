package munn

import (
	"time"
)

// Portfolio represents a person's financial portfolio.
type Portfolio struct {
	Accounts          []*Account
	Transactions      []*Transaction
	ManualAdjustments []*ManualAdjustment
	Debug             bool
}

// NewAccount adds a new account to the portfolio.
func (p *Portfolio) NewAccount(name string) *Account {
	a := &Account{
		Portfolio: p,
		Name:      name,
		Balance:   0,
	}
	p.Accounts = append(p.Accounts, a)
	return a
}

// NewManualAdjustment adds a new manual adjustment to the portfolio.
// It should be used to set the initial balance for an account or to log significant intended changes in the value of an account.
func (p *Portfolio) NewManualAdjustment(acc *Account, t time.Time, balance float32) {
	m := &ManualAdjustment{
		Portfolio: p,
		Account:   acc,
		Time:      t,
		Balance:   balance,
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
		Portfolio:   p,
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
	Portfolio *Portfolio
	Account   *Account
	Time      time.Time
	Balance   float32
	applied   bool
}

// Apply the manual adjustment.
func (a *ManualAdjustment) Apply(now time.Time) bool {
	if a.applied || now.Before(a.Time) {
		return false
	}

	a.applied = true

	diff := a.Balance - a.Account.Balance
	a.Portfolio.logDebug("%s, Applied manual adjustment for account %s from %.2f to %.2f (%.2f difference)\n",
		now.Format("2006-01-02"),
		a.Account.Name,
		a.Account.Balance,
		a.Balance,
		diff,
	)
	a.Account.Balance = a.Balance
	return true
}

// Transaction is a transaction from one account to another.
// It may have a schedule to repeat the transaction on some interval.
// If FromAccount or ToAccount is nil, this transaction represents money in/out of the portfolio (payments, income, etc.).
// Otherwise it is a transfer between two accounts in the portfolio.
type Transaction struct {
	Description string
	Portfolio   *Portfolio
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

	t.Portfolio.logDebug("%s, Applied transaction %s\n", now.Format("2006-01-02"), t.Description)
	return true
}

// Account is a named account with a balance.
// An account may also have an annual interest rate which is applied monthly.
type Account struct {
	Name               string
	Portfolio          *Portfolio
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
	a.Portfolio.logDebug("%s, Account %s gained interest\n", now.Format("2006-01-02"), a.Name)

	monthlyInterest := a.AnnualInterestRate / 12
	a.Balance = a.Balance * (1 + monthlyInterest)

	return true
}
