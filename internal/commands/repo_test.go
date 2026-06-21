package commands

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/jaaacki/woodpecker-cli/internal/auth"
	"github.com/jaaacki/woodpecker-cli/internal/config"
	"github.com/jaaacki/woodpecker-cli/internal/output"
)

func TestRepoList(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/api/repos" {
			_, _ = w.Write([]byte(`[
				{"id": 1, "owner": "owner1", "name": "repo1", "full_name": "owner1/repo1", "scm": "git", "active": true, "default_branch": "main"},
				{"id": 2, "owner": "owner2", "name": "repo2", "full_name": "owner2/repo2", "scm": "git", "active": false, "default_branch": "dev"}
			]`))
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer ts.Close()

	ctx := output.NewJSONContext()
	acct := config.Account{Alias: "test", Server: ts.URL, APIBase: "/api", TimeoutSeconds: 30}
	if err := acct.Save(); err != nil {
		t.Fatal(err)
	}
	defer config.RemoveAccount("test")
	_ = auth.NewToken("test").Save("token")
	defer auth.NewToken("test").Remove()

	var repos []any
	cmd := newRepoListCommand("test", func() output.Context { return ctx })
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	_ = repos
}

func TestRepoShowLookup(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/api/repos/lookup/owner/repo" {
			_, _ = w.Write([]byte(`{"id": 42, "owner": "owner", "name": "repo", "full_name": "owner/repo", "scm": "git", "active": true, "default_branch": "main"}`))
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer ts.Close()

	acct := config.Account{Alias: "test", Server: ts.URL, APIBase: "/api", TimeoutSeconds: 30}
	if err := acct.Save(); err != nil {
		t.Fatal(err)
	}
	defer config.RemoveAccount("test")
	_ = auth.NewToken("test").Save("token")
	defer auth.NewToken("test").Remove()

	ctx := output.NewJSONContext()
	cmd := newRepoShowCommand("test", func() output.Context { return ctx })
	cmd.SetArgs([]string{"owner/repo"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestRepoCommandUsesSubprocess(t *testing.T) {
	repoRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	tmpDir := t.TempDir()
	if err := config.SetConfigDir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer config.ResetConfigDir()
	if err := config.EnsureConfigDirs(); err != nil {
		t.Fatal(err)
	}
	acct := config.Account{Alias: "test", Server: "https://ci.example.com", APIBase: "/api", TimeoutSeconds: 30}
	_ = acct.Save()
	defer config.RemoveAccount("test")

	bin := filepath.Join(tmpDir, "wpci")
	build := exec.Command("go", "build", "-o", bin, "./cmd/wpci")
	build.Dir = repoRoot
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("building wpci: %v\n%s", err, out)
	}
	defer os.Remove(bin)

	cmd := exec.Command(bin, "test", "repo", "--help")
	cmd.Env = append(os.Environ(), "WPCI_CONFIG_DIR="+tmpDir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("wpci test repo --help: %v\n%s", err, out)
	}
	if !bytes.Contains(out, []byte("Repository operations")) {
		t.Fatalf("expected repo help, got:\n%s", out)
	}
}
