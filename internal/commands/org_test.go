package commands

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jaaacki/woodpecker-cli/internal/output"
)

func TestOrgDelete(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/orgs/lookup/acme":
			_, _ = w.Write([]byte(`{"id": 9, "name": "acme"}`))
		case r.Method == http.MethodDelete && r.URL.Path == "/api/orgs/9":
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer ts.Close()

	cleanup := setupTestAccount(ts.URL)
	defer cleanup()

	ctx := output.NewJSONContext()
	cmd := newOrgDeleteCommand("test", func() output.Context { return ctx })
	addSafetyFlags(cmd)
	cmd.SetArgs([]string{"--write", "--confirm", "acme", "acme"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}
