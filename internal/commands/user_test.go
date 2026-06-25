package commands

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jaaacki/woodpecker-cli/internal/output"
)

func TestUserCreate(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost && r.URL.Path == "/api/users" {
			body, _ := io.ReadAll(r.Body)
			if !contains(string(body), `"login":"alice"`) {
				http.Error(w, "expected user body", http.StatusBadRequest)
				return
			}
			_, _ = w.Write([]byte(`{"id": 1, "login": "alice"}`))
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer ts.Close()

	cleanup := setupTestAccount(ts.URL)
	defer cleanup()

	ctx := output.NewJSONContext()
	cmd := newUserCreateCommand("test", func() output.Context { return ctx })
	addSafetyFlags(cmd)
	cmd.SetArgs([]string{"--write", "alice"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestUserDelete(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodDelete && r.URL.Path == "/api/users/alice" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer ts.Close()

	cleanup := setupTestAccount(ts.URL)
	defer cleanup()

	ctx := output.NewJSONContext()
	cmd := newUserDeleteCommand("test", func() output.Context { return ctx })
	addSafetyFlags(cmd)
	cmd.SetArgs([]string{"--write", "--confirm", "alice", "alice"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestUserEditAdmin(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPatch && r.URL.Path == "/api/users/alice" {
			body, _ := io.ReadAll(r.Body)
			if !contains(string(body), `"admin":true`) {
				http.Error(w, "expected admin=true", http.StatusBadRequest)
				return
			}
			_, _ = w.Write([]byte(`{"id": 1, "login": "alice", "admin": true}`))
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer ts.Close()

	cleanup := setupTestAccount(ts.URL)
	defer cleanup()

	ctx := output.NewJSONContext()
	cmd := newUserEditCommand("test", func() output.Context { return ctx })
	addSafetyFlags(cmd)
	cmd.SetArgs([]string{"--write", "alice", "--admin"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestUserTokenShow(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/api/user/token" {
			_, _ = w.Write([]byte("plain-token-value"))
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer ts.Close()

	cleanup := setupTestAccount(ts.URL)
	defer cleanup()

	ctx := output.NewJSONContext()
	cmd := newUserTokenShowCommand("test", func() output.Context { return ctx })
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestUserTokenReset(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete && r.URL.Path == "/api/user/token" {
			_, _ = w.Write([]byte("new-token-value"))
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer ts.Close()

	cleanup := setupTestAccount(ts.URL)
	defer cleanup()

	ctx := output.NewJSONContext()
	cmd := newUserTokenResetCommand("test", func() output.Context { return ctx })
	addSafetyFlags(cmd)
	cmd.SetArgs([]string{"--write"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}
