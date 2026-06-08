package testutil

import (
	"context"
	"fmt"
	"sync"
	"time"

	"edu-platform/internal/s3"
)

// S3Mock is an in-memory implementation of s3.Storage for tests.
// Because the presigned flow means the real PUT never reaches the backend,
// tests simulate a completed upload with SeedObject before calling confirm.
type S3Mock struct {
	mu      sync.RWMutex
	objects map[string]s3.ObjectMeta
}

// SeedObject marks a key as present (simulating a finished direct upload).
func (m *S3Mock) SeedObject(key string, size int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.objects == nil {
		m.objects = make(map[string]s3.ObjectMeta)
	}
	m.objects[key] = s3.ObjectMeta{ContentType: "application/pdf", Size: size, LastModified: time.Now()}
}

func (m *S3Mock) PresignPut(_ context.Context, key, _ string, _ time.Duration) (string, error) {
	return "https://s3.test/put/" + key, nil
}

func (m *S3Mock) PresignGet(_ context.Context, key string, _ time.Duration) (string, error) {
	return "https://s3.test/get/" + key, nil
}

func (m *S3Mock) Stat(_ context.Context, key string) (s3.ObjectMeta, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	meta, ok := m.objects[key]
	if !ok {
		return s3.ObjectMeta{}, fmt.Errorf("object not found: %s", key)
	}
	return meta, nil
}

func (m *S3Mock) Delete(_ context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.objects, key)
	return nil
}
