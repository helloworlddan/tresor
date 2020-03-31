package cmd

import (
	"encoding/hex"
	"fmt"

	tresor "github.com/helloworlddan/tresor/lib"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Get info on remote objects.",
	Long:  `Get info on remote objects.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check for correct number of arguments
		if len(args) != 1 {
			fail(fmt.Errorf("no object key specified"))
		}
		key := args[0]

		attrs, err := tresor.ReadMetadata(viper.Get("bucket").(string), key)
		if err != nil {
			fail(err)
		}

		fmt.Printf("Name\t\t%v\n", attrs.Name)
		fmt.Printf("Size\t\t%v bytes\n", attrs.Size)
		fmt.Printf("MD5\t\t%v\n", hex.EncodeToString(attrs.MD5))
		fmt.Printf("Type\t\t%v\n", attrs.ContentType)
		fmt.Printf("Modified\t%v\n", attrs.Updated)

		for k, v := range attrs.Metadata {
			fmt.Printf("%v\t%v\n", k, v)
		}

		versions, err := tresor.QueryStorage(viper.Get("bucket").(string), key, true)
		if err != nil {
			fail(err)
		}

		if len(versions) > 1 {
			// Reverse versions
			for i, j := 0, len(versions)-1; i < j; i, j = i+1, j-1 {
				versions[i], versions[j] = versions[j], versions[i]
			}
			fmt.Println()
			for _, v := range versions {
				fmt.Printf("Version\t\t%v\n- Modified\t%v\n", v.Generation, v.Updated.String())
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
