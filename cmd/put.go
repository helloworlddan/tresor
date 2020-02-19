package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/crypto/openpgp"
)

var localReadPath string

// putCmd represents the put command
var putCmd = &cobra.Command{
	Use:   "put",
	Short: "Encrypt a local object and put it to remote storage.",
	Long:  `Encrypt a local object and put it to remote storage.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fail(fmt.Errorf("no object key specified"))
		}
		key := args[0]

		// Read local file
		plainBytes, err := ioutil.ReadFile(localReadPath)
		if err != nil {
			fail(err)
		}

		// Load keys
		recipient, err := loadArmoredKey(viper.Get("public_key").(string))
		if err != nil {
			fail(err)
		}

		// Load private keys for signature
		signer, err := loadArmoredKey(viper.Get("private_key").(string))
		if err != nil {
			fail(err)
		}

		// Get password
		password, err := callbackForPassword([]openpgp.Key{}, false)
		if err != nil {
			fail(err)
		}

		// Decrypt private key
		signer.PrivateKey.Decrypt(password)

		// Encrypt and sign
		encryptedBytes, err := encryptBytes(recipient, signer, plainBytes)
		if err != nil {
			fail(err)
		}

		// Write to storage
		if err = writeObject(viper.Get("bucket").(string), key, encryptedBytes); err != nil {
			fail(err)
		}

		// Write metadata
		if err = writeMetadata(viper.Get("bucket").(string), key, recipient, signer, filepath.Ext(localReadPath)); err != nil {
			fail(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(putCmd)
	putCmd.Flags().StringVarP(&localReadPath, "in", "i", "", "Input file to read from.")
}

func writeObject(bucketName string, key string, payload []byte) (err error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		fail(fmt.Errorf("failed to create storage client: %v", err))
	}
	bucket := client.Bucket(bucketName)

	ctx, cancel := context.WithTimeout(ctx, time.Second*300)
	defer cancel()

	reader := bytes.NewReader(payload)
	writer := bucket.Object(key).NewWriter(ctx)
	if _, err = io.Copy(writer, reader); err != nil {
		return fmt.Errorf("failed to copy bytes to remote storage object: %v", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close write connection to remote storage: %v", err)
	}

	return nil
}

func writeMetadata(bucketName string, key string, recipient *openpgp.Entity, signer *openpgp.Entity, extension string) (err error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create storage client: %v", err)
	}
	bucket := client.Bucket(bucketName)
	object := bucket.Object(key)

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	metadataAttrs := storage.ObjectAttrsToUpdate{
		ContentType:     "application/pgp-encrypted",
		ContentEncoding: "",
		Metadata: map[string]string{
			"Signing-Key":    strings.ToUpper(strconv.FormatUint(signer.PrimaryKey.KeyId, 16)),
			"Encryption-Key": strings.ToUpper(strconv.FormatUint(recipient.PrimaryKey.KeyId, 16)),
			"File-Extension": extension,
		},
	}

	if _, err := object.Update(ctx, metadataAttrs); err != nil {
		return fmt.Errorf("failed to update metadata: %v", err)
	}
	return nil
}

func encryptBytes(recipient *openpgp.Entity, signer *openpgp.Entity, plainBytes []byte) (encryptedBytes []byte, err error) {
	recipients := make([]*openpgp.Entity, 1)
	recipients[0] = recipient

	cryptoBuffer := new(bytes.Buffer)
	cryptoWriter, err := openpgp.Encrypt(cryptoBuffer, recipients, signer, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open stream writer: %v", err)
	}
	if _, err = cryptoWriter.Write(plainBytes); err != nil {
		return nil, fmt.Errorf("failed to write stream: %v", err)
	}
	if err = cryptoWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close stream writer: %v", err)
	}
	return cryptoBuffer.Bytes(), nil
}
