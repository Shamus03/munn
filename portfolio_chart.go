package munn

import (
	"fmt"
	"time"

	chart "github.com/wcharczuk/go-chart"
)

// Chart generates a chart for the projection.
func (p Portfolio) Chart(recs []ProjectionRecord) chart.Chart {
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

	graph := chart.Chart{
		Series: series,
		XAxis: chart.XAxis{
			Name: "Date",
		},
		YAxis: chart.YAxis{
			Name: "Account Balance",
			ValueFormatter: func(v interface{}) string {
				n := v.(float64)
				if n == 0 {
					return "0"
				}
				return fmt.Sprintf("%.0fk", n/1000)
			},
		},
	}

	graph.Elements = []chart.Renderable{
		chart.Legend(&graph),
	}

	return graph
}
