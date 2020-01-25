package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(completionCmd)
	rootCmd.BashCompletionFunction = `
__munn_filename()
{
	if [[ ${#nouns[@]} -ge 1 ]]; then
		return 1
	fi
	COMPREPLY=( $(compgen -W "$(find . -maxdepth 3 -name '*.yaml')" -- $cur) )
}

__munn_custom_func() {
	case ${last_command} in
		munn)
			__munn_filename
			return
			;;
		*)
			;;
	esac
}`
}

var completionCmd = &cobra.Command{
	Use:    "completion",
	Hidden: true,
	Args:   cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return cmd.Root().GenBashCompletion(cmd.OutOrStdout())
		default:
			return fmt.Errorf("unknown shell type")
		}
	},
}
