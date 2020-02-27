package cmd

import (
	"fmt"

	tresor "github.com/helloworlddan/tresor/lib"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cpCmd = &cobra.Command{
	Use:   "rm",
	Short: "Copy a remote object to a new key.",
	Long:  `Copy a remote object to a new key.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check for correct number of arguments
		if len(args) != 2 {
			fail(fmt.Errorf("specify to keys: source and destination"))
		}
		sourceKey := args[0]
		destinationKey := args[1]

		if err := tresor.CopyObject(viper.Get("bucket").(string), sourceKey, destinationKey); err != nil {
			fail(err)
		}

		if err := tresor.CopyMetadata(viper.Get("bucket").(string), sourceKey, destinationKey); err != nil {
			fail(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(cpCmd)
}
