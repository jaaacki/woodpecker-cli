package config

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

// Env suffix convention: an account alias resolves to environment variables
//   WPCI_<SUFFIX>_SERVER   (server base URL)
//   WPCI_<SUFFIX>_TOKEN    (bearer token)
// where <SUFFIX> is the alias uppercased with every non-[A-Z0-9] rune replaced
// by '_'. This lets a shell define an account without running `wpci account add`:
//
//	export WPCI_HOME_SERVER=https://ci.example.com
//	export WPCI_HOME_TOKEN=ghp_xxx
//	alias wpci-home='wpci home'
//	wpci-home repo ls
//
// enabling zero-config multi-account use from zsh aliases.

const envPrefix = "WPCI_"

// EnvSuffix converts an account alias to the env-var suffix used by the
// WPCI_<ALIAS>_SERVER / WPCI_<ALIAS>_TOKEN convention.
func EnvSuffix(alias string) string {
	var b strings.Builder
	for _, r := range strings.ToUpper(alias) {
		switch {
		case r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	return b.String()
}

// EnvServerName returns the env var holding an account's server URL.
func EnvServerName(alias string) string { return envPrefix + EnvSuffix(alias) + "_SERVER" }

// EnvTokenName returns the env var holding an account's bearer token.
func EnvTokenName(alias string) string { return envPrefix + EnvSuffix(alias) + "_TOKEN" }

// EnvAccounts returns aliases configured only via the WPCI_<ALIAS>_SERVER env
// var (a server is required to define an account). Aliases are lowercased
// suffixes and sorted/deduped. Use this to register alias commands for accounts
// that have never been added with `wpci account add`.
func EnvAccounts() []string {
	serverSuffix := "_SERVER"
	seen := map[string]bool{}
	for _, kv := range os.Environ() {
		if !strings.HasPrefix(kv, envPrefix) {
			continue
		}
		eq := strings.IndexByte(kv, '=')
		if eq < 0 {
			continue
		}
		name := kv[:eq]
		if !strings.HasSuffix(name, serverSuffix) {
			continue
		}
		if kv[eq+1:] == "" { // empty server value does not define an account
			continue
		}
		suffix := strings.TrimPrefix(name, envPrefix)
		suffix = strings.TrimSuffix(suffix, serverSuffix)
		if suffix == "" {
			continue
		}
		alias := strings.ToLower(suffix)
		seen[alias] = true
	}
	out := make([]string, 0, len(seen))
	for a := range seen {
		out = append(out, a)
	}
	sort.Strings(out)
	return out
}

// ResolveAccount returns the effective account for an alias. A stored account
// (from `wpci account add`) wins. If none exists, an ephemeral account is built
// from WPCI_<ALIAS>_SERVER (required) and WPCI_<ALIAS>_TOKEN (token resolved
// later by the client).
func ResolveAccount(alias string) (Account, error) {
	if acct, err := LoadAccount(alias); err == nil {
		return acct, nil
	}
	server := os.Getenv(EnvServerName(alias))
	if server == "" {
		return Account{}, fmt.Errorf(
			"no account %q configured; add one with `wpci account add %s --server <url>`, or set %s",
			alias, alias, EnvServerName(alias),
		)
	}
	acct := Account{Alias: alias, Server: server}
	NormalizeAccount(&acct)
	return acct, nil
}
