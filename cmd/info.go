package cmd

import (
	"context"
	"fmt"
    "time"

    "cloud.google.com/go/storage"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// infoCmd represents the info command
var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Get info on remote objects.",
	Long:  `Get info on remote objects.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check for correct number of arguments
		if len(args) != 1 {
			fail(fmt.Errorf("no object key specified."))
        }
        
		ctx := context.Background()
		client, err := storage.NewClient(ctx)
		if err != nil {
			fail(fmt.Errorf("failed to create storage client."))
        }
        
        key := args[0]
        bucket := client.Bucket(viper.Get("bucket").(string))
        
		ctx, cancel := context.WithTimeout(ctx, time.Second*10)
        defer cancel()

        object := bucket.Object(key)
	    attrs, err := object.Attrs(ctx)
	    if err != nil {
	    	fail(fmt.Errorf("failed to retrieve object metadata for key: %s", object))
	    }

        fmt.Printf("Name\t%v\n", attrs.Name)
        fmt.Printf("Size\t%v bytes\n", attrs.Size)
        fmt.Printf("Created\t%v\n", attrs.Created)
        fmt.Printf("Updated\t%v\n", attrs.Updated)
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}

