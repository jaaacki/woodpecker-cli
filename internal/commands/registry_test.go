package commands

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jaaacki/woodpecker-cli/internal/output"
)

func TestRegistryAddRepo(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/repos/lookup/owner/repo":
			_, _ = w.Write([]byte(`{"id": 42, "owner": "owner", "name": "repo", "full_name": "owner/repo"}`))
		case r.Method == http.MethodPost && r.URL.Path == "/api/repos/42/registries":
			body, _ := io.ReadAll(r.Body)
			if !contains(string(body), `"address":"docker.io"`) {
				http.Error(w, "expected registry body", http.StatusBadRequest)
				return
			}
			_, _ = w.Write([]byte(`{"id": 1, "address": "docker.io", "username": "u"}`))
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer ts.Close()

	cleanup := setupTestAccount(ts.URL)
	defer cleanup()

	ctx := output.NewJSONContext()
	cmd := newRegistryAddCommand("test", func() output.Context { return ctx })
	addSafetyFlags(cmd)
	cmd.SetArgs([]string{"--write", "--username", "u", "--password", "p", "repo", "owner/repo", "docker.io"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestRegistryDeleteRepo(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/repos/lookup/owner/repo":
			_, _ = w.Write([]byte(`{"id": 42, "owner": "owner", "name": "repo", "full_name": "owner/repo"}`))
		case r.Method == http.MethodDelete && r.URL.Path == "/api/repos/42/registries/docker.io":
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer ts.Close()

	cleanup := setupTestAccount(ts.URL)
	defer cleanup()

	ctx := output.NewJSONContext()
	cmd := newRegistryDeleteCommand("test", func() output.Context { return ctx })
	addSafetyFlags(cmd)
	cmd.SetArgs([]string{"--write", "--confirm", "docker.io", "repo", "owner/repo", "docker.io"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}
