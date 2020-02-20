package tresor

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"syscall"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/ssh/terminal"
)

// LoadArmoredKey loads an armored GPG keys from local disk
func LoadArmoredKey(location string) (key *openpgp.Entity, err error) {
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

// CallbackForPassword implements https://godoc.org/golang.org/x/crypto/openpgp#PromptFunction
func CallbackForPassword(keys []openpgp.Key, symmetric bool) ([]byte, error) {
	if symmetric {
		return nil, fmt.Errorf("asked for symmetric key")
	}

	if len(keys) != 1 {
		return nil, fmt.Errorf("too many keys received")
	}

	if keys[0].PrivateKey == nil {
		return nil, fmt.Errorf("no private key detected")
	}

	passwordBytes, err := GetUserPassword(keys[0].PrivateKey.KeyIdString())
	if err != nil {
		return nil, err
	}

	keys[0].PrivateKey.Decrypt(passwordBytes)

	return passwordBytes, nil
}

// GetUserPassword promtps for a user password to decrypt private keys
func GetUserPassword(keyID string) ([]byte, error) {
	fmt.Fprintf(os.Stderr, "Enter Password for key %s: ", keyID)
	passwordBytes, err := terminal.ReadPassword(int(syscall.Stdin))
	fmt.Fprintln(os.Stderr, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get user password: %v", err)
	}

	return passwordBytes, nil
}

// EncryptBytes encrypts and signs a byte sequence
func EncryptBytes(recipient *openpgp.Entity, signer *openpgp.Entity, plainBytes []byte, armored bool) (encryptedBytes []byte, err error) {
	if armored {
		return encryptArmored(recipient, signer, plainBytes)
	}
	return encryptBinary(recipient, signer, plainBytes)
}

func encryptBinary(recipient *openpgp.Entity, signer *openpgp.Entity, plainBytes []byte) ([]byte, error) {
	recipients := make([]*openpgp.Entity, 1)
	recipients[0] = recipient

	cryptoBuffer := bytes.NewBuffer(nil)

	cryptoWriter, err := openpgp.Encrypt(cryptoBuffer, recipients, signer, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open stream writer: %v", err)
	}
	if _, err = cryptoWriter.Write(plainBytes); err != nil {
		return nil, fmt.Errorf("failed to write stream: %v", err)
	}
	if err = cryptoWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close stream: %v", err)
	}
	return cryptoBuffer.Bytes(), nil
}

func encryptArmored(recipient *openpgp.Entity, signer *openpgp.Entity, plainBytes []byte) ([]byte, error) {
	recipients := make([]*openpgp.Entity, 1)
	recipients[0] = recipient

	cryptoBuffer := bytes.NewBuffer(nil)

	armorWriter, err := armor.Encode(cryptoBuffer, "Message", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open armor writer: %v", err)
	}
	cryptoWriter, err := openpgp.Encrypt(armorWriter, recipients, signer, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open stream writer: %v", err)
	}
	if _, err = cryptoWriter.Write(plainBytes); err != nil {
		return nil, fmt.Errorf("failed to write stream: %v", err)
	}
	if err = cryptoWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close stream: %v", err)
	}
	if err = armorWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to armor stream: %v", err)
	}
	return cryptoBuffer.Bytes(), nil
}

// DecryptBytes decrypts and verifies a byte sequence
func DecryptBytes(ring openpgp.EntityList, payload []byte) (plain []byte, err error) {
	// Attempt to find and decode ASCII armor
	var message *openpgp.MessageDetails

	armoredBlock, err := armor.Decode(bytes.NewReader(payload))
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to decode object: %v", err)
	}

	if armoredBlock != nil {
		message, err = openpgp.ReadMessage(armoredBlock.Body, ring, CallbackForPassword, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to read armored gpg message: %v", err)
		}
	} else {
		message, err = openpgp.ReadMessage(bytes.NewReader(payload), ring, CallbackForPassword, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to read gpg message: %v", err)
		}
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
