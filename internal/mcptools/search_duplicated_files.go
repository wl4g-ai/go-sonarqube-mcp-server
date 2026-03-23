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

type searchDupFilesResponse struct {
	Duplications []searchDupFilesEntry `json:"duplications"`
	Paging       searchDupFilesPaging  `json:"paging"`
}
type searchDupFilesEntry struct {
	File             string `json:"file"`
	Project          string `json:"project"`
	DuplicatedBlocks int    `json:"duplicatedBlocks"`
	DuplicatedLines  int    `json:"duplicatedLines"`
}
type searchDupFilesPaging struct {
	PageIndex int `json:"pageIndex"`
	PageSize  int `json:"pageSize"`
	Total     int `json:"total"`
}

func NewSearchDuplicatedFilesMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"search_duplicated_files",
		"Search Duplicated Files — Search for files with code duplications in a project. Auto-fetches all pages by default.",
		json.RawMessage(
			`{
			"type": "object",
			"properties": {
				"projectKey": {"type": "string", "description": "SonarQube project key. Required unless a default is configured via SONARQUBE_PROJECT_KEY."},
				"branch": {"type": "string", "description": "Branch name."},
				"pullRequest": {"type": "string", "description": "Pull request key."},
				"pageSize": {"type": "integer", "description": "Page size. Max 500.", "default": 100},
				"pageIndex": {"type": "integer", "description": "1-based page index. Defaults to 1.", "default": 1}
			},
			"additionalProperties": false
}`))
}

func SearchDuplicatedFilesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
	var resp searchDupFilesResponse
	if err := client.DoGet(ctx, "/api/duplications/list", params, &resp); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Search duplicated files failed: %v", err)), nil
	}

	text := fmt.Sprintf("Found %d duplicated files:\n", resp.Paging.Total)
	for _, d := range resp.Duplications {
		text += fmt.Sprintf("- %s: %d blocks, %d duplicated lines\n", d.File, d.DuplicatedBlocks, d.DuplicatedLines)
	}
	if len(resp.Duplications) == 0 {
		text = "No duplicated files found."
	}
	return mcp.NewToolResultText(text), nil
}
