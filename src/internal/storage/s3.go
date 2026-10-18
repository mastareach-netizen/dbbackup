package storage

import (
	"context"
	"fmt"
	"io"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3 struct {
	client *s3.Client
	bucket string
	prefix string
}

func NewS3(cfg S3Config) (*S3, error) {
	if cfg.Bucket == "" {
		return nil, fmt.Errorf("s3 storage requires a bucket (S3_BUCKET or --s3-bucket)")
	}
	loadOpts := []func(*config.LoadOptions) error{}
	if cfg.Region != "" {
		loadOpts = append(loadOpts, config.WithRegion(cfg.Region))
	}
	awsCfg, err := config.LoadDefaultConfig(context.Background(), loadOpts...)
	if err != nil {
		return nil, fmt.Errorf("load AWS config: %w", err)
	}
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
			o.UsePathStyle = true
		}
	})
	return &S3{
		client: client,
		bucket: cfg.Bucket,
		prefix: strings.Trim(cfg.Prefix, "/"),
	}, nil
}

func (s *S3) Name() string { return "s3" }

func (s *S3) objectKey(key string) string {
	if s.prefix == "" {
		return key
	}
	return path.Join(s.prefix, key)
}

func (s *S3) Upload(ctx context.Context, key string, r io.Reader) (string, error) {
	objKey := s.objectKey(key)
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objKey),
		Body:   r,
	})
	if err != nil {
		return "", fmt.Errorf("s3 upload: %w", err)
	}
	return fmt.Sprintf("s3://%s/%s", s.bucket, objKey), nil
}

func (s *S3) Download(ctx context.Context, key string, w io.Writer) error {
	objKey := s.objectKey(key)
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objKey),
	})
	if err != nil {
		return fmt.Errorf("s3 download: %w", err)
	}
	defer out.Body.Close()
	if _, err := io.Copy(w, out.Body); err != nil {
		return fmt.Errorf("read s3 object: %w", err)
	}
	return nil
}

func (s *S3) List(ctx context.Context, prefix string) ([]ObjectInfo, error) {
	objPrefix := s.objectKey(prefix)
	paginator := s3.NewListObjectsV2Paginator(s.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(objPrefix),
	})
	var result []ObjectInfo
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("s3 list: %w", err)
		}
		for _, obj := range page.Contents {
			key := aws.ToString(obj.Key)
			if s.prefix != "" {
				key = strings.TrimPrefix(key, s.prefix+"/")
			}
			var modified time.Time
			if obj.LastModified != nil {
				modified = *obj.LastModified
			}
			var size int64
			if obj.Size != nil {
				size = *obj.Size
			}
			result = append(result, ObjectInfo{Key: key, Size: size, Modified: modified})
		}
	}
	return result, nil
}

func (s *S3) Delete(ctx context.Context, key string) error {
	objKey := s.objectKey(key)
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objKey),
	})
	if err != nil {
		return fmt.Errorf("s3 delete: %w", err)
	}
	return nil
} 
