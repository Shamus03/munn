package main

import update "github.com/Shamus03/cobra-update"

func init() {
	rootCmd.AddCommand(update.Command("Shamus03", "munn"))
}
