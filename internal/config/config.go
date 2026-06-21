package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Account stores non-secret configuration for a Woodpecker server.
type Account struct {
	Alias          string `json:"alias"`
	Server         string `json:"server"`
	APIBase        string `json:"api_base"`
	Auth           string `json:"auth"`
	TLSSkipVerify  bool   `json:"tls_skip_verify"`
	TimeoutSeconds int    `json:"timeout_seconds"`
}

// DefaultAPIBase is used when an account does not set one.
const DefaultAPIBase = "/api"

// DefaultAuth is the only supported auth scheme in v1.
const DefaultAuth = "bearer-token"

// DefaultTimeoutSeconds is the HTTP client timeout.
const DefaultTimeoutSeconds = 30

var configDirOverride string

// SetConfigDir overrides the configuration directory. Used by tests.
func SetConfigDir(dir string) error {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	configDirOverride = dir
	return nil
}

// ResetConfigDir clears the override.
func ResetConfigDir() {
	configDirOverride = ""
}

// Dir returns the base configuration directory.
func Dir() string {
	if configDirOverride != "" {
		return configDirOverride
	}
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			// Fallback: this should not happen in practice.
			home = "."
		}
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "wpci")
}

// AccountsDir returns the directory holding account JSON files.
func AccountsDir() string {
	return filepath.Join(Dir(), "accounts")
}

// TokensDir returns the directory holding bearer token files.
func TokensDir() string {
	return filepath.Join(Dir(), "tokens")
}

// accountPath returns the filesystem path for an alias.
func accountPath(alias string) string {
	return filepath.Join(AccountsDir(), alias+".json")
}

// NormalizeServer strips trailing slashes and whitespace from a server URL.
func NormalizeServer(server string) string {
	return strings.TrimRight(strings.TrimSpace(server), "/")
}

// NormalizeAccount fills in defaults and normalizes the server URL.
func NormalizeAccount(a *Account) {
	if a.APIBase == "" {
		a.APIBase = DefaultAPIBase
	}
	if a.Auth == "" {
		a.Auth = DefaultAuth
	}
	if a.TimeoutSeconds <= 0 {
		a.TimeoutSeconds = DefaultTimeoutSeconds
	}
	a.Server = NormalizeServer(a.Server)
}

// Save writes an account to disk, creating directories with the required
// permissions. The account file itself is written with 0600.
func (a Account) Save() error {
	NormalizeAccount(&a)

	accountsDir := AccountsDir()
	if err := os.MkdirAll(accountsDir, 0o700); err != nil {
		return fmt.Errorf("creating accounts directory: %w", err)
	}
	baseDir := Dir()
	if err := os.MkdirAll(baseDir, 0o700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	b, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding account: %w", err)
	}

	path := accountPath(a.Alias)
	if err := os.WriteFile(path, b, 0o600); err != nil {
		return fmt.Errorf("writing account file: %w", err)
	}
	return nil
}

// LoadAccount reads an account by alias.
func LoadAccount(alias string) (Account, error) {
	var a Account
	path := accountPath(alias)
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return a, fmt.Errorf("account %q not found", alias)
		}
		return a, fmt.Errorf("reading account: %w", err)
	}
	if err := json.Unmarshal(b, &a); err != nil {
		return a, fmt.Errorf("parsing account: %w", err)
	}
	NormalizeAccount(&a)
	return a, nil
}

// ListAccounts returns all configured account aliases, sorted.
func ListAccounts() ([]string, error) {
	entries, err := os.ReadDir(AccountsDir())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var aliases []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".json") {
			continue
		}
		aliases = append(aliases, strings.TrimSuffix(name, ".json"))
	}
	return aliases, nil
}

// RemoveAccount deletes the account config file.
func RemoveAccount(alias string) error {
	return os.Remove(accountPath(alias))
}

// TokenPath returns the filesystem path for an account token.
func TokenPath(alias string) string {
	return filepath.Join(TokensDir(), alias)
}

// SanitizeForDisplay returns a copy of the account safe for printing.
func (a Account) SanitizeForDisplay() Account {
	return a
}

// EnsureConfigDirs creates the base config and accounts/tokens directories with
// the permissions required by the PDR.
func EnsureConfigDirs() error {
	for _, d := range []string{Dir(), AccountsDir(), TokensDir()} {
		if err := os.MkdirAll(d, 0o700); err != nil {
			return err
		}
	}
	return nil
}

// UserAgent returns the CLI user agent string for HTTP requests.
func UserAgent() string {
	return "wpci/0.1.0"
}

// Platform returns a short OS/arch identifier.
func Platform() string {
	return fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
}
