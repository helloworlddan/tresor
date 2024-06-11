package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const (
	version = "1.2.0"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information.",
	Long:  `Print version information.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("tresor: v%s\n", version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
