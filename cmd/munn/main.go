package main

import (
	"flag"
	"log"
	"munn"
	"os"
	"time"
)

func main() {
	years := flag.Int("years", 3, "Number of years to project")
	flag.Parse()
	args := flag.Args()

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
	p.Project(from, to)
}
