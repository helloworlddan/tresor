package cmd

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/storage"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// rmCmd represents the info command
var rmCmd = &cobra.Command{
	Use:   "rm",
	Short: "Remove a remote objects.",
	Long:  `Remove a remote objects.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check for correct number of arguments
		if len(args) != 1 {
			fail(fmt.Errorf("no object key specified."))
		}
		key := args[0]

		if err := removeObject(viper.Get("bucket").(string), key); err != nil {
			fail(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(rmCmd)
}

func removeObject(bucketName string, key string) (err error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create storage client: %v", err)
	}

	bucket := client.Bucket(bucketName)

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	object := bucket.Object(key)
	if err = object.Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete object: %v", err)
	}
	return nil
}
