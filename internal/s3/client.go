package s3

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// ObjectMeta holds metadata returned by Stat (HEAD).
type ObjectMeta struct {
	ContentType  string
	Size         int64
	LastModified time.Time
}

// Storage is the interface handlers use to interact with object storage.
// PDF bytes never flow through the backend: the frontend uploads/downloads
// directly to S3 via the presigned URLs we hand out.
type Storage interface {
	PresignPut(ctx context.Context, key, contentType string, expiry time.Duration) (string, error)
	PresignGet(ctx context.Context, key string, expiry time.Duration) (string, error)
	Stat(ctx context.Context, key string) (ObjectMeta, error)
	Delete(ctx context.Context, key string) error
}

type Client struct {
	mc     *minio.Client
	bucket string
}

func New(endpoint, accessKey, secretKey, bucket string, useSSL bool) (*Client, error) {
	mc, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("minio client: %w", err)
	}
	return &Client{mc: mc, bucket: bucket}, nil
}

// EnsureBucket creates the bucket if it does not already exist.
func (c *Client) EnsureBucket(ctx context.Context) error {
	exists, err := c.mc.BucketExists(ctx, c.bucket)
	if err != nil {
		return fmt.Errorf("bucket check: %w", err)
	}
	if !exists {
		if err := c.mc.MakeBucket(ctx, c.bucket, minio.MakeBucketOptions{}); err != nil {
			return fmt.Errorf("make bucket: %w", err)
		}
	}
	return nil
}

// PresignPut returns a temporary URL the client uses to PUT the object directly.
func (c *Client) PresignPut(ctx context.Context, key, contentType string, expiry time.Duration) (string, error) {
	u, err := c.mc.PresignedPutObject(ctx, c.bucket, key, expiry)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

// PresignGet returns a temporary URL the client uses to GET the object directly.
func (c *Client) PresignGet(ctx context.Context, key string, expiry time.Duration) (string, error) {
	u, err := c.mc.PresignedGetObject(ctx, c.bucket, key, expiry, url.Values{})
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

// Stat performs a HEAD to confirm the object exists and read its size/type.
func (c *Client) Stat(ctx context.Context, key string) (ObjectMeta, error) {
	info, err := c.mc.StatObject(ctx, c.bucket, key, minio.StatObjectOptions{})
	if err != nil {
		return ObjectMeta{}, err
	}
	return ObjectMeta{
		ContentType:  info.ContentType,
		Size:         info.Size,
		LastModified: info.LastModified,
	}, nil
}

func (c *Client) Delete(ctx context.Context, key string) error {
	return c.mc.RemoveObject(ctx, c.bucket, key, minio.RemoveObjectOptions{})
}
