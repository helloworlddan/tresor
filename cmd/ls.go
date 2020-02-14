package cmd

import (
	"context"
	"fmt"
    "time"

    "cloud.google.com/go/storage"
    "google.golang.org/api/iterator"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// lsCmd represents the ls command
var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List remote directory.",
	Long:  `List remote directory.`,
	Run: func(cmd *cobra.Command, args []string) {
        query := &storage.Query{}
		// Check for correct number of arguments
		if len(args) == 1 {
			query = &storage.Query{Prefix: args[0], Delimiter: "/"}
        }
        
		ctx := context.Background()
		client, err := storage.NewClient(ctx)
		if err != nil {
			fail(fmt.Errorf("failed to create storage client."))
        }
        
        bucket := client.Bucket(viper.Get("bucket").(string))
        
		ctx, cancel := context.WithTimeout(ctx, time.Second*10)
        defer cancel()

		it := bucket.Objects(ctx, query)
		for {
			attrs, err := it.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				fail(fmt.Errorf("failed to read storage keys."))
			}
			fmt.Printf("%s\n", attrs.Name)
		}
	},
}

func init() {
	rootCmd.AddCommand(lsCmd)
}

