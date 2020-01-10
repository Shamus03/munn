package munn

import (
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

// Parse will read a portfolio from an io.Reader.
func Parse(r io.Reader) (*Portfolio, error) {
	var spec portfolioSpec
	if err := yaml.NewDecoder(r).Decode(&spec); err != nil {
		return nil, err
	}

	p := &Portfolio{}

	accountsMap := make(map[int]*Account)

	for _, accSpec := range spec.Accounts {
		if _, ok := accountsMap[accSpec.ID]; ok {
			return nil, fmt.Errorf("duplicate account ID: %d", accSpec.ID)
		}
		acc := p.NewAccount(accSpec.Name)
		accountsMap[accSpec.ID] = acc

		if accSpec.AnnualInterestRate != 0 {
			acc.AnnualInterestRate = accSpec.AnnualInterestRate
		}
	}

	for _, man := range spec.ManualAdjustments {
		acc, ok := accountsMap[man.Account]
		if !ok {
			return nil, fmt.Errorf("invalid account: %d", man.Account)
		}
		if man.Balance == nil {
			return nil, fmt.Errorf("manual adjustment missing balance")
		}
		p.NewManualAdjustment(acc, time.Time(man.Time), *man.Balance)
	}

	for _, trans := range spec.Transactions {
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
		if trans.Schedule.parsed == nil {
			return nil, fmt.Errorf("transaction '%s' missing schedule", trans.Description)
		}
		if trans.Amount == nil {
			return nil, fmt.Errorf("transaction '%s' missing amount", trans.Description)
		}
		p.NewTransaction(from, to, trans.Description, trans.Schedule.parsed, *trans.Amount)
	}

	return p, nil
}

type portfolioSpec struct {
	Accounts []struct {
		ID                 int     `yaml:"id"`
		Name               string  `yaml:"name"`
		AnnualInterestRate float32 `yaml:"annualInterestRate"`
	}
	ManualAdjustments []struct {
		Account int      `yaml:"account"`
		Time    laxTime  `yaml:"time"`
		Balance *float32 `yaml:"balance"`
	} `yaml:"manualAdjustments"`
	Transactions []struct {
		FromAccount int          `yaml:"fromAccount"`
		ToAccount   int          `yaml:"toAccount"`
		Description string       `yaml:"description"`
		Amount      *float32     `yaml:"amount"`
		Schedule    jsonSchedule `yaml:"schedule"`
	} `yaml:"transactions"`
}

type laxTime time.Time

func (l *laxTime) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
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

type jsonSchedule struct {
	parsed Schedule
}

var jsonScheduleRegex = regexp.MustCompile(`^(\w+)(\(.*\))?$`)

var daysOfWeek = map[string]time.Weekday{
	"sunday":    time.Sunday,
	"monday":    time.Monday,
	"tuesday":   time.Tuesday,
	"wednesday": time.Wednesday,
	"thursday":  time.Thursday,
	"friday":    time.Friday,
	"saturday":  time.Saturday,
}

func (f *jsonSchedule) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}

	matches := jsonScheduleRegex.FindStringSubmatch(s)
	if len(matches) > 1 {
		var args []string
		if len(matches) > 2 {
			args = strings.Fields(strings.Trim(matches[2], "()"))
		}

		switch matches[1] {
		case "Biweekly":
			day := time.Sunday
			if len(args) > 0 {
				var ok bool
				if day, ok = daysOfWeek[strings.ToLower(args[0])]; !ok {
					return fmt.Errorf("invalid weekday: %s", args[0])
				}
			}
			*f = jsonSchedule{Biweekly(day)}
			return nil
		case "Weekly":
			day := time.Sunday
			if len(args) > 0 {
				var ok bool
				if day, ok = daysOfWeek[strings.ToLower(args[0])]; !ok {
					return fmt.Errorf("invalid weekday: %s", args[0])
				}
			}
			*f = jsonSchedule{Weekly(day)}
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
			f.parsed = Monthly(day)
			return nil
		case "Once":
			if len(args) != 1 {
				return fmt.Errorf("Once schedule requires a date")
			}

			t, err := time.Parse("2006-01-02", args[0])
			if err != nil {
				return err
			}
			f.parsed = Once(t)
			return nil
		}
	}
	return fmt.Errorf("invalid schedule: %s", s)
}
