package storage

import (
	"context"
	"fmt"
	"io"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// R2Client wraps the AWS SDK S3 client configured for Cloudflare R2.
type R2Client struct {
	client    *s3.Client
	bucket    string
	publicURL string
}

// NewR2Client creates a new R2-compatible S3 client.
// Cloudflare R2 uses the S3 API with a custom endpoint.
func NewR2Client(accountID, accessKeyID, secretAccessKey, bucket, publicURL string) (*R2Client, error) {
	endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID)

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			accessKeyID, secretAccessKey, "",
		)),
		config.WithRegion("auto"),
	)
	if err != nil {
		return nil, fmt.Errorf("load R2 config: %w", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true
	})

	return &R2Client{
		client:    client,
		bucket:    bucket,
		publicURL: publicURL,
	}, nil
}

// UploadAvatar uploads a member avatar image to R2 and returns the public URL.
// key should be of the form "avatars/{member_id}.jpg"
func (r *R2Client) UploadAvatar(ctx context.Context, memberID string, reader io.Reader, contentType string) (string, error) {
	ext := ".jpg"
	if contentType == "image/png" {
		ext = ".png"
	} else if contentType == "image/webp" {
		ext = ".webp"
	}

	key := fmt.Sprintf("avatars/%s%s", memberID, ext)

	_, err := r.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(r.bucket),
		Key:         aws.String(key),
		Body:        reader,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("upload avatar to R2: %w", err)
	}

	return r.GetPublicURL(key), nil
}

// UploadResource uploads an arbitrary chapter resource file to R2.
func (r *R2Client) UploadResource(ctx context.Context, chapterID, filename string, reader io.Reader, contentType string) (string, error) {
	ext := filepath.Ext(filename)
	key := fmt.Sprintf("resources/%s/%s%s", chapterID, chapterID, ext)

	_, err := r.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(r.bucket),
		Key:         aws.String(key),
		Body:        reader,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("upload resource to R2: %w", err)
	}

	return r.GetPublicURL(key), nil
}

// GetPublicURL returns the public CDN URL for an object key.
func (r *R2Client) GetPublicURL(key string) string {
	return fmt.Sprintf("%s/%s", r.publicURL, key)
}

// Ping verifies R2 connectivity by checking the bucket exists. Used by /healthz.
func (r *R2Client) Ping(ctx context.Context) error {
	_, err := r.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(r.bucket),
	})
	if err != nil {
		return fmt.Errorf("R2 ping failed: %w", err)
	}
	return nil
}
