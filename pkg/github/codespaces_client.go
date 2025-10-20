package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// CodespacesClient is a thin wrapper around the GitHub Codespaces REST endpoints.
// Place this in pkg/github so other parts of the MCP can reuse it.
type CodespacesClient struct {
	HTTPClient *http.Client
	BaseURL    *url.URL
	UserAgent  string
}

// NewCodespacesClient returns a configured client. If httpClient is nil, a default is used.
func NewCodespacesClient(httpClient *http.Client) *CodespacesClient {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 20 * time.Second}
	}
	base, _ := url.Parse("https://api.github.com")
	return &CodespacesClient{
		HTTPClient: httpClient,
		BaseURL:    base,
		UserAgent:  "github-mcp-server/codespaces-client",
	}
}

func (c *CodespacesClient) newRequest(ctx context.Context, method, path, token string, body interface{}) (*http.Request, error) {
	u := c.BaseURL.ResolveReference(&url.URL{Path: path})
	var buf io.Reader
	if body != nil {
		bs, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal body: %w", err)
		}
		buf = bytes.NewReader(bs)
	}
	req, err := http.NewRequestWithContext(ctx, method, u.String(), buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json;apiVersion=2022-11-28")
	req.Header.Set("User-Agent", c.UserAgent)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		// use "token " prefix to match current repo examples which use token-style auth
		req.Header.Set("Authorization", "token "+token)
	}
	return req, nil
}

func (c *CodespacesClient) doRaw(req *http.Request) (int, http.Header, []byte, error) {
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return 0, nil, nil, err
	}
	defer resp.Body.Close()
	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, resp.Header, nil, err
	}
	return resp.StatusCode, resp.Header, bs, nil
}

// GetTokenScopes returns the token scopes from the X-OAuth-Scopes header (GET /).
func (c *CodespacesClient) GetTokenScopes(ctx context.Context, token string) ([]string, error) {
	req, err := c.newRequest(ctx, "GET", "/", token, nil)
	if err != nil {
		return nil, err
	}
	status, hdr, _, err := c.doRaw(req)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, fmt.Errorf("github returned status %d", status)
	}
	raw := hdr.Get("X-OAuth-Scopes")
	if raw == "" {
		return []string{}, nil
	}
	var scopes []string
	for _, s := range bytes.Split([]byte(raw), []byte(",")) {
		scopes = append(scopes, string(bytes.TrimSpace(s)))
	}
	return scopes, nil
}

// ListCodespaces GET /user/codespaces
func (c *CodespacesClient) ListCodespaces(ctx context.Context, token string) (int, []byte, http.Header, error) {
	req, err := c.newRequest(ctx, "GET", "/user/codespaces", token, nil)
	if err != nil {
		return 0, nil, nil, err
	}
	return c.doRaw(req)
}

// GetCodespace GET /user/codespaces/{codespace_name}
func (c *CodespacesClient) GetCodespace(ctx context.Context, token, name string) (int, []byte, http.Header, error) {
	path := fmt.Sprintf("/user/codespaces/%s", url.PathEscape(name))
	req, err := c.newRequest(ctx, "GET", path, token, nil)
	if err != nil {
		return 0, nil, nil, err
	}
	return c.doRaw(req)
}

// CreateCodespace POST /user/codespaces
func (c *CodespacesClient) CreateCodespace(ctx context.Context, token string, body interface{}) (int, []byte, http.Header, error) {
	req, err := c.newRequest(ctx, "POST", "/user/codespaces", token, body)
	if err != nil {
		return 0, nil, nil, err
	}
	return c.doRaw(req)
}

// StartCodespace POST /user/codespaces/{codespace_name}/start
func (c *CodespacesClient) StartCodespace(ctx context.Context, token, name string) (int, []byte, http.Header, error) {
	path := fmt.Sprintf("/user/codespaces/%s/start", url.PathEscape(name))
	req, err := c.newRequest(ctx, "POST", path, token, nil)
	if err != nil {
		return 0, nil, nil, err
	}
	return c.doRaw(req)
}

// StopCodespace POST /user/codespaces/{codespace_name}/stop
func (c *CodespacesClient) StopCodespace(ctx context.Context, token, name string) (int, []byte, http.Header, error) {
	path := fmt.Sprintf("/user/codespaces/%s/stop", url.PathEscape(name))
	req, err := c.newRequest(ctx, "POST", path, token, nil)
	if err != nil {
		return 0, nil, nil, err
	}
	return c.doRaw(req)
}

// DeleteCodespace DELETE /user/codespaces/{codespace_name}
func (c *CodespacesClient) DeleteCodespace(ctx context.Context, token, name string) (int, []byte, http.Header, error) {
	path := fmt.Sprintf("/user/codespaces/%s", url.PathEscape(name))
	req, err := c.newRequest(ctx, "DELETE", path, token, nil)
	if err != nil {
		return 0, nil, nil, err
	}
	return c.doRaw(req)
}
