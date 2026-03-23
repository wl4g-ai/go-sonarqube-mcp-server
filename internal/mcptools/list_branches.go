package mcptools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/mark3labs/mcp-go/mcp"
	mcputils "sonarqube-mcp/internal/helpers"
)

type listBranchesResponse struct {
	Branches []listBranchesEntry `json:"branches"`
}
type listBranchesEntry struct {
	Name   string `json:"name"`
	IsMain bool   `json:"isMain"`
	Type   string `json:"type"`
}

func NewListBranchesMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"list_branches",
		"List Branches — List long-lived branches for a project (main, develop, etc.).",
		json.RawMessage(
			`{
			"type": "object",
			"properties": {
				"projectKey": {"type": "string", "description": "SonarQube project key. Required unless a default is configured via SONARQUBE_PROJECT_KEY."}
			},
			"additionalProperties": false
}`))
}

func ListBranchesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	projectKey, err := mcputils.ResolveProjectKey(args, "projectKey")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	params := url.Values{}
	params.Set("project", projectKey)

	client := mcputils.NewSQClient()
	var resp listBranchesResponse
	if err := client.DoGet(ctx, "/api/project_branches/list", params, &resp); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("List branches failed: %v", err)), nil
	}

	text := fmt.Sprintf("Branches for project %s:\n", projectKey)
	for _, b := range resp.Branches {
		main := ""
		if b.IsMain {
			main = " (main)"
		}
		text += fmt.Sprintf("- %s [%s]%s\n", b.Name, b.Type, main)
	}
	return mcp.NewToolResultText(text), nil
}
