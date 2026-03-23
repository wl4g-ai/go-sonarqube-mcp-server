package mcptools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/mark3labs/mcp-go/mcp"
	mcputils "sonarqube-mcp/internal/helpers"
)

type listPRsResponse struct {
	PullRequests []listPREntry `json:"pullRequests"`
}
type listPREntry struct {
	Key    string `json:"key"`
	Title  string `json:"title"`
	Branch string `json:"branch"`
	Base   string `json:"base"`
	Status string `json:"status"`
}

func NewListPullRequestsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"list_pull_requests",
		"List Pull Requests — List all pull requests for a project with their key/ID and source branch.",
		json.RawMessage(
			`{
			"type": "object",
			"properties": {
				"projectKey": {"type": "string", "description": "SonarQube project key. Required unless a default is configured via SONARQUBE_PROJECT_KEY."}
			},
			"additionalProperties": false
}`))
}

func ListPullRequestsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	projectKey, err := mcputils.ResolveProjectKey(args, "projectKey")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	params := url.Values{}
	params.Set("project", projectKey)

	client := mcputils.NewSQClient()
	var resp listPRsResponse
	if err := client.DoGet(ctx, "/api/project_pull_requests/list", params, &resp); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("List pull requests failed: %v", err)), nil
	}

	text := fmt.Sprintf("Pull requests for project %s:\n", projectKey)
	for _, pr := range resp.PullRequests {
		text += fmt.Sprintf("- %s: %s → %s [%s] (%s)\n", pr.Key, pr.Branch, pr.Base, pr.Status, pr.Title)
	}
	if len(resp.PullRequests) == 0 {
		text = "No pull requests found."
	}
	return mcp.NewToolResultText(text), nil
}
