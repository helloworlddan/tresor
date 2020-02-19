package tresor

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"time"

	"cloud.google.com/go/storage"
	"golang.org/x/crypto/openpgp"
	"google.golang.org/api/iterator"
)

// QueryStorage queries the remote storage to find keys
func QueryStorage(bucketName string, prefixFilter string) (attributes []*storage.ObjectAttrs, err error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage client: %v", err)
	}

	bucket := client.Bucket(bucketName)

	var query *storage.Query
	if prefixFilter != "" {
		query = &storage.Query{Prefix: prefixFilter}
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	var attrs []*storage.ObjectAttrs

	it := bucket.Objects(ctx, query)
	for {
		attr, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read storage keys: %v", err)
		}
		attrs = append(attrs, attr)
	}
	return attrs, nil
}

// ReadObject reads a remote object
func ReadObject(bucketName string, key string) (payload []byte, err error) {
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

// ReadMetadata reads remote metadata for an object
func ReadMetadata(bucketName string, key string) (attributes *storage.ObjectAttrs, err error) {
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

// WriteObject write a byte sequence to remote storage
func WriteObject(bucketName string, key string, payload []byte) (err error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create storage client: %v", err)
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

// WriteMetadata writes a set of tags on a remote object
func WriteMetadata(bucketName string, key string, recipient *openpgp.Entity, signer *openpgp.Entity, extension string) (err error) {
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
			"Signing-Key":    signer.PrimaryKey.KeyIdString(),
			"Encryption-Key": recipient.PrimaryKey.KeyIdString(),
			"File-Extension": extension,
		},
	}

	if _, err := object.Update(ctx, metadataAttrs); err != nil {
		return fmt.Errorf("failed to update metadata: %v", err)
	}
	return nil
}

// RemoveObject removes an object from remote storage
func RemoveObject(bucketName string, key string) (err error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create storage client: %v", err)
	}

	bucket := client.Bucket(bucketName)

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	object := bucket.Object(key)
	if err = object.Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete object: %v", err)
	}
	return nil
}
