package github

import (
	"context"
	"net/http"
	"testing"

	"github.com/github/github-mcp-server/pkg/translations"
	"github.com/google/go-github/v74/github"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
)

func Test_ListCodespaces(t *testing.T) {
	mockClient := github.NewClient(nil)
	tool, _ := ListCodespaces(stubGetClientFn(mockClient), translations.NullTranslationHelper)

	assert.Equal(t, "list_codespaces", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.True(t, *tool.Annotation.ReadOnlyHint)
}

func Test_CreateCodespace(t *testing.T) {
	mockClient := github.NewClient(nil)
	tool, _ := CreateCodespace(stubGetClientFn(mockClient), translations.NullTranslationHelper)

	assert.Equal(t, "create_codespace", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "owner")
	assert.Contains(t, tool.InputSchema.Properties, "repo")
	assert.Contains(t, tool.InputSchema.Properties, "branch")
	assert.Contains(t, tool.InputSchema.Properties, "machine")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"owner", "repo"})
}

func Test_StopCodespace(t *testing.T) {
	mockClient := github.NewClient(nil)
	tool, _ := StopCodespace(stubGetClientFn(mockClient), translations.NullTranslationHelper)

	assert.Equal(t, "stop_codespace", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "name")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"name"})
}

func Test_DeleteCodespace(t *testing.T) {
	mockClient := github.NewClient(nil)
	tool, _ := DeleteCodespace(stubGetClientFn(mockClient), translations.NullTranslationHelper)

	assert.Equal(t, "delete_codespace", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "name")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"name"})
}