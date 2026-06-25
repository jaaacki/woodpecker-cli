package commands

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jaaacki/woodpecker-cli/internal/output"
)

func TestSecretAddRepo(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/repos/lookup/owner/repo":
			_, _ = w.Write([]byte(`{"id": 42, "owner": "owner", "name": "repo", "full_name": "owner/repo"}`))
		case r.Method == http.MethodPost && r.URL.Path == "/api/repos/42/secrets":
			body, _ := io.ReadAll(r.Body)
			if !contains(string(body), `"name":"docker"`) {
				http.Error(w, "expected secret body", http.StatusBadRequest)
				return
			}
			_, _ = w.Write([]byte(`{"id": 1, "name": "docker"}`))
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer ts.Close()

	cleanup := setupTestAccount(ts.URL)
	defer cleanup()

	ctx := output.NewJSONContext()
	cmd := newSecretAddCommand("test", func() output.Context { return ctx })
	addSafetyFlags(cmd)
	cmd.SetArgs([]string{"--write", "--value", "secret", "repo", "owner/repo", "docker"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestSecretDeleteRepo(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/repos/lookup/owner/repo":
			_, _ = w.Write([]byte(`{"id": 42, "owner": "owner", "name": "repo", "full_name": "owner/repo"}`))
		case r.Method == http.MethodDelete && r.URL.Path == "/api/repos/42/secrets/docker":
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer ts.Close()

	cleanup := setupTestAccount(ts.URL)
	defer cleanup()

	ctx := output.NewJSONContext()
	cmd := newSecretDeleteCommand("test", func() output.Context { return ctx })
	addSafetyFlags(cmd)
	cmd.SetArgs([]string{"--write", "--confirm", "docker", "repo", "owner/repo", "docker"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}
