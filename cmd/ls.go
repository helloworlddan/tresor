package cmd

import (
	"fmt"

	tresor "github.com/helloworlddan/tresor/lib"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List remote directory.",
	Long:  `List remote directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check for correct number of arguments
		prefixFilter := ""
		if len(args) == 1 {
			prefixFilter = args[0]
		}

		attrs, err := tresor.QueryStorage(viper.Get("bucket").(string), prefixFilter, false)
		if err != nil {
			fail(err)
		}

		for _, v := range attrs {
			fmt.Printf("%s\n", v.Name)
		}
	},
}

func init() {
	rootCmd.AddCommand(lsCmd)
}
