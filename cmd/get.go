package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/crypto/openpgp"
)

var localWritePath string

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a remote object from storage and decrypt it.",
	Long:  `Get a remote object from storage and decrypt it.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check for correct number of arguments
		if len(args) != 1 {
			fail(fmt.Errorf("no object key specified"))
		}
		key := args[0]

		recipient, err := loadArmoredKey(viper.Get("private_key").(string))
		if err != nil {
			fail(err)
		}

		// Read remote object
		encryptedBytes, err := readObject(viper.Get("bucket").(string), key)
		if err != nil {
			fail(err)
		}

		// Decrypt data
		plainBytes, err := decryptBytes(openpgp.EntityList{recipient}, encryptedBytes)
		if err != nil {
			fail(err)
		}

		if localWritePath == "" {
			fmt.Printf("%s\n", string(plainBytes))
			return
		}

		if err = ioutil.WriteFile(localWritePath, plainBytes, 0644); err != nil {
			fail(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
	getCmd.Flags().StringVarP(&localWritePath, "out", "o", "", "Output file to write to.")
}

func decryptBytes(ring openpgp.EntityList, payload []byte) (plain []byte, err error) {
	message, err := openpgp.ReadMessage(bytes.NewBuffer(payload), ring, callbackForPassword, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to read gpg message: %v", err)
	}

	bytes, err := ioutil.ReadAll(message.UnverifiedBody)
	if err != nil {
		return nil, fmt.Errorf("failed to read gpg data: %v", err)
	}

	if message.SignatureError != nil {
		return nil, message.SignatureError
	}

	return bytes, nil
}
