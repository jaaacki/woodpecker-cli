package commands

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jaaacki/woodpecker-cli/internal/output"
)

func TestCronAdd(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/repos/lookup/owner/repo":
			_, _ = w.Write([]byte(`{"id": 42, "owner": "owner", "name": "repo", "full_name": "owner/repo"}`))
		case r.Method == http.MethodPost && r.URL.Path == "/api/repos/42/cron":
			body, _ := io.ReadAll(r.Body)
			if !contains(string(body), `"name":"nightly"`) || !contains(string(body), `"schedule":"0 0 * * *"`) {
				http.Error(w, "expected cron body", http.StatusBadRequest)
				return
			}
			_, _ = w.Write([]byte(`{"id": 1, "name": "nightly", "schedule": "0 0 * * *", "branch": "main"}`))
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer ts.Close()

	cleanup := setupTestAccount(ts.URL)
	defer cleanup()

	ctx := output.NewJSONContext()
	cmd := newCronAddCommand("test", func() output.Context { return ctx })
	addSafetyFlags(cmd)
	cmd.SetArgs([]string{"--write", "--schedule", "0 0 * * *", "--branch", "main", "owner/repo", "nightly"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestCronDelete(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/repos/lookup/owner/repo":
			_, _ = w.Write([]byte(`{"id": 42, "owner": "owner", "name": "repo", "full_name": "owner/repo"}`))
		case r.Method == http.MethodDelete && r.URL.Path == "/api/repos/42/cron/nightly":
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer ts.Close()

	cleanup := setupTestAccount(ts.URL)
	defer cleanup()

	ctx := output.NewJSONContext()
	cmd := newCronDeleteCommand("test", func() output.Context { return ctx })
	addSafetyFlags(cmd)
	cmd.SetArgs([]string{"--write", "--confirm", "owner/repo/nightly", "owner/repo", "nightly"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func contains(s, substr string) bool {
	return len(substr) <= len(s) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
