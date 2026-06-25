package commands

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jaaacki/woodpecker-cli/internal/output"
)

func TestAgentCreate(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost && r.URL.Path == "/api/agents" {
			body, _ := io.ReadAll(r.Body)
			if !contains(string(body), `"name":"agent-1"`) {
				http.Error(w, "expected agent body", http.StatusBadRequest)
				return
			}
			_, _ = w.Write([]byte(`{"id": 1, "name": "agent-1"}`))
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer ts.Close()

	cleanup := setupTestAccount(ts.URL)
	defer cleanup()

	ctx := output.NewJSONContext()
	cmd := newAgentCreateCommand("test", func() output.Context { return ctx })
	addSafetyFlags(cmd)
	cmd.SetArgs([]string{"--write", "--name", "agent-1"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestAgentDelete(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodDelete && r.URL.Path == "/api/agents/7" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer ts.Close()

	cleanup := setupTestAccount(ts.URL)
	defer cleanup()

	ctx := output.NewJSONContext()
	cmd := newAgentDeleteCommand("test", func() output.Context { return ctx })
	addSafetyFlags(cmd)
	cmd.SetArgs([]string{"--write", "--confirm", "7", "7"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestAgentEdit(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPatch && r.URL.Path == "/api/agents/7" {
			body, _ := io.ReadAll(r.Body)
			// Server should receive only the changed field, not an empty name.
			if !contains(string(body), `"no_schedule":true`) {
				http.Error(w, "expected no_schedule=true", http.StatusBadRequest)
				return
			}
			if contains(string(body), `"name":""`) {
				http.Error(w, "unexpected empty name", http.StatusBadRequest)
				return
			}
			_, _ = w.Write([]byte(`{"id": 7, "name": "agent-7"}`))
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer ts.Close()

	cleanup := setupTestAccount(ts.URL)
	defer cleanup()

	ctx := output.NewJSONContext()
	cmd := newAgentEditCommand("test", func() output.Context { return ctx })
	addSafetyFlags(cmd)
	cmd.SetArgs([]string{"--write", "--confirm", "7", "7", "--no-schedule"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}
