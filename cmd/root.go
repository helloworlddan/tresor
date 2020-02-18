package cmd

import (
	"fmt"
	"os"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/crypto/openpgp"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "tresor",
	Short: "Tresor is a tool to manage asymmetric client-side encryption for GCS.",
	Long:  `Tresor is a tool to manage asymmetric client-side encryption for GCS.`,
}

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
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		viper.AddConfigPath(home)
		viper.SetConfigName(".tresor")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		fail(fmt.Errorf("failed to read config: %v", viper.ConfigFileUsed()))
	}
}

func fail(err error) {
	fmt.Printf("error: %v\n", err)
	os.Exit(1)
}

func loadKeyring(location string) (ring openpgp.EntityList, err error) {
	file, err := os.Open(location)
	if err != nil {
		return nil, fmt.Errorf("failed to read keyring: %v", err)
	}
	defer file.Close()

	list, err := openpgp.ReadKeyRing(file)
	if err != nil {
		return nil, fmt.Errorf("failed to load keyring: %v", err)
	}

	return list, nil
}

func getKey(ring openpgp.EntityList, identity string) (key *openpgp.Entity, err error) {
	for _, v := range ring {
		for k, _ := range v.Identities {
			if strings.Contains(k, identity) {
				return v, nil
			}
		}
	}
	return nil, fmt.Errorf("identity not found in keyring.")
}
