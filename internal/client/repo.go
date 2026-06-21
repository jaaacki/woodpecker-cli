package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/jaaacki/woodpecker-cli/internal/api"
)

// ResolveRepo finds a repo by owner/repo full name. It prefers the lookup
// endpoint and falls back to paginating /repos when the lookup fails.
func (c *Client) ResolveRepo(fullName string) (api.Repo, error) {
	var repo api.Repo
	if strings.TrimSpace(fullName) == "" {
		return repo, api.RepoNotFoundError{FullName: fullName}
	}

	lookupURL := c.URL("repos", "lookup", fullName)
	req, err := c.Request(http.MethodGet, lookupURL, nil)
	if err != nil {
		return repo, err
	}
	_, b, err := c.Do(req)
	if err == nil {
		if err := json.Unmarshal(b, &repo); err != nil {
			return repo, fmt.Errorf("parsing lookup response: %w", err)
		}
		return repo, nil
	}

	var apiErr api.APIError
	if errors.As(err, &apiErr) && (apiErr.NotFound() || apiErr.BadRequest()) {
		return c.findRepoByScan(fullName)
	}
	return repo, err
}

// findRepoByScan paginates through /repos to find a matching full_name.
func (c *Client) findRepoByScan(fullName string) (api.Repo, error) {
	var repo api.Repo
	page := 1
	for {
		params := url.Values{}
		params.Set("page", fmt.Sprintf("%d", page))
		params.Set("per_page", "100")
		urlStr := SetQuery(c.URL("repos"), params)
		var repos []api.Repo
		if err := c.GetJSON(urlStr, &repos); err != nil {
			return repo, err
		}
		if len(repos) == 0 {
			break
		}
		for _, r := range repos {
			if r.FullName == fullName {
				return r, nil
			}
		}
		page++
	}
	return repo, api.RepoNotFoundError{FullName: fullName}
}

// RepoID returns the ID of a repo by full name.
func (c *Client) RepoID(fullName string) (int64, error) {
	repo, err := c.ResolveRepo(fullName)
	if err != nil {
		return 0, err
	}
	return repo.ID, nil
}
