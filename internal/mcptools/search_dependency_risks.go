package mcptools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/mark3labs/mcp-go/mcp"
	mcputils "sonarqube-mcp/internal/helpers"
)

type searchDepRisksResponse struct {
	DependencyRisks []searchDepRisksEntry `json:"dependencyRisks"`
	Paging          searchDepRisksPaging  `json:"paging"`
}
type searchDepRisksEntry struct {
	Key       string `json:"key"`
	Component string `json:"component"`
	Rule      string `json:"rule"`
	Severity  string `json:"severity"`
	Status    string `json:"status"`
	Message   string `json:"message"`
	Release   string `json:"release"`
}
type searchDepRisksPaging struct {
	PageIndex int `json:"pageIndex"`
	PageSize  int `json:"pageSize"`
	Total     int `json:"total"`
}

func NewSearchDependencyRisksMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"search_dependency_risks",
		"Search Dependency Risks — Search for SCA/dependency risk issues paired with releases. Requires Advanced Security entitlement.",
		json.RawMessage(
			`{
			"type": "object",
			"properties": {
				"projectKey": {"type": "string", "description": "SonarQube project key. Required unless a default is configured via SONARQUBE_PROJECT_KEY."},
				"branch": {"type": "string", "description": "Branch name."},
				"pullRequest": {"type": "string", "description": "Pull request key."},
				"pageIndex": {"type": "integer", "description": "1-based page index. Defaults to 1.", "default": 1},
				"pageSize": {"type": "integer", "description": "Page size. Max 500. Defaults to 100.", "default": 100}
			},
			"additionalProperties": false
}`))
}

func SearchDependencyRisksHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	projectKey, err := mcputils.ResolveProjectKey(args, "projectKey")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	params := url.Values{}
	params.Set("project", projectKey)

	pageIndex := mcputils.GetIntOrDefault(args, "pageIndex", 1)
	pageSize := mcputils.GetIntOrDefault(args, "pageSize", 100)
	params.Set("p", strconv.Itoa(pageIndex))
	params.Set("ps", strconv.Itoa(pageSize))

	if branch := mcputils.GetOptionalString(args, "branch"); branch != "" {
		params.Set("branch", branch)
	}
	if pr := mcputils.GetOptionalString(args, "pullRequest"); pr != "" {
		params.Set("pullRequest", pr)
	}

	client := mcputils.NewSQClient()
	var resp searchDepRisksResponse
	if err := client.DoGet(ctx, "/api/dependency_risks/search", params, &resp); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Search dependency risks failed: %v", err)), nil
	}

	text := fmt.Sprintf("Dependency risks (total %d):\n", resp.Paging.Total)
	for _, d := range resp.DependencyRisks {
		text += fmt.Sprintf("- %s [%s] %s — %s (release: %s)\n", d.Key, d.Severity, d.Component, d.Message, d.Release)
	}
	if len(resp.DependencyRisks) == 0 {
		text = "No dependency risks found."
	}
	return mcp.NewToolResultText(text), nil
}
