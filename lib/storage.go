package tresor

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"time"

	"cloud.google.com/go/storage"
	"golang.org/x/crypto/openpgp"
	"google.golang.org/api/iterator"
)

const (
	emptyMetadata = "null"
)

// QueryStorage queries the remote storage to find keys
func QueryStorage(bucketName string, prefixFilter string, versions bool) (attributes []*storage.ObjectAttrs, err error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage client: %v", err)
	}

	bucket := client.Bucket(bucketName)

	var query *storage.Query
	if prefixFilter != "" {
		query = &storage.Query{Prefix: prefixFilter, Versions: versions}
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
func ReadObject(bucketName string, key string, version int64) (payload []byte, err error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage client: %v", err)
	}

	bucket := client.Bucket(bucketName)

	ctx, cancel := context.WithTimeout(ctx, time.Second*300)
	defer cancel()

	object := bucket.Object(key)
	if version != 0 {
		object = object.Generation(version)
	}
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
func WriteMetadata(bucketName string, key string, meta storage.ObjectAttrsToUpdate) (err error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create storage client: %v", err)
	}
	bucket := client.Bucket(bucketName)
	object := bucket.Object(key)

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	if _, err := object.Update(ctx, meta); err != nil {
		return fmt.Errorf("failed to update metadata: %v", err)
	}
	return nil
}

// CreateMetadata create metadata to be stored along with GCS objects
func CreateMetadata(recipient *openpgp.Entity, signer *openpgp.Entity, extension string, armored bool) storage.ObjectAttrsToUpdate {
	signingKey := emptyMetadata

	if signer != nil {
		signingKey = signer.PrimaryKey.KeyIdString()
	}

	if extension == "" {
		extension = emptyMetadata
	}

	return storage.ObjectAttrsToUpdate{
		ContentType:     "application/pgp-encrypted",
		ContentEncoding: "",
		Metadata: map[string]string{
			"Signing-Key":    signingKey,
			"Encryption-Key": recipient.PrimaryKey.KeyIdString(),
			"File-Extension": extension,
			"ASCII-Armor":    strconv.FormatBool(armored),
		},
	}
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

// CopyObject copies a remote object to a different remote key
func CopyObject(bucketName string, sourceKey string, destinationKey string) (err error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create storage client: %v", err)
	}

	bucket := client.Bucket(bucketName)

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	source := bucket.Object(sourceKey)
	destination := bucket.Object(destinationKey)

	if _, err := destination.CopierFrom(source).Run(ctx); err != nil {
		return fmt.Errorf("failed copy remote objects: %v", err)
	}
	return nil
}

// CopyMetadata copies custom meta data from a remote object to another
func CopyMetadata(bucketName string, sourceKey string, destinationKey string) error {
	metadata, err := ReadMetadata(bucketName, sourceKey)
	if err != nil {
		return fmt.Errorf("failed to read metadata: %v", err)
	}

	metaUpdate := storage.ObjectAttrsToUpdate{
		ContentType:     "application/pgp-encrypted",
		ContentEncoding: "",
		Metadata:        metadata.Metadata,
	}

	return WriteMetadata(bucketName, destinationKey, metaUpdate)
}
