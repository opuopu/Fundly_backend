package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDotEnv_InlineComment(t *testing.T) {
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")
	content := `
TEST_KEY=secret123 # inline comment
QUOTED_KEY="hello # world"
`
	if err := os.WriteFile(envPath, []byte(content), 0600); err != nil {
		t.Fatalf("failed to write test .env: %v", err)
	}

	loadDotEnv(envPath)

	if got := os.Getenv("TEST_KEY"); got != "secret123" {
		t.Errorf("expected 'secret123', got %q", got)
	}
	if got := os.Getenv("QUOTED_KEY"); got != "hello # world" {
		t.Errorf("expected 'hello # world', got %q", got)
	}
}
