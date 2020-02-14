package cmd

import (
    "bytes"
    "context"
    "fmt"
    "io"
    "io/ioutil"
    "time"

    "cloud.google.com/go/storage"
    "github.com/spf13/cobra"
    "github.com/spf13/viper"
)

var LocalPath string

// putCmd represents the put command
var putCmd = &cobra.Command{
	Use:   "put",
	Short: "Put a local object to remote storage and encrypt it.",
	Long: `Put a local object to remote storage and encrypt it.`,
	Run: func(cmd *cobra.Command, args []string) {
        if len(args) != 1 {
			fail(fmt.Errorf("no object key specified."))
        }
        key := args[0]

        transparentBytes, err := ioutil.ReadFile(LocalPath)
        if err != nil {
            fail(fmt.Errorf("failed to read local file: %v", LocalPath))
        }



        
        // Load keyring and encrypt byte array
        // reference https://snippets.cacher.io/snippet/1e07a360f65bb99ca0c5




        ctx := context.Background()
        client, err := storage.NewClient(ctx)
		if err != nil {
			fail(fmt.Errorf("failed to create storage client."))
        }
        bucket := client.Bucket(viper.Get("bucket").(string))

	    ctx, cancel := context.WithTimeout(ctx, time.Second*50)
        defer cancel()

        reader := bytes.NewReader(transparentBytes)
	    writer := bucket.Object(key).NewWriter(ctx)
	    if _, err = io.Copy(writer, reader); err != nil {
	    	fail(fmt.Errorf("failed to copy bytes to remote storage object."))
	    }
	    if err := writer.Close(); err != nil {
	    	fail(fmt.Errorf("failed to close write connection to remote storage."))
	    }
	},
}

func init() {
	rootCmd.AddCommand(putCmd)

	putCmd.Flags().StringVarP(&LocalPath, "file", "f", "", "Local file to read.")
}
