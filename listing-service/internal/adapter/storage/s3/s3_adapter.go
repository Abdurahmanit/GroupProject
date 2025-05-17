package s3

import (
	"bytes"
	"context"
	"log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type S3Storage struct {
	client *minio.Client
	bucket string
}

type Storage interface {
	Upload(ctx context.Context, fileName string, data []byte) (string, error)
}

func NewS3Storage(endpoint, accessKey, secretKey, bucket string) (*S3Storage, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false, // Use true for HTTPS
	})
	if err != nil {
		return nil, err
	}

	// Create bucket if it doesn't exist
	err = client.MakeBucket(context.Background(), bucket, minio.MakeBucketOptions{})
	if err != nil {
		exists, errBucketExists := client.BucketExists(context.Background(), bucket)
		if errBucketExists == nil && exists {
			log.Printf("Bucket %s already exists", bucket)
		} else {
			return nil, err
		}
	}

	return &S3Storage{client: client, bucket: bucket}, nil
}

func (s *S3Storage) Upload(ctx context.Context, fileName string, data []byte) (string, error) {
	_, err := s.client.PutObject(ctx, s.bucket, fileName, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{})
	if err != nil {
		return "", err
	}
	return "http://" + s.client.EndpointURL().String() + "/" + s.bucket + "/" + fileName, nil
}