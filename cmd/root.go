package cmd

import (
	"fmt"
	"os"
    
	"github.com/spf13/cobra"
    "golang.org/x/crypto/openpgp"
    "golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/packet"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "tresor",
	Short: "Tresor is a tool to manage asymmetric client-side encryption for GCS.",
	Long: `Tresor is a tool to manage asymmetric client-side encryption for GCS.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
    rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.tresor.yaml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".tresor" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".tresor")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
        fail(fmt.Errorf("failed to read config: %v", viper.ConfigFileUsed()))
    }
}

func fail(err error) {
	fmt.Printf("error: %v\n", err)
	os.Exit(1)
}

func loadKey(location string) (key *openpgp.Entity, err error){
    file, err := os.Open(location)
    if err != nil {
        return nil, fmt.Errorf("failed to read key: %v", location)
    }
    defer file.Close()

    armored, err := armor.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode ASCII armor for key: %v", location)
	}

    entity, err := openpgp.ReadEntity(packet.NewReader(armored.Body))
    if err != nil {
        return nil, fmt.Errorf("failed to load key: %v", location)
    }

    return entity, nil
}

