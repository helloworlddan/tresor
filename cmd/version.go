package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const (
    version = "0.0.1"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information.",
	Long: `Print version information.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("tresor: v%s\n", version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
