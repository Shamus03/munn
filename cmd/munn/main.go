package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Shamus03/munn"
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
		name := strings.TrimSuffix(args[0], filepath.Ext(args[0]))
		f, err := os.Create(name + ".png")
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		if err := p.Chart(recs).Render(chart.PNG, f); err != nil {
			log.Fatal(err)
		}
	} else {
		for _, r := range recs {
			fmt.Printf("%s\t%s\t%.2f\n", r.Time.Format("2006-01-02"), r.AccountName, r.Balance)
		}
	}
}
