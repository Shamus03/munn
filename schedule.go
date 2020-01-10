package munn

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func init() {
	registerScheduleParser("Weekly", &weeklySchedule{})
	registerScheduleParser("Biweekly", &biweeklySchedule{})
	registerScheduleParser("Monthly", &monthlySchedule{})
	registerScheduleParser("Once", &onceSchedule{})
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
