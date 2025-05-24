// internal/adapter/storage/s3/s3_adapter.go
package s3

import (
	"bytes"
	"context"
	// "log" // Заменим на кастомный логгер
	"fmt" // Для формирования URL и ошибок

	"github.com/Abdurahmanit/GroupProject/listing-service/internal/platform/logger" // <--- ИМПОРТ ТВОЕГО ЛОГГЕРА
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/google/uuid" // Для генерации уникальных имен файлов
	"path/filepath" // Для работы с расширениями файлов
)

type S3Storage struct {
	client *minio.Client
	bucket string
	logger *logger.Logger // <--- ДОБАВЛЯЕМ ПОЛЕ ДЛЯ ЛОГГЕРА
	// endpointURL string    // Можно сохранить endpoint для формирования URL
}

// Интерфейс Storage лучше определить в domain слое, как мы обсуждали,
// чтобы PhotoUsecase зависел от domain.Storage, а не от конкретной реализации s3.Storage.
// Если он здесь для удобства, то убедись, что он совпадает с domain.Storage.
/*
type Storage interface {
	Upload(ctx context.Context, fileName string, data []byte) (string, error)
}
*/

// NewS3Storage теперь принимает логгер
func NewS3Storage(endpoint, accessKey, secretKey, bucketName string, useSSL bool, log *logger.Logger) (*S3Storage, error) { // <--- ДОБАВЛЕН log *logger.Logger и useSSL
	log.Info("Initializing S3 MinIO Storage", "endpoint", endpoint, "bucket", bucketName, "use_ssl", useSSL)

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL, // Используем параметр useSSL
	})
	if err != nil {
		log.Error("S3Storage: failed to create MinIO client", "endpoint", endpoint, "error", err)
		return nil, fmt.Errorf("failed to create minio client for endpoint %s: %w", endpoint, err)
	}

	// Проверяем соединение (опционально, но полезно)
	// Например, можно попробовать ListBuckets или что-то подобное, но это может требовать доп. прав
	// log.Debug("S3Storage: MinIO client created, checking bucket existence...")

	// Create bucket if it doesn't exist
	err = client.MakeBucket(context.Background(), bucketName, minio.MakeBucketOptions{})
	if err != nil {
		exists, errBucketExists := client.BucketExists(context.Background(), bucketName)
		if errBucketExists == nil && exists {
			log.Info("S3Storage: Bucket already exists", "bucket", bucketName)
		} else {
			log.Error("S3Storage: failed to make or verify bucket", "bucket", bucketName, "make_bucket_error", err, "check_exists_error", errBucketExists)
			// Если errBucketExists не nil, то исходная ошибка err от MakeBucket более релевантна
			return nil, fmt.Errorf("failed to make/verify bucket %s: (make: %v / exists_check: %v)", bucketName, err, errBucketExists)
		}
	} else {
		log.Info("S3Storage: Bucket created successfully or already existed implicitly", "bucket", bucketName)
	}

	return &S3Storage{
		client: client,
		bucket: bucketName,
		logger: log, // <--- СОХРАНЯЕМ ЛОГГЕР
		// endpointURL: client.EndpointURL().String(), // Можно сохранить, если нужно для URL
	}, nil
}

func (s *S3Storage) Upload(ctx context.Context, originalFileName string, data []byte) (string, error) {
	// Генерируем уникальное имя файла, сохраняя расширение
	ext := filepath.Ext(originalFileName)
	objectKey := fmt.Sprintf("photos/%s%s", uuid.New().String(), ext) // Пример: photos/uuid.ext

	s.logger.Info("S3Storage.Upload: attempting to upload file",
		"bucket", s.bucket,
		"object_key", objectKey,
		"original_filename", originalFileName,
		"size_bytes", len(data))

	uploadInfo, err := s.client.PutObject(ctx, s.bucket, objectKey, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{
		// ContentType можно установить, если известен, например:
		// ContentType: http.DetectContentType(data),
		// UserMetadata: map[string]string{"original-filename": originalFileName},
	})
	if err != nil {
		s.logger.Error("S3Storage.Upload: PutObject failed", "bucket", s.bucket, "key", objectKey, "error", err)
		return "", fmt.Errorf("failed to upload object %s to bucket %s: %w", objectKey, s.bucket, err)
	}

	s.logger.Info("S3Storage.Upload: file uploaded successfully",
		"bucket", uploadInfo.Bucket,
		"key", uploadInfo.Key,
		"etag", uploadInfo.ETag,
		"size_uploaded", uploadInfo.Size)

	// Формирование URL для MinIO: http(s)://<endpoint>/<bucket>/<objectKey>
	// client.EndpointURL() возвращает URL с протоколом, который был использован при создании клиента.
	// Если endpoint был "minio.example.com:9000", а Secure=false, то EndpointURL() вернет "http://minio.example.com:9000"
	// Если Secure=true, то "https://minio.example.com:9000"
	fileURL := fmt.Sprintf("%s/%s/%s", s.client.EndpointURL().String(), s.bucket, objectKey)

	s.logger.Info("S3Storage.Upload: generated file URL", "url", fileURL)
	return fileURL, nil
}