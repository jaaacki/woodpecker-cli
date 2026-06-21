package config

import (
	"os"
	"testing"
)

func TestEnvSuffix(t *testing.T) {
	cases := map[string]string{
		"home":    "HOME",
		"lab-001": "LAB_001",
		"my.org":  "MY_ORG",
		"a.b-c":   "A_B_C",
		"":        "",
	}
	for alias, want := range cases {
		if got := EnvSuffix(alias); got != want {
			t.Errorf("EnvSuffix(%q) = %q, want %q", alias, got, want)
		}
	}
}

func TestEnvNames(t *testing.T) {
	if got := EnvServerName("home"); got != "WPCI_HOME_SERVER" {
		t.Errorf("EnvServerName(home) = %q", got)
	}
	if got := EnvTokenName("lab-001"); got != "WPCI_LAB_001_TOKEN" {
		t.Errorf("EnvTokenName(lab-001) = %q", got)
	}
}

func TestEnvAccounts(t *testing.T) {
	t.Setenv("WPCI_HOME_SERVER", "https://ci.example.com")
	t.Setenv("WPCI_HOME_TOKEN", "tok-home")
	t.Setenv("WPCI_LAB_001_SERVER", "https://lab.example.com")
	t.Setenv("WPCI_NOPE_TOKEN", "token-only-should-not-count") // no server
	t.Setenv("WPCI_EMPTY_SERVER", "")                          // empty server ignored
	t.Setenv("UNRELATED_VAR", "x")
	t.Setenv("WPCI_TRAILING_SERVER", "https://t.example.com")

	got := EnvAccounts()
	want := []string{"home", "lab_001", "trailing"}
	if len(got) != len(want) {
		t.Fatalf("EnvAccounts() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("EnvAccounts()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestResolveAccountStoredWins(t *testing.T) {
	t.Setenv("WPCI_RESOLVE_SERVER", "https://env.example.com")
	// No stored account under the test config dir; env must be used.
	dir := t.TempDir()
	t.Setenv("WPCI_CONFIG_DIR", dir)
	acct, err := ResolveAccount("resolve")
	if err != nil {
		t.Fatalf("ResolveAccount: %v", err)
	}
	if acct.Server != "https://env.example.com" {
		t.Errorf("server = %q, want https://env.example.com", acct.Server)
	}
	if acct.Alias != "resolve" {
		t.Errorf("alias = %q, want resolve", acct.Alias)
	}
}

func TestResolveAccountMissing(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("WPCI_CONFIG_DIR", dir)
	// Ensure no stray env server for this alias.
	os.Unsetenv("WPCI_DEFMISS_SERVER")
	if _, err := ResolveAccount("defmiss"); err == nil {
		t.Fatal("ResolveAccount: expected error for missing account, got nil")
	}
}
