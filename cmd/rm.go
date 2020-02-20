package cmd

import (
	"fmt"

	tresor "github.com/helloworlddan/tresor/lib"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rmCmd = &cobra.Command{
	Use:   "rm",
	Short: "Remove a remote object.",
	Long:  `Remove a remote object.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check for correct number of arguments
		if len(args) != 1 {
			fail(fmt.Errorf("no object key specified"))
		}
		key := args[0]

		if err := tresor.RemoveObject(viper.Get("bucket").(string), key); err != nil {
			fail(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(rmCmd)
}
