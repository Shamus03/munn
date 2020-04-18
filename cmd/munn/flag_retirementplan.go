package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Shamus03/munn"
)

type retirementPlanFlag struct {
	RetirementPlan *munn.RetirementPlan
}

func (f *retirementPlanFlag) Set(s string) error {
	spl := strings.Split(s, ":")
	if len(spl) != 2 {
		return fmt.Errorf("expected 'deathDate:yearlyExpenses' eg. '2006-01-02:123'")
	}
	date, err := time.Parse("2006-01-02", spl[0])
	if err != nil {
		return err
	}

	y, err := strconv.ParseFloat(spl[1], 32)
	if err != nil {
		return err
	}

	f.RetirementPlan = &munn.RetirementPlan{
		DeathDate:      date,
		YearlyExpenses: float32(y),
	}
	return nil
}

func (f *retirementPlanFlag) String() string {
	if f.RetirementPlan == nil {
		return ""
	}
	return fmt.Sprintf("%s:%.2f", f.RetirementPlan.DeathDate.Format("2006-01-02"), f.RetirementPlan.YearlyExpenses)
}

func (f *retirementPlanFlag) Type() string {
	return "RetirementPlan"
}
