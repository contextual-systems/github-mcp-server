package github

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/github/github-mcp-server/pkg/translations"
	"github.com/google/go-github/v74/github"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// ListCodespaces creates a tool to list all codespaces for the authenticated user
func ListCodespaces(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("list_codespaces",
			mcp.WithDescription(t("TOOL_LIST_CODESPACES_DESCRIPTION", "List all codespaces for the authenticated user")),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{
				Title:        t("TOOL_LIST_CODESPACES_USER_TITLE", "List codespaces"),
				ReadOnlyHint: ToBoolPtr(true),
			}),
		), func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub client: %w", err)
			}

			list, _, err := client.Codespaces.List(ctx, nil)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("failed to list codespaces: %v", err)), nil
			}

			r, err := json.Marshal(list)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}

			return mcp.NewToolResultText(string(r)), nil
		}
}

// CreateCodespace creates a tool to create a new codespace in a repository
func CreateCodespace(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("create_codespace",
			mcp.WithDescription(t("TOOL_CREATE_CODESPACE_DESCRIPTION", "Create a new codespace for a repository")),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:        t("TOOL_CREATE_CODESPACE_USER_TITLE", "Create codespace"),
			ReadOnlyHint: ToBoolPtr(false),
			}),
			mcp.WithString("owner",
				mcp.Required(),
				mcp.Description(DescriptionRepositoryOwner),
			),
			mcp.WithString("repo",
				mcp.Required(),
				mcp.Description(DescriptionRepositoryName),
			),
			mcp.WithString("branch",
				mcp.Description("The branch to create the codespace from"),
			),
			mcp.WithString("machine",
				mcp.Description("The machine type to use for this codespace"),
			),
		), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			owner, err := RequiredParam[string](req, "owner")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			repo, err := RequiredParam[string](req, "repo")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			branch, _ := OptionalParam[string](req, "branch")
			machine, _ := OptionalParam[string](req, "machine")

			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub client: %w", err)
			}

			opts := &github.CreateCodespaceOptions{}
			if branch != "" {
				opts.Ref = &branch
			}
			if machine != "" {
				opts.Machine = &machine
			}

			codespace, _, err := client.Codespaces.CreateInRepo(ctx, owner, repo, opts)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("failed to create codespace: %v", err)), nil
			}

			r, err := json.Marshal(codespace)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}

			return mcp.NewToolResultText(string(r)), nil
		}
}

// StopCodespace creates a tool to stop a running codespace
func StopCodespace(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("stop_codespace",
			mcp.WithDescription(t("TOOL_STOP_CODESPACE_DESCRIPTION", "Stop a running codespace")),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:        t("TOOL_STOP_CODESPACE_USER_TITLE", "Stop codespace"),
			ReadOnlyHint: ToBoolPtr(false),
			}),
			mcp.WithString("name",
				mcp.Required(),
				mcp.Description("The name of the codespace to stop"),
			),
		), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			name, err := RequiredParam[string](req, "name")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub client: %w", err)
			}

			_, _, err = client.Codespaces.Stop(ctx, name)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("failed to stop codespace: %v", err)), nil
			}

			return mcp.NewToolResultText("Codespace stopped successfully"), nil
		}
}

// DeleteCodespace creates a tool to delete a codespace
func DeleteCodespace(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("delete_codespace",
			mcp.WithDescription(t("TOOL_DELETE_CODESPACE_DESCRIPTION", "Delete a codespace")),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:        t("TOOL_DELETE_CODESPACE_USER_TITLE", "Delete codespace"),
			ReadOnlyHint: ToBoolPtr(false),
			}),
			mcp.WithString("name",
				mcp.Required(),
				mcp.Description("The name of the codespace to delete"),
			),
		), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			name, err := RequiredParam[string](req, "name")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub client: %w", err)
			}

			_, err = client.Codespaces.Delete(ctx, name)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("failed to delete codespace: %v", err)), nil
			}

			return mcp.NewToolResultText("Codespace deleted successfully"), nil
		}
}
