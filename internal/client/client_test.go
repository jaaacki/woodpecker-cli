package client

import (
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
}
