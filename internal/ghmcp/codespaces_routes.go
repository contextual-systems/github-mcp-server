package ghmcp

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/github/github-mcp-server/pkg/github"
)

// RegisterCodespacesRoutes registers endpoints under /api/codespaces.
// Keep handlers small here; use pkg/github for the API client logic.
// The MCP should wire token/session retrieval into the extractToken function or replace it.
func RegisterCodespacesRoutes(mux *http.ServeMux, client *github.CodespacesClient) {
	mux.HandleFunc("/api/codespaces", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleListCodespaces(w, r, client)
		case http.MethodPost:
			handleCreateCodespace(w, r, client)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/codespaces/", func(w http.ResponseWriter, r *http.Request) {
		rest := strings.TrimPrefix(r.URL.Path, "/api/codespaces/")
		if rest == "" {
			http.Error(w, "missing codespace name", http.StatusBadRequest)
			return
		}
		if strings.HasSuffix(rest, "/start") && r.Method == http.MethodPost {
			name := strings.TrimSuffix(rest, "/start")
			handleStartCodespace(w, r, client, name)
			return
		}
		if strings.HasSuffix(rest, "/stop") && r.Method == http.MethodPost {
			name := strings.TrimSuffix(rest, "/stop")
			handleStopCodespace(w, r, client, name)
			return
		}
		switch r.Method {
		case http.MethodGet:
			handleGetCodespace(w, r, client, rest)
		case http.MethodDelete:
			handleDeleteCodespace(w, r, client, rest)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

func extractToken(r *http.Request) string {
	// Default extraction: Authorization header. In-proc MCP code should replace this
	// with secure token/session retrieval and not require callers to send raw PATs.
	auth := r.Header.Get("Authorization")
	if auth == "" {
		auth = r.Header.Get("X-Github-Token")
	}
	if auth == "" {
		return ""
	}
	parts := strings.Fields(auth)
	if len(parts) == 1 {
		return parts[0]
	}
	return parts[1]
}

func ensureScopes(ctx context.Context, client *github.CodespacesClient, token string, required []string) error {
	scopes, err := client.GetTokenScopes(ctx, token)
	if err != nil {
		return err
	}
	have := map[string]bool{}
	for _, s := range scopes {
		have[strings.ToLower(strings.TrimSpace(s))] = true
	}
	for _, r := range required {
		if !have[strings.ToLower(r)] {
			return &InsufficientScopeError{Required: required, Granted: scopes}
		}
	}
	return nil
}

type InsufficientScopeError struct {
	Required []string
	Granted  []string
}

func (e *InsufficientScopeError) Error() string { return "insufficient scopes" }

func writeProxyResponse(w http.ResponseWriter, status int, hdr http.Header, body []byte) {
	if ct := hdr.Get("Content-Type"); ct != "" {
		w.Header().Set("Content-Type", ct)
	} else {
		w.Header().Set("Content-Type", "application/json")
	}
	w.WriteHeader(status)
	if len(body) > 0 {
		_, _ = w.Write(body)
	}
}

func handleListCodespaces(w http.ResponseWriter, r *http.Request, client *github.CodespacesClient) {
	ctx := r.Context()
	token := extractToken(r)
	if token == "" {
		http.Error(w, "missing Authorization token", http.StatusUnauthorized)
		return
	}
	if err := ensureScopes(ctx, client, token, []string{"codespaces"}); err != nil {
		if _, ok := err.(*InsufficientScopeError); ok {
			http.Error(w, "insufficient token scopes: requires codespaces", http.StatusForbidden)
			return
		}
		log.Printf("scope check error: %v", err)
		http.Error(w, "failed to validate token scopes", http.StatusInternalServerError)
		return
	}
	status, body, hdr, err := client.ListCodespaces(ctx, token)
	if err != nil {
		log.Printf("ListCodespaces error: %v", err)
		http.Error(w, "failed to call GitHub", http.StatusBadGateway)
		return
	}
	writeProxyResponse(w, status, hdr, body)
}

func handleGetCodespace(w http.ResponseWriter, r *http.Request, client *github.CodespacesClient, name string) {
	ctx := r.Context()
	token := extractToken(r)
	if token == "" {
		http.Error(w, "missing Authorization token", http.StatusUnauthorized)
		return
	}
	if err := ensureScopes(ctx, client, token, []string{"codespaces"}); err != nil {
		if _, ok := err.(*InsufficientScopeError); ok {
			http.Error(w, "insufficient token scopes: requires codespaces", http.StatusForbidden)
			return
		}
		log.Printf("scope check error: %v", err)
		http.Error(w, "failed to validate token scopes", http.StatusInternalServerError)
		return
	}
	status, body, hdr, err := client.GetCodespace(ctx, token, name)
	if err != nil {
		log.Printf("GetCodespace error: %v", err)
		http.Error(w, "failed to call GitHub", http.StatusBadGateway)
		return
	}
	writeProxyResponse(w, status, hdr, body)
}

func handleCreateCodespace(w http.ResponseWriter, r *http.Request, client *github.CodespacesClient) {
	ctx := r.Context()
	token := extractToken(r)
	if token == "" {
		http.Error(w, "missing Authorization token", http.StatusUnauthorized)
		return
	}
	if err := ensureScopes(ctx, client, token, []string{"codespaces"}); err != nil {
		if _, ok := err.(*InsufficientScopeError); ok {
			http.Error(w, "insufficient token scopes: requires codespaces", http.StatusForbidden)
			return
		}
		log.Printf("scope check error: %v", err)
		http.Error(w, "failed to validate token scopes", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()
	var body interface{}
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if len(bs) > 0 {
		if err := json.Unmarshal(bs, &body); err != nil {
			http.Error(w, "invalid json body", http.StatusBadRequest)
			return
		}
	}
	status, respBody, hdr, err := client.CreateCodespace(ctx, token, body)
	if err != nil {
		log.Printf("CreateCodespace error: %v", err)
		http.Error(w, "failed to call GitHub", http.StatusBadGateway)
		return
	}
	writeProxyResponse(w, status, hdr, respBody)
}

func handleStartCodespace(w http.ResponseWriter, r *http.Request, client *github.CodespacesClient, name string) {
	ctx := r.Context()
	token := extractToken(r)
	if token == "" {
		http.Error(w, "missing Authorization token", http.StatusUnauthorized)
		return
	}
	if err := ensureScopes(ctx, client, token, []string{"codespaces"}); err != nil {
		if _, ok := err.(*InsufficientScopeError); ok {
			http.Error(w, "insufficient token scopes: requires codespaces", http.StatusForbidden)
			return
		}
		log.Printf("scope check error: %v", err)
		http.Error(w, "failed to validate token scopes", http.StatusInternalServerError)
		return
	}
	status, body, hdr, err := client.StartCodespace(ctx, token, name)
	if err != nil {
		log.Printf("StartCodespace error: %v", err)
		http.Error(w, "failed to call GitHub", http.StatusBadGateway)
		return
	}
	writeProxyResponse(w, status, hdr, body)
}

func handleStopCodespace(w http.ResponseWriter, r *http.Request, client *github.CodespacesClient, name string) {
	ctx := r.Context()
	token := extractToken(r)
	if token == "" {
		http.Error(w, "missing Authorization token", http.StatusUnauthorized)
		return
	}
	if err := ensureScopes(ctx, client, token, []string{"codespaces"}); err != nil {
		if _, ok := err.(*InsufficientScopeError); ok {
			http.Error(w, "insufficient token scopes: requires codespaces", http.StatusForbidden)
			return
		}
		log.Printf("scope check error: %v", err)
		http.Error(w, "failed to validate token scopes", http.StatusInternalServerError)
		return
	}
	status, body, hdr, err := client.StopCodespace(ctx, token, name)
	if err != nil {
		log.Printf("StopCodespace error: %v", err)
		http.Error(w, "failed to call GitHub", http.StatusBadGateway)
		return
	}
	writeProxyResponse(w, status, hdr, body)
}

func handleDeleteCodespace(w http.ResponseWriter, r *http.Request, client *github.CodespacesClient, name string) {
	ctx := r.Context()
	token := extractToken(r)
	if token == "" {
		http.Error(w, "missing Authorization token", http.StatusUnauthorized)
		return
	}
	if err := ensureScopes(ctx, client, token, []string{"codespaces"}); err != nil {
		if _, ok := err.(*InsufficientScopeError); ok {
			http.Error(w, "insufficient token scopes: requires codespaces", http.StatusForbidden)
			return
		}
		log.Printf("scope check error: %v", err)
		http.Error(w, "failed to validate token scopes", http.StatusInternalServerError)
		return
	}
	status, body, hdr, err := client.DeleteCodespace(ctx, token, name)
	if err != nil {
		log.Printf("DeleteCodespace error: %v", err)
		http.Error(w, "failed to call GitHub", http.StatusBadGateway)
		return
	}
	writeProxyResponse(w, status, hdr, body)
}
