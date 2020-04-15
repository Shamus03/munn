package main

import (
	"fmt"
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

func init() {
	rootCmd.Flags().IntP("years", "y", 3, "Number of years to project")
	rootCmd.Flags().BoolP("image", "i", false, "Generate an image")
	rootCmd.Flags().BoolP("stats", "s", false, "Print stats for the portfolio")
	rootCmd.Flags().BoolP("debug", "d", false, "Debug account changes")
	rootCmd.Flags().BoolP("watch", "w", false, "Watch input file")
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

			recs := p.Project(years)

			if stats {
				fmt.Println(p.Stats())
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
				fmt.Printf("Wrote image to %s\n", name)
			} else {
				for _, r := range recs {
					fmt.Printf("%s\t%s\t%.2f\n", r.Time.Format("2006-01-02"), r.AccountName, r.Balance)
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
					fmt.Println("Watching for changes...")
					select {
					case <-w.Event:
						if err := run(); err != nil {
							fmt.Println(err)
						}
					case err := <-w.Error:
						fmt.Println(err)
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
