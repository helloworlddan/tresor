package cmd

import (
    "bytes"
    "context"
    "fmt"
    "io/ioutil"
    "time"
    "syscall"

    "cloud.google.com/go/storage"
    "github.com/spf13/cobra"
    "github.com/spf13/viper"
    "golang.org/x/crypto/openpgp"
    "golang.org/x/crypto/ssh/terminal"
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
		key := args[0]
        
        ring, err := loadKeyring(viper.Get("keyring").(string))
        if err != nil {
            fail(err)
        }

		// Read remote object
        encryptedBytes, err := readObject(viper.Get("bucket").(string), key)
		if err != nil {
			fail(fmt.Errorf("failed to read remote object: %v", err))
        }

        // Decrypt data
        plainBytes, err := decryptBytes(ring, encryptedBytes)
		if err != nil {
			fail(fmt.Errorf("failed to decrypt data: %v", err))
        }

        fmt.Printf("read message: %s", string(plainBytes))

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

func decryptBytes(ring openpgp.EntityList, payload []byte) (plain []byte, err error) {
    // key, err := getKey(ring, viper.Get("identity").(string))
    // if err != nil {
    //     return nil, fmt.Errorf("failed to get gpg key: %v", err)
    // }

    // if err = key.PrivateKey.Decrypt(password); err != nil {
	// 	return nil, fmt.Errorf("failed to decrypt private key: %v", err)
	// }
    // for _, subkey := range key.Subkeys {
    //     if err = subkey.PrivateKey.Decrypt(password); err != nil {
	// 	    return nil, fmt.Errorf("failed to decrypt private subkey: %v", err)
	//     }
    // }

    message, err := openpgp.ReadMessage(bytes.NewBuffer(payload), ring, calbackForPassword, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to read gpg message: %v", err)
    }
    bytes, err := ioutil.ReadAll(message.UnverifiedBody)
    if err != nil {
        return nil, fmt.Errorf("failed to read gpg data: %v", err)
    }
    return bytes, nil
}

func calbackForPassword(keys []openpgp.Key, symmetric bool) ([]byte, error){
    fmt.Print("Enter Password: ")
    bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
    fmt.Println()
    if err != nil {
        return nil, fmt.Errorf("failed to get user password: %v", err)
    }
    return bytePassword, nil
}