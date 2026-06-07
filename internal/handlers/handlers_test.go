package handlers

import (
	"testing"
)

func TestAllowedExts(t *testing.T) {
	allowed := []string{".pdf"}
	for _, ext := range allowed {
		if !allowedExts[ext] {
			t.Errorf("expected %s to be allowed", ext)
		}
	}

	disallowed := []string{".exe", ".sh", ".js", ".docx", ".png"}
	for _, ext := range disallowed {
		if allowedExts[ext] {
			t.Errorf("expected %s to be disallowed", ext)
		}
	}
}
