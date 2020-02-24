package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"

	tresor "github.com/helloworlddan/tresor/lib"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	localReadPath     string
	interactivePrompt bool
)

var putCmd = &cobra.Command{
	Use:   "put",
	Short: "Encrypt a local object and put it to remote storage.",
	Long:  `Encrypt a local object and put it to remote storage.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fail(fmt.Errorf("no object key specified"))
		}
		key := args[0]

		// Read input
		plainBytes, err := readInput(localReadPath, interactivePrompt, viper.Get("object_signing").(bool))
		if err != nil {
			fail(err)
		}

		// Load keys
		recipient, err := tresor.LoadArmoredKey(viper.Get("public_key").(string))
		if err != nil {
			fail(err)
		}

		var signer *openpgp.Entity

		// Sign object if configured
		if viper.Get("object_signing").(bool) {
			// Load private keys for signature
			signer, err = tresor.LoadArmoredKey(viper.Get("private_key").(string))
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
		}

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

func readInput(localPath string, interactive bool, objectSigning bool) ([]byte, error) {
	// Read local file if flag given
	if localPath != "" {
		return ioutil.ReadFile(localPath)
	}
	// Read interactive prompt
	if interactive {
		for {
			plainBytes, err := getSecret("Enter input to store: ")
			if err != nil {
				return nil, err
			}
			confirmBytes, err := getSecret("Confirm input to continue: ")
			if err != nil {
				return nil, err
			}
			if string(plainBytes) == string(confirmBytes) {
				return plainBytes, nil
			}
			fmt.Fprintln(os.Stderr, "Input does not match repetition.")
		}
	}
	// Read from STDIN
	if objectSigning {
		return nil, fmt.Errorf("refusing to read both password and payload from STDIN. Turn off 'object_signing' or supply input differently")
	}
	fmt.Fprintln(os.Stderr, "Reading from STDIN...")
	return ioutil.ReadAll(os.Stdin)
}

func getSecret(prompt string) ([]byte, error) {
	fmt.Fprintf(os.Stderr, "%s", prompt)
	plainBytes, err := terminal.ReadPassword(int(syscall.Stdin))
	fmt.Fprintln(os.Stderr, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get secret from prompt: %v", err)
	}
	return plainBytes, nil
}

func init() {
	rootCmd.AddCommand(putCmd)
	putCmd.Flags().StringVarP(&localReadPath, "in", "i", "", "Input file to read from.")
	putCmd.Flags().BoolVarP(&interactivePrompt, "prompt", "p", false, "Use an interactive prompt for input.")
}
