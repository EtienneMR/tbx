// Package system registers the `tbx system` sub-tree.
package system

import "github.com/spf13/cobra"

var Cmd = &cobra.Command{
	Use:   "system",
	Short: "System health and update utilities",
}

func init() {
	Cmd.AddCommand(statusCmd, updateCmd)
}
