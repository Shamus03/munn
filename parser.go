package munn

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type portfolioSpec struct {
	Accounts []struct {
		ID   int
		Name string
	}
	ManualAdjustments []struct {
		Account int
		Time    laxTime
		Balance float32
	}
	ScheduledTransactions []struct {
		FromAccount int
		ToAccount   int
		Description string
		Amount      float32
		Frequency   jsonFrequency
	}
}

type laxTime time.Time

func (l *laxTime) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	formats := []string{
		"2006-01-02",
		time.RFC3339,
	}
	var t time.Time
	var err error
	for _, f := range formats {
		t, err = time.Parse(f, s)
		if err == nil {
			*l = laxTime(t)
			return nil
		}
	}
	return fmt.Errorf("failed to parse time as any of the valid formats: last error: %v", err)
}

type jsonFrequency struct{ Frequency }

var jsonFrequencyRegex = regexp.MustCompile(`(\w+)(\(.*\))?`)

var daysOfWeek = map[string]time.Weekday{
	"Sunday":    time.Sunday,
	"Monday":    time.Monday,
	"Tuesday":   time.Tuesday,
	"Wednesday": time.Wednesday,
	"Thursday":  time.Thursday,
	"Friday":    time.Friday,
	"Saturday":  time.Saturday,
}

func (f *jsonFrequency) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	matches := jsonFrequencyRegex.FindStringSubmatch(s)
	if len(matches) > 1 {
		var args []string
		if len(matches) > 2 {
			args = strings.Split(strings.Trim(matches[2], "()"), ",")
		}

		switch matches[1] {
		case "Weekly":
			day := time.Sunday
			if len(args) > 0 {
				var ok bool
				if day, ok = daysOfWeek[args[0]]; !ok {
					return fmt.Errorf("invalid weekday: %s", args[0])
				}
			}
			*f = jsonFrequency{Weekly(day)}
			return nil
		case "Monthly":
			day := 1
			if len(args) > 0 {
				var err error
				day, err = strconv.Atoi(args[0])
				if err != nil {
					return err
				}
			}
			*f = jsonFrequency{Monthly(day)}
			return nil
		}
	}
	return fmt.Errorf("invalid frequency: %s", s)
}

func Parse(r io.Reader) (*Portfolio, error) {
	var spec portfolioSpec
	if err := json.NewDecoder(r).Decode(&spec); err != nil {
		return nil, err
	}

	p := &Portfolio{}

	accountsMap := make(map[int]*Account)

	for _, acc := range spec.Accounts {
		accountsMap[acc.ID] = p.NewAccount(acc.Name)
	}

	for _, man := range spec.ManualAdjustments {
		acc, ok := accountsMap[man.Account]
		if !ok {
			return nil, fmt.Errorf("invalid account: %d", man.Account)
		}
		p.NewManualAdjustment(acc, time.Time(man.Time), man.Balance)
	}

	for _, trans := range spec.ScheduledTransactions {
		var from *Account
		var to *Account
		if trans.FromAccount != 0 {
			var ok bool
			from, ok = accountsMap[trans.FromAccount]
			if !ok {
				return nil, fmt.Errorf("invalid account: %d", trans.FromAccount)
			}
		}
		if trans.ToAccount != 0 {
			var ok bool
			to, ok = accountsMap[trans.ToAccount]
			if !ok {
				return nil, fmt.Errorf("invalid account: %d", trans.ToAccount)
			}
		}
		p.NewScheduledTransaction(from, to, trans.Description, trans.Frequency, trans.Amount)
	}

	return p, nil
}
