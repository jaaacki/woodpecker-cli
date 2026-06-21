package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNormalizeServer(t *testing.T) {
	if got := NormalizeServer("https://ci.example.com/"); got != "https://ci.example.com" {
		t.Fatalf("unexpected normalized server: %s", got)
	}
	if got := NormalizeServer("  https://ci.example.com//  "); got != "https://ci.example.com" {
		t.Fatalf("unexpected normalized server: %s", got)
	}
}

func TestNormalizeAccountDefaults(t *testing.T) {
	a := Account{Alias: "test", Server: "https://ci.example.com"}
	NormalizeAccount(&a)
	if a.APIBase != DefaultAPIBase {
		t.Fatalf("api_base default mismatch: %s", a.APIBase)
	}
	if a.Auth != DefaultAuth {
		t.Fatalf("auth default mismatch: %s", a.Auth)
	}
	if a.TimeoutSeconds != DefaultTimeoutSeconds {
		t.Fatalf("timeout default mismatch: %d", a.TimeoutSeconds)
	}
}

func TestSaveAndLoadAccount(t *testing.T) {
	dir := t.TempDir()
	accountsDir := filepath.Join(dir, "accounts")
	if err := os.MkdirAll(accountsDir, 0o700); err != nil {
		t.Fatal(err)
	}

	// Temporarily override the config directory via XDG_CONFIG_HOME.
	t.Setenv("XDG_CONFIG_HOME", dir)

	a := Account{
		Alias:          "test",
		Server:         "https://ci.example.com/",
		APIBase:        "/api",
		Auth:           "bearer-token",
		TimeoutSeconds: 15,
	}
	if err := a.Save(); err != nil {
		t.Fatal(err)
	}

	aliases, err := ListAccounts()
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, alias := range aliases {
		if alias == "test" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected test account in list, got %v", aliases)
	}

	loaded, err := LoadAccount("test")
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Server != "https://ci.example.com" {
		t.Fatalf("server not normalized in loaded account: %s", loaded.Server)
	}
	if loaded.TimeoutSeconds != 15 {
		t.Fatalf("timeout mismatch: %d", loaded.TimeoutSeconds)
	}

	if err := RemoveAccount("test"); err != nil {
		t.Fatal(err)
	}
	_, err = LoadAccount("test")
	if err == nil {
		t.Fatal("expected error loading removed account")
	}
}

func TestTokenPath(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	if got := TokenPath("home"); !filepath.IsAbs(got) {
		t.Fatalf("expected absolute token path, got %s", got)
	}
}
