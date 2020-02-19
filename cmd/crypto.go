package cmd

import (
	"fmt"
	"os"
	"syscall"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/ssh/terminal"
)

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
