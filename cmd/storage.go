package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"golang.org/x/crypto/openpgp"
)

func readObject(bucketName string, key string) (payload []byte, err error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage client: %v", err)
	}

	bucket := client.Bucket(bucketName)

	ctx, cancel := context.WithTimeout(ctx, time.Second*300)
	defer cancel()

	object := bucket.Object(key)
	reader, err := object.NewReader(ctx)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return data, nil
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

func writeObject(bucketName string, key string, payload []byte) (err error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		fail(fmt.Errorf("failed to create storage client: %v", err))
	}
	bucket := client.Bucket(bucketName)

	ctx, cancel := context.WithTimeout(ctx, time.Second*300)
	defer cancel()

	reader := bytes.NewReader(payload)
	writer := bucket.Object(key).NewWriter(ctx)
	if _, err = io.Copy(writer, reader); err != nil {
		return fmt.Errorf("failed to copy bytes to remote storage object: %v", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close write connection to remote storage: %v", err)
	}

	return nil
}

func writeMetadata(bucketName string, key string, recipient *openpgp.Entity, signer *openpgp.Entity, extension string) (err error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create storage client: %v", err)
	}
	bucket := client.Bucket(bucketName)
	object := bucket.Object(key)

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	metadataAttrs := storage.ObjectAttrsToUpdate{
		ContentType:     "application/pgp-encrypted",
		ContentEncoding: "",
		Metadata: map[string]string{
			"Signing-Key":    strings.ToUpper(strconv.FormatUint(signer.PrimaryKey.KeyId, 16)),
			"Encryption-Key": strings.ToUpper(strconv.FormatUint(recipient.PrimaryKey.KeyId, 16)),
			"File-Extension": extension,
		},
	}

	if _, err := object.Update(ctx, metadataAttrs); err != nil {
		return fmt.Errorf("failed to update metadata: %v", err)
	}
	return nil
}
