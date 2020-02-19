package cmd

import (
	"context"
	"fmt"
	"os"
	"syscall"
	"time"

	"cloud.google.com/go/storage"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/ssh/terminal"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "tresor",
	Short: "Tresor is a tool to manage asymmetric client-side encryption for GCS.",
	Long:  `Tresor is a tool to manage asymmetric client-side encryption for GCS.`,
}

// Execute for root CMD
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

func loadArmoredKey(location string) (key *openpgp.Entity, err error) {
	file, err := os.Open(location)
	if err != nil {
		return nil, fmt.Errorf("failed to read key: %v", err)
	}
	defer file.Close()

	list, err := openpgp.ReadArmoredKeyRing(file)
	if err != nil {
		return nil, fmt.Errorf("failed to load keyring: %v", err)
	}

	return list[0], nil
}

func callbackForPassword(keys []openpgp.Key, symmetric bool) ([]byte, error) {
	if symmetric {
		return nil, fmt.Errorf("asked for symmetric key")
	}

	if len(keys) > 1 {
		return nil, fmt.Errorf("too many keys received")
	}

	fmt.Print("Enter Password: ")
	passwordBytes, err := terminal.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return nil, fmt.Errorf("failed to get user password: %v", err)
	}
	if len(keys) == 1 && keys[0].PrivateKey != nil {
		keys[0].PrivateKey.Decrypt(passwordBytes)
	}
	return passwordBytes, nil
}

func readMetadata(bucketName string, key string) (attributes *storage.ObjectAttrs, err error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage client: %v", err)
	}

	bucket := client.Bucket(bucketName)

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	object := bucket.Object(key)
	attrs, err := object.Attrs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve object metadata: %v", err)
	}
	return attrs, err
}
