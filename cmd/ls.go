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
        // Check for correct number of arguments
        prefixFilter := ""
        delimiter := ""
		if len(args) == 1 {
            prefixFilter = args[0]
            delimiter = "/"
        }
        
        attrs, err := queryStorage(viper.Get("bucket").(string), prefixFilter, delimiter)
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

func queryStorage(bucketName string, prefixFilter string, delimiter string) (attributes []*storage.ObjectAttrs, err error) {
    ctx := context.Background()
    client, err := storage.NewClient(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to create storage client: %v", err)
    }

    bucket := client.Bucket(bucketName)
    query := &storage.Query{
        Prefix: prefixFilter, 
        Delimiter: delimiter,
    }
    
    ctx, cancel := context.WithTimeout(ctx, time.Second*10)
    defer cancel()

    var attrs []*storage.ObjectAttrs

    it := bucket.Objects(ctx, query)
    for {
        attr, err := it.Next()
        if err == iterator.Done {
            break
        }
        if err != nil {
            return nil, fmt.Errorf("failed to read storage keys: %v", err)
        }
        attrs = append(attrs, attr)
    }
    return attrs, nil
}
