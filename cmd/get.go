package cmd

import (
    "context"
    "fmt"
    "io/ioutil"
    "time"

    "cloud.google.com/go/storage"
    "github.com/spf13/cobra"
    "github.com/spf13/viper"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a remote object from storage and decrypt it.",
	Long: `Get a remote object from storage and decrypt it.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check for correct number of arguments
		if len(args) != 1 {
			fail(fmt.Errorf("no object key specified."))
		}
		//key := args[0]
        
        _ , err := loadKeyring(viper.Get("keyring").(string))
        if err != nil {
            fail(err)
        }

		// Read remote object
        // encryptedBytes, err = readObject(viper.Get("bucket").(string), key)
		// if err != nil {
		// 	fail(fmt.Errorf("failed to read remote object: %v", err))
        // }


        // TODO continue with metadata download, signature verification, payload decryption and output
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}

func readObject(bucketName string, key string) (payload []byte, err error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage client: %v", err)
	}

	bucket := client.Bucket(bucketName)

	ctx, cancel := context.WithTimeout(ctx, time.Second*300)
	defer cancel()

	object := bucket.Object(key)
	reader, err := object.NewReader(ctx)
    if err != nil {
        return nil, err
    }
    defer reader.Close()

    data, err := ioutil.ReadAll(reader)
    if err != nil {
        return nil, err
    }
    return data, nil
}
