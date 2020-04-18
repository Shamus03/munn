package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Shamus03/munn"
	"github.com/radovskyb/watcher"
	"github.com/spf13/cobra"
	chart "github.com/wcharczuk/go-chart"
)

var retirementPlan retirementPlanFlag

func init() {
	setupRootCmd()
}

func setupRootCmd() {
	rootCmd.Flags().IntP("years", "y", 3, "Number of years to project")
	rootCmd.Flags().BoolP("image", "i", false, "Generate an image")
	rootCmd.Flags().BoolP("stats", "s", false, "Print stats for the portfolio")
	rootCmd.Flags().BoolP("debug", "d", false, "Debug account changes")
	rootCmd.Flags().BoolP("watch", "w", false, "Watch input file")
	retirementPlan.RetirementPlan = nil
	rootCmd.Flags().VarP(&retirementPlan, "retire", "r", "Use a retirement plan")
	rootCmd.SetOut(os.Stdout)
}

var rootCmd = &cobra.Command{
	Use:  "munn",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		years, _ := cmd.Flags().GetInt("years")
		image, _ := cmd.Flags().GetBool("image")
		stats, _ := cmd.Flags().GetBool("stats")
		debug, _ := cmd.Flags().GetBool("debug")
		watch, _ := cmd.Flags().GetBool("watch")
		fileName := args[0]

		run := func() error {
			f, err := os.Open(fileName)
			if err != nil {
				return err
			}

			p, err := munn.Parse(f)
			if err != nil {
				return err
			}
			p.Debug = debug

			if retirementPlan.RetirementPlan != nil {
				p.RetirementPlan = retirementPlan.RetirementPlan
			}

			recs := p.Project(years)

			if stats {
				cmd.Println(p.Stats())
			}

			if image {
				name := strings.TrimSuffix(fileName, filepath.Ext(fileName)) + ".png"
				f, err := os.Create(name)
				if err != nil {
					return err
				}
				defer f.Close()

				if err := p.Chart(recs).Render(chart.PNG, f); err != nil {
					return err
				}
				cmd.Printf("Wrote image to %s\n", name)
			} else {
				for _, r := range recs {
					cmd.Printf("%s\t%s\t%.2f\n", r.Time.Format("2006-01-02"), r.AccountName, r.Balance)
				}
			}

			if retirementPlan.RetirementPlan != nil {
				date, ok := retirementPlan.RetirementPlan.RetireDate()
				if ok {
					cmd.Printf("Retirement date: %s\n", date.Format("2006-01-02"))
				} else {
					cmd.Printf("Retirement date: could not find\n")
				}
			}

			return nil
		}

		if err := run(); err != nil {
			log.Fatal(err)
		}

		if watch {
			w := watcher.New()
			w.SetMaxEvents(1)
			w.FilterOps(watcher.Write)
			w.Add(fileName)

			go func() {
				for {
					cmd.Println("Watching for changes...")
					select {
					case <-w.Event:
						if err := run(); err != nil {
							cmd.Println(err)
						}
					case err := <-w.Error:
						cmd.Println(err)
					case <-w.Closed:
						return
					}
				}
			}()

			if err := w.Start(time.Millisecond * 100); err != nil {
				log.Fatal(err)
			}
		}
	},
}
