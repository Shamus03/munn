package munn

import "fmt"

// Stats gets stats for a portfolio.
func (p *Portfolio) Stats() PortfolioStats {
	var yearlyExpenses float32
	var yearlyIncome float32
	for _, t := range p.Transactions {
		if t.ToAccount == nil {
			yearlyExpenses += t.Amount * t.Schedule.YearlyFactor()
		} else if len(t.FromAccounts) > 0 {
			yearlyIncome += t.Amount * t.Schedule.YearlyFactor()
		}
	}
	return PortfolioStats{
		AverageMonthlyExpenses: yearlyExpenses / 12,
		AverageMonthlyIncome:   yearlyIncome / 12,
		AverageMonthlyGrowth:   (yearlyIncome - yearlyExpenses) / 12,
	}
}

// PortfolioStats is a collection of stats about the portfolio.
type PortfolioStats struct {
	AverageMonthlyExpenses float32
	AverageMonthlyIncome   float32
	AverageMonthlyGrowth   float32
}

func (s PortfolioStats) String() string {
	var o string
	o += fmt.Sprintf("Average monthly expenses:  $%.2f\n", s.AverageMonthlyExpenses)
	o += fmt.Sprintf("Average monthly income:    $%.2f\n", s.AverageMonthlyIncome)
	o += fmt.Sprintf("Average monthly growth:    $%.2f", s.AverageMonthlyGrowth)
	return o
}
