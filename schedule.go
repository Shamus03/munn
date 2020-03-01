package munn

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func init() {
	RegisterScheduleParser("Weekly", &weeklySchedule{})
	RegisterScheduleParser("Biweekly", &biweeklySchedule{})
	RegisterScheduleParser("Monthly", &monthlySchedule{})
	RegisterScheduleParser("Once", &onceSchedule{})
}

// Schedule determines the next time for a transaction to be applied, based on the last time it was applied.
// YearlyFactor should return the average number of times the schedule will be applied in a year (eg. a weekly schedule is applied 52 times in a year)
type Schedule interface {
	ShouldApply(time.Time) bool
	YearlyFactor() float32
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

func (s *weeklySchedule) YearlyFactor() float32 {
	return 52
}

func (s *weeklySchedule) ParseSchedule(args []string) (Schedule, error) {
	day := time.Sunday
	if len(args) > 0 {
		var ok bool
		day, ok = daysOfWeek[strings.ToLower(args[0])]
		if !ok {
			return nil, fmt.Errorf("invalid weekday: %s", args[0])
		}
	}
	return Weekly(day), nil
}

// Biweekly schedule will run biweekly on the given weekday.
func Biweekly(day time.Weekday) Schedule {
	return &biweeklySchedule{
		weekday: day,
	}
}

type biweeklySchedule struct {
	weekday     time.Weekday
	lastApplied time.Time
}

func (s *biweeklySchedule) ShouldApply(t time.Time) bool {
	n := s.lastApplied.AddDate(0, 0, 14)
	for n.Weekday() != s.weekday {
		n = n.AddDate(0, 0, -1)
	}
	if t.Before(n) {
		return false
	}
	s.lastApplied = t
	return true
}

func (s *biweeklySchedule) YearlyFactor() float32 {
	return 26
}

func (s *biweeklySchedule) ParseSchedule(args []string) (Schedule, error) {
	day := time.Sunday
	if len(args) > 0 {
		var ok bool
		day, ok = daysOfWeek[strings.ToLower(args[0])]
		if !ok {
			return nil, fmt.Errorf("invalid weekday: %s", args[0])
		}
	}
	return Biweekly(day), nil
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

func (s *monthlySchedule) YearlyFactor() float32 {
	return 12
}

func (s *monthlySchedule) ParseSchedule(args []string) (Schedule, error) {
	day := 1
	if len(args) > 0 {
		var err error
		day, err = strconv.Atoi(args[0])
		if err != nil {
			return nil, err
		}
	}
	return Monthly(day), nil
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

// Since the transaction only applies once, don't consider it in yearly projections.
func (s *onceSchedule) YearlyFactor() float32 {
	return 0
}

func (s *onceSchedule) ParseSchedule(args []string) (Schedule, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("Once schedule requires a date")
	}

	t, err := time.Parse("2006-01-02", args[0])
	if err != nil {
		return nil, err
	}
	return Once(t), nil
}
