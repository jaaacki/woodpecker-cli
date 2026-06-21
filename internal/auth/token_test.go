package auth

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTokenSaveLoadRemove(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	tok := NewToken("home")
	if err := tok.Save("secret-token"); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(filepath.Join(dir, "wpci", "tokens", "home"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("token file permissions should be 0600, got %o", info.Mode().Perm())
	}

	value, err := tok.Load()
	if err != nil {
		t.Fatal(err)
	}
	if value != "secret-token" {
		t.Fatalf("token value mismatch: %s", value)
	}

	status := tok.Status()
	if status == "missing" {
		t.Fatal("token should be present")
	}

	if err := tok.Remove(); err != nil {
		t.Fatal(err)
	}
	if tok.Exists() {
		t.Fatal("token should not exist after removal")
	}
}

func TestRedact(t *testing.T) {
	if got := Redact("abcd", 2); got != "ab**" {
		t.Fatalf("redaction mismatch: %s", got)
	}
	if got := Redact("a", 2); got != "*" {
		t.Fatalf("short redaction mismatch: %s", got)
	}
	if got := Redact("", 2); got != "" {
		t.Fatalf("empty redaction mismatch: %s", got)
	}
}
