package main

import (
	"flag"
	"fmt"
	"log"
	"munn"
	"os"
	"path/filepath"
	"strings"
	"time"

	chart "github.com/wcharczuk/go-chart"
)

func main() {
	years := flag.Int("years", 3, "Number of years to project")
	image := flag.Bool("image", false, "Generate an image")
	debug := flag.Bool("debug", false, "Debug account changes")
	flag.Parse()
	args := flag.Args()

	if *debug {
		munn.DEBUG = true
	}

	if len(args) < 1 {
		log.Fatal("missing file name")
	}

	f, err := os.Open(args[0])
	if err != nil {
		log.Fatal(err)
	}

	p, err := munn.Parse(f)
	if err != nil {
		log.Fatal(err)
	}

	from := time.Now()
	for _, man := range p.ManualAdjustments {
		from = man.Time
		break
	}

	to := from.AddDate(*years, 0, 0)
	recs := p.Project(from, to)

	if *image {
		var sum float32
		var lastTime time.Time
		seriesMap := make(map[string]*chart.TimeSeries)
		for _, rec := range recs {
			if rec.Time != lastTime {
				sum = 0
			}
			lastTime = rec.Time
			s, ok := seriesMap[rec.AccountName]
			if !ok {
				s = &chart.TimeSeries{
					Name: rec.AccountName,
				}
				seriesMap[rec.AccountName] = s
			}
			sum += rec.Balance
			s.XValues = append(s.XValues, rec.Time)
			s.YValues = append(s.YValues, float64(sum))
		}

		var series []chart.Series
		for i := len(p.Accounts) - 1; i >= 0; i-- {
			series = append(series, seriesMap[p.Accounts[i].Name])
		}

		name := strings.TrimSuffix(args[0], filepath.Ext(args[0]))

		graph := chart.Chart{
			Series: series,
			XAxis: chart.XAxis{
				Name: "Date",
			},
			YAxis: chart.YAxis{
				Name: "Account Balance",
				ValueFormatter: func(v interface{}) string {
					n := v.(float64)
					return fmt.Sprintf("%.0fk", n/1000)
				},
			},
		}

		graph.Elements = []chart.Renderable{
			chart.Legend(&graph),
		}

		f, err := os.Create(name + ".png")
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		if err := graph.Render(chart.PNG, f); err != nil {
			log.Fatal(err)
		}
	} else {
		for _, r := range recs {
			fmt.Printf("%s\t%s\t%.2f\n", r.Time.Format("2006-01-02"), r.AccountName, r.Balance)
		}
	}
}
