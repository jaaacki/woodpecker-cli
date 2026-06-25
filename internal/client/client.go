package client

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jaaacki/woodpecker-cli/internal/api"
	"github.com/jaaacki/woodpecker-cli/internal/auth"
	"github.com/jaaacki/woodpecker-cli/internal/config"
	"github.com/jaaacki/woodpecker-cli/internal/output"
)

// Client is a configured Woodpecker API HTTP client.
type Client struct {
	Account    config.Account
	Token      auth.Token
	TokenValue string
	HTTP       *http.Client
	BaseURL    string
	Out        output.Context
}

// New creates a Client from an account alias. Account and token resolve in
// order: a stored account (`wpci account add`) first, then the
// WPCI_<ALIAS>_SERVER / WPCI_<ALIAS>_TOKEN env vars as a fallback so an account
// can be used with zero configuration from a shell.
func New(alias string, out output.Context) (*Client, error) {
	acct, err := config.ResolveAccount(alias)
	if err != nil {
		return nil, err
	}
	value, _ := auth.NewToken(alias).Load()
	if value == "" {
		value = os.Getenv(config.EnvTokenName(alias))
	}
	if value == "" {
		return nil, fmt.Errorf(
			"no token for account %q: set it with `wpci account token set %s` or the %s env var",
			alias, alias, config.EnvTokenName(alias),
		)
	}
	return NewWithToken(acct, value, out), nil
}

// NewWithToken creates a Client with an explicit token.
func NewWithToken(acct config.Account, token string, out output.Context) *Client {
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: acct.TLSSkipVerify}

	httpClient := &http.Client{
		Transport: tr,
		Timeout:   time.Duration(acct.TimeoutSeconds) * time.Second,
	}
	if acct.TimeoutSeconds <= 0 {
		httpClient.Timeout = 30 * time.Second
	}

	base := acct.Server + acct.APIBase
	return &Client{
		Account:    acct,
		Token:      auth.NewToken(acct.Alias),
		TokenValue: token,
		HTTP:       httpClient,
		BaseURL:    strings.TrimRight(base, "/"),
		Out:        out,
	}
}

// URL builds an absolute API URL from path segments.
func (c *Client) URL(parts ...string) string {
	path := strings.Join(parts, "/")
	path = strings.TrimLeft(path, "/")
	return c.BaseURL + "/" + path
}

// SetQuery appends query parameters to a URL string.
func SetQuery(rawURL string, params url.Values) (string, error) {
	if len(params) == 0 {
		return rawURL, nil
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("parsing URL %q: %w", rawURL, err)
	}
	u.RawQuery = params.Encode()
	return u.String(), nil
}

// Request builds a standard GET/POST/... request with bearer auth.
func (c *Client) Request(method, urlStr string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, urlStr, body)
	if err != nil {
		return nil, err
	}
	value := c.TokenValue
	if value == "" {
		loaded, err := c.Token.Load()
		if err != nil {
			return nil, err
		}
		value = loaded
	}
	req.Header.Set("Authorization", "Bearer "+value)
	req.Header.Set("User-Agent", config.UserAgent())
	req.Header.Set("Accept", "application/json")
	if body != nil && method != http.MethodGet && method != http.MethodHead {
		req.Header.Set("Content-Type", "application/json")
	}
	return req, nil
}

// Do performs a request and returns the response or a normalized error.
func (c *Client) Do(req *http.Request) (*http.Response, []byte, error) {
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("reading response: %w", err)
	}
	if resp.StatusCode >= 400 {
		return resp, b, api.APIError{StatusCode: resp.StatusCode, Message: string(b)}
	}
	return resp, b, nil
}

// GetJSON fetches and unmarshals JSON from a URL, ignoring any raw-output flag.
// Use GetRaw when the caller only wants the raw response body.
func (c *Client) GetJSON(urlStr string, target any) error {
	req, err := c.Request(http.MethodGet, urlStr, nil)
	if err != nil {
		return err
	}
	_, b, err := c.Do(req)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, target); err != nil {
		return fmt.Errorf("parsing JSON: %w", err)
	}
	return nil
}

// GetRaw prints the raw upstream response body to the configured output.
func (c *Client) GetRaw(urlStr string) ([]byte, error) {
	req, err := c.Request(http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, err
	}
	_, b, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	c.Out.RawBytes(b)
	return b, nil
}

// GetPage fetches a page of a list endpoint.
func (c *Client) GetPage(urlStr string, target any) error {
	return c.GetJSON(urlStr, target)
}

// PostJSON sends a JSON body and unmarshals the response.
func (c *Client) PostJSON(urlStr string, body any, target any) error {
	return c.sendJSON(http.MethodPost, urlStr, body, target)
}

// Post sends a JSON body and returns the raw response body without unmarshalling.
func (c *Client) Post(urlStr string, body any) ([]byte, error) {
	return c.sendJSONRaw(http.MethodPost, urlStr, body)
}

// PatchJSON sends a JSON patch body and unmarshals the response.
func (c *Client) PatchJSON(urlStr string, body any, target any) error {
	return c.sendJSON(http.MethodPatch, urlStr, body, target)
}

// PutJSON sends a JSON body with PUT and unmarshals the response.
func (c *Client) PutJSON(urlStr string, body any, target any) error {
	return c.sendJSON(http.MethodPut, urlStr, body, target)
}

// Delete sends a DELETE request and returns any response body.
func (c *Client) Delete(urlStr string) ([]byte, error) {
	req, err := c.Request(http.MethodDelete, urlStr, nil)
	if err != nil {
		return nil, err
	}
	_, b, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (c *Client) sendJSON(method, urlStr string, body any, target any) error {
	b, err := c.sendJSONRaw(method, urlStr, body)
	if err != nil {
		return err
	}
	if target == nil {
		return nil
	}
	if err := json.Unmarshal(b, target); err != nil {
		return fmt.Errorf("parsing JSON: %w", err)
	}
	return nil
}

func (c *Client) sendJSONRaw(method, urlStr string, body any) ([]byte, error) {
	var r io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("encoding JSON: %w", err)
		}
		r = bytes.NewReader(data)
	}
	req, err := c.Request(method, urlStr, r)
	if err != nil {
		return nil, err
	}
	_, b, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// Body returns a response body with trailing newline when not in JSON mode.
func Body(b []byte) string {
	return strings.TrimSpace(string(b))
}

// FormatNumber converts an integer number to a string.
func FormatNumber(n int64) string {
	return strconv.FormatInt(n, 10)
}

// FormatBool converts a boolean to "yes"/"no".
func FormatBool(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

// FormatOptional returns the value or "-" if empty.
func FormatOptional(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

// FormatTime converts a Unix timestamp (seconds) to a short local time string.
func FormatTime(t int64) string {
	if t == 0 {
		return "-"
	}
	return time.Unix(t, 0).Format(time.RFC3339)
}

// ExitForError returns an appropriate exit code for an error.
func ExitForError(err error) int {
	if err == nil {
		return output.ExitSuccess
	}
	var repoErr api.RepoNotFoundError
	if errors.As(err, &repoErr) {
		return output.ExitAPI
	}
	var apiErr api.APIError
	if errors.As(err, &apiErr) {
		switch {
		case apiErr.Unauthorized():
			return output.ExitAuth
		case apiErr.Forbidden():
			return output.ExitAuth
		default:
			return output.ExitAPI
		}
	}
	return output.ExitRuntime
}
