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
    "golang.org/x/crypto/openpgp"
)

var LocalPath string

// putCmd represents the put command
var putCmd = &cobra.Command{
	Use:   "put",
	Short: "Encrypt a local object and put it to remote storage.",
	Long: `Encrypt a local object and put it to remote storage.`,
	Run: func(cmd *cobra.Command, args []string) {
        if len(args) != 1 {
			fail(fmt.Errorf("no object key specified."))
        }
        key := args[0]

        // Read local file
        plainBytes, err := ioutil.ReadFile(LocalPath)
        if err != nil {
            fail(fmt.Errorf("failed to read local file: %v", LocalPath))
        }

        // Load public keys for encryption
        recipient, err := loadKey(viper.Get("public_file").(string))
        if err != nil {
            fail(err)
        }

        // Load private key for signature
        signer, err := loadKey(viper.Get("private_file").(string))
        if err != nil {
            fail(err)
        }

        // Encrypt and sign
        encryptedBytes, err := encryptBytes(recipient, signer, plainBytes)
        if err != nil {
            fail(err)
        }

        // Write to storage
        if err = writeObject(viper.Get("bucket").(string), key, encryptedBytes); err != nil {
            fail(err)
        }
	},
}

func init() {
	rootCmd.AddCommand(putCmd)
	putCmd.Flags().StringVarP(&LocalPath, "file", "f", "", "Local file to read.")
}

func writeObject(bucketName string, key string, payload []byte) (err error){
    ctx := context.Background()
    client, err := storage.NewClient(ctx)
    if err != nil {
        fail(fmt.Errorf("failed to create storage client."))
    }
    bucket := client.Bucket(bucketName)
    
    ctx, cancel := context.WithTimeout(ctx, time.Second*300)
    defer cancel()
    
    reader := bytes.NewReader(payload)
    writer := bucket.Object(key).NewWriter(ctx)
    if _, err = io.Copy(writer, reader); err != nil {
        return fmt.Errorf("failed to copy bytes to remote storage object.")
    }
    if err := writer.Close(); err != nil {
        return fmt.Errorf("failed to close write connection to remote storage.")
    }

    return nil
}

func encryptBytes(recipient *openpgp.Entity, signer *openpgp.Entity, plainBytes []byte) (encryptedBytes []byte, err error) {
    recipients := make([]*openpgp.Entity, 1)
    recipients[0] = recipient

    cryptoBuffer := new(bytes.Buffer)
    cryptoWriter, err := openpgp.Encrypt(cryptoBuffer, recipients, nil, nil, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to close stream writer.")
    }

    if _, err = cryptoWriter.Write(plainBytes); err != nil {
        return nil, fmt.Errorf("failed to close stream writer.")
    }
    if err = cryptoWriter.Close(); err != nil {
        return nil, fmt.Errorf("failed to close stream writer.")
    }
    return cryptoBuffer.Bytes(), nil
}
