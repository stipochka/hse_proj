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
type Storage interface {
	PresignPut(ctx context.Context, key, contentType string, expiry time.Duration) (string, error)
	PresignGet(ctx context.Context, key string, expiry time.Duration) (string, error)
	Stat(ctx context.Context, key string) (ObjectMeta, error)
	Delete(ctx context.Context, key string) error
}

type Client struct {
	mc       *minio.Client // internal: Stat, Delete, EnsureBucket
	mcSign   *minio.Client // public-facing: presign only (signed with browser-reachable host)
	bucket   string
}

// New creates a MinIO client.
// endpoint     — internal Docker address (e.g. "minio:9000"), used for all API ops.
// publicURL    — browser-reachable address (e.g. "http://localhost:9000"), used only
//
//	for presigning so the resulting URL can be reached by the browser.
//	If empty, the same endpoint is used for presigning.
func New(endpoint, accessKey, secretKey, bucket string, useSSL bool, publicURL string) (*Client, error) {
	creds := credentials.NewStaticV4(accessKey, secretKey, "")

	// Region must be set so PresignedXxx skips the getBucketLocation network call.
	// MinIO defaults to us-east-1; the credential string in the URL confirms this.
	const region = "us-east-1"

	mc, err := minio.New(endpoint, &minio.Options{Creds: creds, Secure: useSSL, Region: region})
	if err != nil {
		return nil, fmt.Errorf("minio client: %w", err)
	}

	mcSign := mc // default: same client for presigning
	if publicURL != "" {
		u, err := url.Parse(publicURL)
		if err != nil {
			return nil, fmt.Errorf("parse S3_PUBLIC_URL: %w", err)
		}
		publicSecure := u.Scheme == "https"
		// mcSign uses the public host so presigned URLs are browser-reachable.
		// Region is pre-set to avoid any network call from within Docker.
		mcSign, err = minio.New(u.Host, &minio.Options{Creds: creds, Secure: publicSecure, Region: region})
		if err != nil {
			return nil, fmt.Errorf("minio public client: %w", err)
		}
	}

	return &Client{mc: mc, mcSign: mcSign, bucket: bucket}, nil
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

// PresignPut returns a presigned PUT URL signed for the public host.
func (c *Client) PresignPut(ctx context.Context, key, contentType string, expiry time.Duration) (string, error) {
	u, err := c.mcSign.PresignedPutObject(ctx, c.bucket, key, expiry)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

// PresignGet returns a presigned GET URL signed for the public host.
func (c *Client) PresignGet(ctx context.Context, key string, expiry time.Duration) (string, error) {
	u, err := c.mcSign.PresignedGetObject(ctx, c.bucket, key, expiry, url.Values{})
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

// Stat performs a HEAD to confirm the object exists.
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
