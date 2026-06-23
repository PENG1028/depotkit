package object

import (
	"context"
	"fmt"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Connect creates a new MinIO/S3 client.
func Connect(endpoint, accessKey, secretKey string, useSSL bool) (*minio.Client, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("creating S3 client: %w", err)
	}
	return client, nil
}

// EnsureBucket checks if a bucket exists and creates it if missing.
func EnsureBucket(ctx context.Context, client *minio.Client, bucket string) (bool, error) {
	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return false, fmt.Errorf("checking bucket existence: %w", err)
	}

	if !exists {
		err = client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
		if err != nil {
			return false, fmt.Errorf("creating bucket %s: %w", bucket, err)
		}
		return true, nil
	}

	return true, nil
}

// CheckStatus checks the object storage endpoint and bucket.
type StatusResult struct {
	Endpoint   string
	Bucket     string
	BucketOK   bool
	Writable   bool
}

func CheckStatus(endpoint, accessKey, secretKey, bucket string) (StatusResult, error) {
	client, err := Connect(endpoint, accessKey, secretKey, false)
	if err != nil {
		return StatusResult{}, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	bucketOK, err := EnsureBucket(ctx, client, bucket)
	if err != nil {
		return StatusResult{
			Endpoint: endpoint,
			Bucket:   bucket,
			BucketOK: false,
		}, nil // non-fatal — report status even if bucket check fails
	}

	return StatusResult{
		Endpoint: endpoint,
		Bucket:   bucket,
		BucketOK: bucketOK,
	}, nil
}
