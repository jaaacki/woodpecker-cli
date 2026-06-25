package commands

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"

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

func addSafetyFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("write", false, "")
	cmd.Flags().String("confirm", "", "")
}

func setupTestAccount(server string) func() {
	acct := config.Account{Alias: "test", Server: server, APIBase: "/api", TimeoutSeconds: 30}
	if err := acct.Save(); err != nil {
		panic(err)
	}
	_ = auth.NewToken("test").Save("token")
	return func() {
		config.RemoveAccount("test")
		auth.NewToken("test").Remove()
	}
}

func TestRepoEnable(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost && r.URL.Path == "/api/repos" && r.URL.Query().Get("forge_remote_id") == "12345" {
			_, _ = w.Write([]byte(`{"id": 7, "owner": "owner", "name": "repo", "full_name": "owner/repo"}`))
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer ts.Close()

	cleanup := setupTestAccount(ts.URL)
	defer cleanup()

	ctx := output.NewJSONContext()
	cmd := newRepoEnableCommand("test", func() output.Context { return ctx })
	addSafetyFlags(cmd)
	cmd.SetArgs([]string{"--write", "12345"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestRepoDisable(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/repos/lookup/owner/repo":
			_, _ = w.Write([]byte(`{"id": 7, "owner": "owner", "name": "repo", "full_name": "owner/repo", "active": true}`))
		case r.Method == http.MethodDelete && r.URL.Path == "/api/repos/7" && r.URL.Query().Get("remove") == "false":
			_, _ = w.Write([]byte(`{"id": 7, "owner": "owner", "name": "repo", "full_name": "owner/repo", "active": false}`))
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer ts.Close()

	cleanup := setupTestAccount(ts.URL)
	defer cleanup()

	ctx := output.NewJSONContext()
	cmd := newRepoDisableCommand("test", func() output.Context { return ctx })
	addSafetyFlags(cmd)
	cmd.SetArgs([]string{"--write", "--confirm", "owner/repo", "owner/repo"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestRepoEdit(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/repos/lookup/owner/repo":
			_, _ = w.Write([]byte(`{"id": 7, "owner": "owner", "name": "repo", "full_name": "owner/repo"}`))
		case r.Method == http.MethodPatch && r.URL.Path == "/api/repos/7":
			body, _ := io.ReadAll(r.Body)
			if !bytes.Contains(body, []byte(`"timeout":90`)) || !bytes.Contains(body, []byte(`"visibility":"private"`)) {
				http.Error(w, "expected timeout and visibility", http.StatusBadRequest)
				return
			}
			_, _ = w.Write([]byte(`{"id": 7, "owner": "owner", "name": "repo", "full_name": "owner/repo", "timeout": 90, "visibility": "private"}`))
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer ts.Close()

	cleanup := setupTestAccount(ts.URL)
	defer cleanup()

	ctx := output.NewJSONContext()
	cmd := newRepoEditCommand("test", func() output.Context { return ctx })
	addSafetyFlags(cmd)
	cmd.SetArgs([]string{"--write", "--timeout", "90", "--visibility", "private", "owner/repo"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestRepoDelete(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/repos/lookup/owner/repo":
			_, _ = w.Write([]byte(`{"id": 7, "owner": "owner", "name": "repo", "full_name": "owner/repo"}`))
		case r.Method == http.MethodDelete && r.URL.Path == "/api/repos/7" && r.URL.Query().Get("remove") == "true":
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer ts.Close()

	cleanup := setupTestAccount(ts.URL)
	defer cleanup()

	ctx := output.NewJSONContext()
	cmd := newRepoDeleteCommand("test", func() output.Context { return ctx })
	addSafetyFlags(cmd)
	cmd.SetArgs([]string{"--write", "--confirm", "owner/repo", "owner/repo"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestRepoRepair(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/repos/lookup/owner/repo":
			_, _ = w.Write([]byte(`{"id": 7, "owner": "owner", "name": "repo", "full_name": "owner/repo"}`))
		case r.Method == http.MethodPost && r.URL.Path == "/api/repos/7/repair":
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer ts.Close()

	cleanup := setupTestAccount(ts.URL)
	defer cleanup()

	ctx := output.NewJSONContext()
	cmd := newRepoRepairCommand("test", func() output.Context { return ctx })
	addSafetyFlags(cmd)
	cmd.SetArgs([]string{"--write", "owner/repo"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestRepoChown(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/repos/lookup/owner/repo":
			_, _ = w.Write([]byte(`{"id": 7, "owner": "owner", "name": "repo", "full_name": "owner/repo"}`))
		case r.Method == http.MethodPost && r.URL.Path == "/api/repos/7/chown" && r.URL.Query().Get("user_id") == "99":
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer ts.Close()

	cleanup := setupTestAccount(ts.URL)
	defer cleanup()

	ctx := output.NewJSONContext()
	cmd := newRepoChownCommand("test", func() output.Context { return ctx })
	addSafetyFlags(cmd)
	cmd.SetArgs([]string{"--write", "--confirm", "owner/repo", "owner/repo", "99"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestRepoMove(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/repos/lookup/owner/repo":
			_, _ = w.Write([]byte(`{"id": 7, "owner": "owner", "name": "repo", "full_name": "owner/repo"}`))
		case r.Method == http.MethodPost && r.URL.Path == "/api/repos/7/move" && r.URL.Query().Get("to") == "newowner/repo":
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer ts.Close()

	cleanup := setupTestAccount(ts.URL)
	defer cleanup()

	ctx := output.NewJSONContext()
	cmd := newRepoMoveCommand("test", func() output.Context { return ctx })
	addSafetyFlags(cmd)
	cmd.SetArgs([]string{"--write", "--confirm", "owner/repo", "owner/repo", "newowner/repo"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}
