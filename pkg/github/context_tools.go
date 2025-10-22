package github

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/github/github-mcp-server/pkg/translations"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// GetMe creates a tool to get the authenticated user's profile
func GetMe(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("get_me",
		mcp.WithDescription(t("TOOL_GET_ME_DESCRIPTION", "Get my user profile")),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:        t("TOOL_GET_ME_USER_TITLE", "Get my profile"),
			ReadOnlyHint: ToBoolPtr(true),
		}),
	), func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		client, err := getClient(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get GitHub client: %w", err)
		}

		user, _, err := client.Users.Get(ctx, "")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to get user: %v", err)), nil
		}

		minimalUser := &MinimalUser{
			Login:      *user.Login,
			ID:         *user.ID,
			ProfileURL: user.GetHTMLURL(),
			AvatarURL:  user.GetAvatarURL(),
			Details: &UserDetails{
				PublicRepos:  user.GetPublicRepos(),
				PublicGists:  user.GetPublicGists(),
				Followers:    user.GetFollowers(),
				Following:    user.GetFollowing(),
				CreatedAt:    user.GetCreatedAt().String(),
				UpdatedAt:    user.GetUpdatedAt().String(),
			},
		}

		r, err := json.Marshal(minimalUser)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal response: %w", err)
		}

		return mcp.NewToolResultText(string(r)), nil
	}
}

// GetTeams creates a tool to get teams for a user
func GetTeams(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("get_teams",
		mcp.WithDescription(t("TOOL_GET_TEAMS_DESCRIPTION", "Get teams")),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:        t("TOOL_GET_TEAMS_USER_TITLE", "Get teams"),
			ReadOnlyHint: ToBoolPtr(true),
		}),
		mcp.WithString("user",
			mcp.Description("Username to get teams for. If not provided, uses the authenticated user."),
		),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		client, err := getClient(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get GitHub client: %w", err)
		}

		teams, _, err := client.Teams.ListUserTeams(ctx, nil)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to list teams: %v", err)), nil
		}

		r, err := json.Marshal(teams)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal response: %w", err)
		}

		return mcp.NewToolResultText(string(r)), nil
	}
}

// GetTeamMembers creates a tool to get team members
func GetTeamMembers(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("get_team_members",
		mcp.WithDescription(t("TOOL_GET_TEAM_MEMBERS_DESCRIPTION", "Get team members")),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:        t("TOOL_GET_TEAM_MEMBERS_USER_TITLE", "Get team members"),
			ReadOnlyHint: ToBoolPtr(true),
		}),
		mcp.WithString("org",
			mcp.Required(),
			mcp.Description("Organization login (owner) that contains the team."),
		),
		mcp.WithString("team_slug",
			mcp.Required(),
			mcp.Description("Team slug"),
		),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		org, err := RequiredParam[string](req, "org")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		teamSlug, err := RequiredParam[string](req, "team_slug")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		client, err := getClient(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get GitHub client: %w", err)
		}

		members, _, err := client.Teams.ListTeamMembersBySlug(ctx, org, teamSlug, nil)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to list team members: %v", err)), nil
		}

		r, err := json.Marshal(members)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal response: %w", err)
		}

		return mcp.NewToolResultText(string(r)), nil
	}
}