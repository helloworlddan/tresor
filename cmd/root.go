package cmd

import (
	"fmt"
	"os"
    
	"github.com/spf13/cobra"
    "golang.org/x/crypto/openpgp"
    "golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/packet"
	"github.com/spf13/viper"
	homedir "github.com/mitchellh/go-homedir"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "tresor",
	Short: "Tresor is a tool to manage asymmetric client-side encryption for GCS.",
	Long: `Tresor is a tool to manage asymmetric client-side encryption for GCS.`,
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
func loadKey(location string) (key *openpgp.Entity, err error){
    file, err := os.Open(location)
    if err != nil {
        return nil, fmt.Errorf("failed to read key: %v", err)
    }
    defer file.Close()

    armored, err := armor.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode ASCII armor for key: %v", err)
	}

    entity, err := openpgp.ReadEntity(packet.NewReader(armored.Body))
    if err != nil {
        return nil, fmt.Errorf("failed to load key: %v", err)
    }

    return entity, nil
}
