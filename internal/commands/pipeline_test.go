package commands

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jaaacki/woodpecker-cli/internal/auth"
	"github.com/jaaacki/woodpecker-cli/internal/config"
	"github.com/jaaacki/woodpecker-cli/internal/output"
)

func TestPipelineList(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/repos/lookup/owner/repo":
			_, _ = w.Write([]byte(`{"id": 42, "owner": "owner", "name": "repo", "full_name": "owner/repo"}`))
		case "/api/repos/42/pipelines":
			_, _ = w.Write([]byte(`[
				{"number": 1, "branch": "main", "status": "success", "event": "push", "created": 1718000000},
				{"number": 2, "branch": "main", "status": "running", "event": "push", "created": 1719000000}
			]`))
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer ts.Close()

	ctx := output.NewJSONContext()
	acct := config.Account{Alias: "test", Server: ts.URL, APIBase: "/api", TimeoutSeconds: 30}
	_ = acct.Save()
	defer config.RemoveAccount("test")
	_ = auth.NewToken("test").Save("token")
	defer auth.NewToken("test").Remove()

	cmd := newPipelineListCommand("test", func() output.Context { return ctx })
	cmd.SetArgs([]string{"owner/repo"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestPipelinePs(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/repos/lookup/owner/repo":
			_, _ = w.Write([]byte(`{"id": 42, "owner": "owner", "name": "repo", "full_name": "owner/repo"}`))
		case "/api/repos/42/pipelines/7":
			_, _ = w.Write([]byte(`{"id": 7, "workflows": [{"id": 200, "name": "default", "children": [
				{"id": 101, "name": "build", "state": "success", "started": 1718000000, "stopped": 1718000100, "exit_code": 0},
				{"id": 102, "name": "test", "state": "failure", "started": 1718000200, "stopped": 1718000300, "exit_code": 1}
			]}]}`))
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer ts.Close()

	ctx := output.NewJSONContext()
	acct := config.Account{Alias: "test", Server: ts.URL, APIBase: "/api", TimeoutSeconds: 30}
	_ = acct.Save()
	defer config.RemoveAccount("test")
	_ = auth.NewToken("test").Save("token")
	defer auth.NewToken("test").Remove()

	cmd := newPipelinePsCommand("test", func() output.Context { return ctx })
	cmd.SetArgs([]string{"owner/repo", "7"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestPipelineLogRaw(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/repos/lookup/owner/repo":
			_, _ = w.Write([]byte(`{"id": 42, "owner": "owner", "name": "repo", "full_name": "owner/repo"}`))
		case "/api/repos/42/pipelines/7":
			_, _ = w.Write([]byte(`{"id": 7, "workflows": [{"id": 200, "children": [{"id": 101, "name": "build"}]}]}`))
		case "/api/repos/42/logs/7/101":
			_, _ = w.Write([]byte(`[{"id": 1, "step_id": 101, "data": "bGluZTEK"}]`))
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer ts.Close()

	acct := config.Account{Alias: "test", Server: ts.URL, APIBase: "/api", TimeoutSeconds: 30}
	_ = acct.Save()
	defer config.RemoveAccount("test")
	_ = auth.NewToken("test").Save("token")
	defer auth.NewToken("test").Remove()

	var buf bytes.Buffer
	ctx := output.Context{Raw: true, Out: &buf, Err: output.NewContext().Err}
	cmd := newPipelineLogShowCommand("test", func() output.Context { return ctx })
	cmd.SetArgs([]string{"owner/repo", "7", "build"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(buf.Bytes(), []byte("line1")) {
		t.Fatalf("expected raw log output, got:\n%s", buf.String())
	}
}

func TestPipelineRun(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/repos/lookup/owner/repo":
			_, _ = w.Write([]byte(`{"id": 42, "owner": "owner", "name": "repo", "full_name": "owner/repo"}`))
		case r.Method == http.MethodPost && r.URL.Path == "/api/repos/42/pipelines":
			body, _ := io.ReadAll(r.Body)
			if !bytes.Contains(body, []byte(`"branch":"dev"`)) || !bytes.Contains(body, []byte(`"FOO":"bar"`)) {
				http.Error(w, "expected branch and variables", http.StatusBadRequest)
				return
			}
			_, _ = w.Write([]byte(`{"id": 1, "number": 8, "branch": "dev", "status": "pending"}`))
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer ts.Close()

	cleanup := setupTestAccount(ts.URL)
	defer cleanup()

	ctx := output.NewJSONContext()
	cmd := newPipelineRunCommand("test", func() output.Context { return ctx })
	addSafetyFlags(cmd)
	cmd.SetArgs([]string{"--write", "--branch", "dev", "--var", "FOO=bar", "owner/repo"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestPipelineRestart(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/repos/lookup/owner/repo":
			_, _ = w.Write([]byte(`{"id": 42, "owner": "owner", "name": "repo", "full_name": "owner/repo"}`))
		case r.Method == http.MethodPost && r.URL.Path == "/api/repos/42/pipelines/7":
			_, _ = w.Write([]byte(`{"id": 1, "number": 7, "status": "pending"}`))
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer ts.Close()

	cleanup := setupTestAccount(ts.URL)
	defer cleanup()

	ctx := output.NewJSONContext()
	cmd := newPipelineRestartCommand("test", func() output.Context { return ctx })
	addSafetyFlags(cmd)
	cmd.SetArgs([]string{"--write", "owner/repo", "7"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestPipelineApprove(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/repos/lookup/owner/repo":
			_, _ = w.Write([]byte(`{"id": 42, "owner": "owner", "name": "repo", "full_name": "owner/repo"}`))
		case r.Method == http.MethodPost && r.URL.Path == "/api/repos/42/pipelines/7/approve":
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer ts.Close()

	cleanup := setupTestAccount(ts.URL)
	defer cleanup()

	ctx := output.NewJSONContext()
	cmd := newPipelineApproveCommand("test", func() output.Context { return ctx })
	addSafetyFlags(cmd)
	cmd.SetArgs([]string{"--write", "owner/repo", "7"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestPipelineDecline(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/repos/lookup/owner/repo":
			_, _ = w.Write([]byte(`{"id": 42, "owner": "owner", "name": "repo", "full_name": "owner/repo"}`))
		case r.Method == http.MethodPost && r.URL.Path == "/api/repos/42/pipelines/7/decline":
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer ts.Close()

	cleanup := setupTestAccount(ts.URL)
	defer cleanup()

	ctx := output.NewJSONContext()
	cmd := newPipelineDeclineCommand("test", func() output.Context { return ctx })
	addSafetyFlags(cmd)
	cmd.SetArgs([]string{"--write", "owner/repo", "7"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestPipelineCancel(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/repos/lookup/owner/repo":
			_, _ = w.Write([]byte(`{"id": 42, "owner": "owner", "name": "repo", "full_name": "owner/repo"}`))
		case r.Method == http.MethodPost && r.URL.Path == "/api/repos/42/pipelines/7/cancel":
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer ts.Close()

	cleanup := setupTestAccount(ts.URL)
	defer cleanup()

	ctx := output.NewJSONContext()
	cmd := newPipelineCancelCommand("test", func() output.Context { return ctx })
	addSafetyFlags(cmd)
	cmd.SetArgs([]string{"--write", "owner/repo", "7"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}
