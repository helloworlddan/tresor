package cmd

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	tresor "github.com/helloworlddan/tresor/lib"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var localReadPath string

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
		recipient, err := tresor.LoadArmoredKey(viper.Get("public_key").(string))
		if err != nil {
			fail(err)
		}

		// Load private keys for signature
		signer, err := tresor.LoadArmoredKey(viper.Get("private_key").(string))
		if err != nil {
			fail(err)
		}

		// Get password
		password, err := tresor.GetUserPassword(signer.PrivateKey.KeyIdString())
		if err != nil {
			fail(err)
		}

		// Decrypt private key
		signer.PrivateKey.Decrypt(password)

		// Encrypt and sign
		encryptedBytes, err := tresor.EncryptBytes(recipient, signer, plainBytes, viper.Get("ascii_armor").(bool))
		if err != nil {
			fail(err)
		}

		// Write to storage
		if err = tresor.WriteObject(viper.Get("bucket").(string), key, encryptedBytes); err != nil {
			fail(err)
		}

		// Write metadata
		if err = tresor.WriteMetadata(viper.Get("bucket").(string), key, recipient, signer, filepath.Ext(localReadPath), viper.Get("ascii_armor").(bool)); err != nil {
			fail(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(putCmd)
	putCmd.Flags().StringVarP(&localReadPath, "in", "i", "", "Input file to read from.")
}
