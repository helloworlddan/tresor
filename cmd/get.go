package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	tresor "github.com/helloworlddan/tresor/lib"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/crypto/openpgp"
)

var (
	localWritePath string
	objectVersion  int64
)

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

		recipient, err := tresor.LoadArmoredKey(viper.Get("private_key").(string))
		if err != nil {
			fail(err)
		}

		// Read remote object
		encryptedBytes, err := tresor.ReadObject(viper.Get("bucket").(string), key, objectVersion)
		if err != nil {
			fail(err)
		}

		// Decrypt data
		plainBytes, err := tresor.DecryptBytes(openpgp.EntityList{recipient}, encryptedBytes)
		if err != nil {
			fail(err)
		}

		// Dump to STDOUT if no file specified
		if localWritePath == "" {
			fmt.Printf("%s", string(plainBytes))
			fmt.Fprintln(os.Stderr) // Print newline to STDERR to get prompt break right
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
	getCmd.Flags().Int64VarP(&objectVersion, "version", "v", 0, "Version of the object to get.")
}
