package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Shamus03/munn"
	"github.com/spf13/cobra"
	chart "github.com/wcharczuk/go-chart"
)

func init() {
	rootCmd.Flags().IntP("years", "y", 3, "Number of years to project")
	rootCmd.Flags().BoolP("image", "i", false, "Generate an image")
	rootCmd.Flags().BoolP("debug", "d", false, "Debug account changes")
}

var rootCmd = &cobra.Command{
	Use:  "munn",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		years, _ := cmd.Flags().GetInt("years")
		image, _ := cmd.Flags().GetBool("image")
		debug, _ := cmd.Flags().GetBool("debug")

		f, err := os.Open(args[0])
		if err != nil {
			log.Fatal(err)
		}

		p, err := munn.Parse(f)
		if err != nil {
			log.Fatal(err)
		}
		p.Debug = debug

		from := time.Now()
		for _, man := range p.ManualAdjustments {
			from = man.Time
			break
		}

		to := from.AddDate(years, 0, 0)
		recs := p.Project(from, to)

		if image {
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
	},
}
