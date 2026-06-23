package object

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
)

const testObjectPrefix = ".depotly/test-object.txt"

// PutTestObject uploads a small test object.
func PutTestObject(client *minio.Client, bucket string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	content := strings.NewReader(fmt.Sprintf("Depotly test object — %s\n", time.Now().Format(time.RFC3339)))

	_, err := client.PutObject(ctx, bucket, testObjectPrefix, content, -1,
		minio.PutObjectOptions{
			ContentType: "text/plain",
		})
	if err != nil {
		return fmt.Errorf("uploading test object: %w", err)
	}

	return nil
}

// ListObjects lists objects with a given prefix.
type ObjectInfo struct {
	Key          string
	Size         int64
	LastModified time.Time
}

func ListObjects(client *minio.Client, bucket, prefix string) ([]ObjectInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	}

	var objects []ObjectInfo
	for obj := range client.ListObjects(ctx, bucket, opts) {
		if obj.Err != nil {
			return nil, fmt.Errorf("listing objects: %w", obj.Err)
		}
		objects = append(objects, ObjectInfo{
			Key:          obj.Key,
			Size:         obj.Size,
			LastModified: obj.LastModified,
		})
	}

	return objects, nil
}

// SignedURL generates a presigned URL for an object.
func SignedURL(client *minio.Client, bucket, key string, expiry time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	url, err := client.PresignedGetObject(ctx, bucket, key, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("generating signed URL: %w", err)
	}

	return url.String(), nil
}

// CleanTestObjects deletes test objects created by Depotly.
func CleanTestObjects(client *minio.Client, bucket string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	objectsCh := make(chan minio.ObjectInfo)
	doneCh := make(chan struct{})

	go func() {
		defer close(doneCh)
		for obj := range client.ListObjects(ctx, bucket, minio.ListObjectsOptions{
			Prefix:    testObjectPrefix,
			Recursive: true,
		}) {
			if obj.Err != nil {
				return
			}
			objectsCh <- obj
		}
	}()

	// Read all objects first
	var toDelete []string
	for obj := range objectsCh {
		toDelete = append(toDelete, obj.Key)
	}
	<-doneCh
	close(objectsCh)

	// Quick close: just try to delete
	if len(toDelete) == 0 {
		return 0, nil
	}

	var deleted int
	for _, key := range toDelete {
		opts := minio.RemoveObjectOptions{}
		err := client.RemoveObject(ctx, bucket, key, opts)
		if err == nil {
			deleted++
		}
	}

	_ = os.Remove(testObjectPrefix)
	return deleted, nil
}
