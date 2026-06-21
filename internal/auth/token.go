package auth

import (
	"fmt"
	"os"
	"strings"

	"github.com/jaaacki/woodpecker-cli/internal/config"
)

// Token provides load/save helpers for bearer tokens stored outside the account
// JSON files.
type Token struct {
	Alias string
}

// NewToken returns a token handle for the alias.
func NewToken(alias string) Token {
	return Token{Alias: alias}
}

// path returns the token file path.
func (t Token) path() string {
	return config.TokenPath(t.Alias)
}

// Load reads the token from disk, trimming whitespace.
func (t Token) Load() (string, error) {
	b, err := os.ReadFile(t.path())
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("token for account %q not found", t.Alias)
		}
		return "", fmt.Errorf("reading token: %w", err)
	}
	return strings.TrimSpace(string(b)), nil
}

// Save writes the token file with 0600 permissions.
func (t Token) Save(value string) error {
	if err := os.MkdirAll(config.TokensDir(), 0o700); err != nil {
		return fmt.Errorf("creating tokens directory: %w", err)
	}
	value = strings.TrimSpace(value)
	if err := os.WriteFile(t.path(), []byte(value+"\n"), 0o600); err != nil {
		return fmt.Errorf("writing token: %w", err)
	}
	return nil
}

// Remove deletes the token file.
func (t Token) Remove() error {
	return os.Remove(t.path())
}

// Exists reports whether the token file exists.
func (t Token) Exists() bool {
	_, err := os.Stat(t.path())
	return err == nil
}

// Status returns present/missing and a redacted prefix for display.
func (t Token) Status() string {
	value, err := t.Load()
	if err != nil {
		return "missing"
	}
	if value == "" {
		return "present (empty)"
	}
	return fmt.Sprintf("present (%s...)", Redact(value, 4))
}

// Redact masks all but the requested prefix length of a secret.
func Redact(value string, prefixLen int) string {
	if prefixLen < 0 {
		prefixLen = 0
	}
	if len(value) <= prefixLen {
		return strings.Repeat("*", len(value))
	}
	return value[:prefixLen] + strings.Repeat("*", len(value)-prefixLen)
}

// String returns a redacted representation safe for logging.
func (t Token) String() string {
	return "Token(" + t.Alias + ", " + t.Status() + ")"
}
