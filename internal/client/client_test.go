package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jaaacki/woodpecker-cli/internal/api"
	"github.com/jaaacki/woodpecker-cli/internal/config"
	"github.com/jaaacki/woodpecker-cli/internal/output"
)

func TestNewWithToken(t *testing.T) {
	acct := config.Account{
		Alias:          "test",
		Server:         "https://ci.example.com",
		APIBase:        "/api",
		TimeoutSeconds: 30,
	}
	c := NewWithToken(acct, "token", output.NewContext())
	if c.BaseURL != "https://ci.example.com/api" {
		t.Fatalf("base URL mismatch: %s", c.BaseURL)
	}
}

func TestURL(t *testing.T) {
	acct := config.Account{
		Server:  "https://ci.example.com",
		APIBase: "/api/",
	}
	c := NewWithToken(acct, "token", output.NewContext())
	if got := c.URL("repos", "1"); got != "https://ci.example.com/api/repos/1" {
		t.Fatalf("url mismatch: %s", got)
	}
}

func TestSetQuery(t *testing.T) {
	u := SetQuery("https://example.com/api/repos", map[string][]string{"page": {"2"}})
	if u != "https://example.com/api/repos?page=2" {
		t.Fatalf("query mismatch: %s", u)
	}
}

func TestGetJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer token" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id": 1, "login": "user"}`))
	}))
	defer ts.Close()

	acct := config.Account{Server: ts.URL, APIBase: "/"}
	c := NewWithToken(acct, "token", output.NewContext())
	var user api.User
	if err := c.GetJSON(ts.URL+"/user", &user); err != nil {
		t.Fatal(err)
	}
	if user.ID != 1 || user.Login != "user" {
		t.Fatalf("unexpected user: %+v", user)
	}
}

func TestExitForError(t *testing.T) {
	if got := ExitForError(nil); got != output.ExitSuccess {
		t.Fatalf("expected success exit code, got %d", got)
	}
	if got := ExitForError(api.APIError{StatusCode: 401}); got != output.ExitAuth {
		t.Fatalf("expected auth exit code, got %d", got)
	}
	if got := ExitForError(api.APIError{StatusCode: 500}); got != output.ExitAPI {
		t.Fatalf("expected api exit code, got %d", got)
	}
	if got := ExitForError(api.RepoNotFoundError{FullName: "owner/repo"}); got != output.ExitAPI {
		t.Fatalf("expected api exit code for repo not found, got %d", got)
	}
}

func TestResolveRepo(t *testing.T) {
	var lookupHit bool
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/repos/lookup/owner/repo":
			lookupHit = true
			_, _ = w.Write([]byte(`{"id": 42, "owner": "owner", "name": "repo", "full_name": "owner/repo"}`))
		case "/api/repos":
			_, _ = w.Write([]byte(`[]`))
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer ts.Close()

	acct := config.Account{Server: ts.URL, APIBase: "/api", TimeoutSeconds: 30}
	c := NewWithToken(acct, "token", output.NewContext())
	repo, err := c.ResolveRepo("owner/repo")
	if err != nil {
		t.Fatal(err)
	}
	if repo.ID != 42 {
		t.Fatalf("expected repo id 42, got %d", repo.ID)
	}
	if !lookupHit {
		t.Fatal("expected lookup endpoint to be used")
	}
}

func TestResolveRepoFallback(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/repos/lookup/owner/repo":
			http.Error(w, "not found", http.StatusNotFound)
		case "/api/repos":
			_, _ = w.Write([]byte(`[
				{"id": 1, "owner": "other", "name": "other", "full_name": "other/other"},
				{"id": 42, "owner": "owner", "name": "repo", "full_name": "owner/repo"}
			]`))
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer ts.Close()

	acct := config.Account{Server: ts.URL, APIBase: "/api", TimeoutSeconds: 30}
	c := NewWithToken(acct, "token", output.NewContext())
	repo, err := c.ResolveRepo("owner/repo")
	if err != nil {
		t.Fatal(err)
	}
	if repo.ID != 42 {
		t.Fatalf("expected repo id 42, got %d", repo.ID)
	}
}

func TestPostJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "expected POST", http.StatusMethodNotAllowed)
			return
		}
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "expected json content type", http.StatusBadRequest)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if body["name"] != "value" {
			http.Error(w, "unexpected body", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":1}`))
	}))
	defer ts.Close()

	acct := config.Account{Server: ts.URL, APIBase: "/api", TimeoutSeconds: 30}
	c := NewWithToken(acct, "token", output.NewContext())
	var result struct {
		ID int64 `json:"id"`
	}
	if err := c.PostJSON(c.URL("test"), map[string]any{"name": "value"}, &result); err != nil {
		t.Fatal(err)
	}
	if result.ID != 1 {
		t.Fatalf("expected id 1, got %d", result.ID)
	}
}

func TestDelete(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "expected DELETE", http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	acct := config.Account{Server: ts.URL, APIBase: "/api", TimeoutSeconds: 30}
	c := NewWithToken(acct, "token", output.NewContext())
	if _, err := c.Delete(c.URL("repos", "1")); err != nil {
		t.Fatal(err)
	}
}

func TestPatchJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			http.Error(w, "expected PATCH", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"active":true}`))
	}))
	defer ts.Close()

	acct := config.Account{Server: ts.URL, APIBase: "/api", TimeoutSeconds: 30}
	c := NewWithToken(acct, "token", output.NewContext())
	var repo api.Repo
	if err := c.PatchJSON(c.URL("repos", "1"), api.Repo{Active: true}, &repo); err != nil {
		t.Fatal(err)
	}
	if !repo.Active {
		t.Fatal("expected active repo")
	}
}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/repos/lookup/owner/repo":
			http.Error(w, "not found", http.StatusNotFound)
		case "/api/repos":
			_, _ = w.Write([]byte(`[]`))
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer ts.Close()

	acct := config.Account{Server: ts.URL, APIBase: "/api", TimeoutSeconds: 30}
	c := NewWithToken(acct, "token", output.NewContext())
	_, err := c.ResolveRepo("owner/repo")
	var repoErr api.RepoNotFoundError
	if !errors.As(err, &repoErr) {
		t.Fatalf("expected RepoNotFoundError, got %T: %v", err, err)
	}
}
