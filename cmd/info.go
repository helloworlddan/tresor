package cmd

import (
	"context"
	"encoding/hex"
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
		key := args[0]

		attrs, err := readMetadata(viper.Get("bucket").(string), key)
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
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}

func readMetadata(bucketName string, key string) (attributes *storage.ObjectAttrs, err error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage client: %v", err)
	}

	bucket := client.Bucket(bucketName)

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	object := bucket.Object(key)
	attrs, err := object.Attrs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve object metadata: %v", err)
	}
	return attrs, err
}
