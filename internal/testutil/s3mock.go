package testutil

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"edu-platform/internal/s3"
)

type s3Object struct {
	data        []byte
	contentType string
}

// S3Mock is an in-memory implementation of s3.Storage for tests.
type S3Mock struct {
	mu      sync.RWMutex
	objects map[string]s3Object
}

func (m *S3Mock) Upload(_ context.Context, key, contentType string, r io.Reader, _ int64) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.objects == nil {
		m.objects = make(map[string]s3Object)
	}
	m.objects[key] = s3Object{data: data, contentType: contentType}
	return nil
}

func (m *S3Mock) Download(_ context.Context, key string) (io.ReadCloser, s3.ObjectMeta, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	obj, ok := m.objects[key]
	if !ok {
		return nil, s3.ObjectMeta{}, fmt.Errorf("object not found: %s", key)
	}
	return io.NopCloser(bytes.NewReader(obj.data)), s3.ObjectMeta{
		ContentType:  obj.contentType,
		Size:         int64(len(obj.data)),
		LastModified: time.Now(),
	}, nil
}

func (m *S3Mock) Delete(_ context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.objects, key)
	return nil
}
