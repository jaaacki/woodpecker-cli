package commands

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spf13/cobra"

	"github.com/jaaacki/woodpecker-cli/internal/auth"
	"github.com/jaaacki/woodpecker-cli/internal/client"
	"github.com/jaaacki/woodpecker-cli/internal/config"
	"github.com/jaaacki/woodpecker-cli/internal/output"
)

func TestNewContextFromCmd(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.PersistentFlags().Bool("json", false, "")
	cmd.PersistentFlags().Bool("raw", false, "")
	_ = cmd.ParseFlags([]string{"--json"})

	ctx := NewContextFromCmd(cmd)
	if !ctx.JSON {
		t.Fatal("expected json flag to be true")
	}
}

func TestVersionCommand(t *testing.T) {
	cmd := NewVersionCommand()
	// Verify it can be invoked without arguments.
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatal(err)
	}
}

func TestProbeVersionTreatsHTMLAsUnavailable(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<!doctype html><html></html>"))
	}))
	defer ts.Close()

	acct := config.Account{Alias: "home", Server: ts.URL, APIBase: "/api"}
	c := client.NewWithToken(acct, "token", output.NewContext())

	probe := probeVersion(c)
	if probe.Available {
		t.Fatal("expected version to be unavailable")
	}
	if probe.Note == "" {
		t.Fatal("expected unavailable note")
	}
}

func TestAccountTestUsesUserWhenVersionUnavailable(t *testing.T) {
	dir := t.TempDir()
	if err := config.SetConfigDir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(config.ResetConfigDir)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer token" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		switch r.URL.Path {
		case "/api/user":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":1,"login":"kylerhuang-ux","email":"kyler.huang@ada.asia"}`))
		case "/api/version":
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte("<!doctype html><html></html>"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	acct := config.Account{Alias: "home", Server: ts.URL, APIBase: "/api"}
	if err := acct.Save(); err != nil {
		t.Fatal(err)
	}
	if err := auth.NewToken("home").Save("token"); err != nil {
		t.Fatal(err)
	}

	cmd := accountTestCommand()
	cmd.SetArgs([]string{"home"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

// TestWriter is a minimal io.Writer for future command tests.
type TestWriter struct {
	bytes []byte
}

func (w *TestWriter) Write(p []byte) (int, error) {
	w.bytes = append(w.bytes, p...)
	return len(p), nil
}
