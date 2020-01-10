package munn

import (
	"fmt"
	"io"
	"regexp"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v2"
)

var (
	scheduleParsersLock sync.Mutex
	scheduleParsers     = make(map[string]ScheduleParser)
)

// RegisterSchedulerParser registers a schedule parser
func RegisterScheduleParser(name string, parser ScheduleParser) {
	if parser == nil {
		panic("parser cannot be nil")
	}

	scheduleParsersLock.Lock()
	defer scheduleParsersLock.Unlock()

	if _, ok := scheduleParsers[name]; ok {
		panic(fmt.Sprintf("parser already registered for name: %s", name))
	}

	scheduleParsers[name] = parser
}

// GetScheduleParser gets a schedule parser
func GetScheduleParser(name string) (ScheduleParser, bool) {
	scheduleParsersLock.Lock()
	defer scheduleParsersLock.Unlock()

	parser, ok := scheduleParsers[name]
	if !ok {
		return nil, false
	}
	return parser, true
}

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

// ScheduleParser parses a schedule
type ScheduleParser interface {
	ParseSchedule(args []string) (Schedule, error)
}

func (f *jsonSchedule) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}

	matches := jsonScheduleRegex.FindStringSubmatch(s)
	if len(matches) == 0 {
		return fmt.Errorf("invalid schedule: %s", s)
	}

	var args []string
	if len(matches) > 2 {
		args = strings.Fields(strings.Trim(matches[2], "()"))
	}

	parser, ok := getScheduleParser(matches[1])
	if !ok {
		return fmt.Errorf("no schedule parser registered for name: %s", matches[1])
	}

	schedule, err := parser.ParseSchedule(args)
	if err != nil {
		return fmt.Errorf("error parsing schedule: %v", err)
	}

	*f = jsonSchedule{schedule}
	return nil
}
