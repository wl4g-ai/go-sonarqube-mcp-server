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

type searchProjectsResponse struct {
	Paging     searchProjectsPaging `json:"paging"`
	Components []searchProjectsComp `json:"components"`
}
type searchProjectsPaging struct {
	PageIndex int `json:"pageIndex"`
	PageSize  int `json:"pageSize"`
	Total     int `json:"total"`
}
type searchProjectsComp struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

func NewSearchMyProjectsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"search_my_sonarqube_projects",
		"Search My SonarQube Projects — Find SonarQube projects. Supports searching by project name or key. Use this first when projectKey is unknown — most other tools require the project key from this response.",
		json.RawMessage(
			`{
			"type": "object",
			"properties": {
				"q": {"type": "string", "description": "Optional search query to filter projects by name (partial match) or key (exact match)."},
				"page": {"type": "integer", "description": "Optional page number. Defaults to 1.", "default": 1},
				"pageSize": {"type": "integer", "description": "Optional page size. Min 1, max 500. Defaults to 500.", "default": 500}
			},
			"additionalProperties": false
}`))
}

func SearchMyProjectsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	page := mcputils.GetIntOrDefault(args, "page", 1)
	pageSize := mcputils.GetIntOrDefault(args, "pageSize", 500)
	q := mcputils.GetOptionalString(args, "q")

	if pageSize <= 0 || pageSize > 500 {
		return mcp.NewToolResultError("pageSize must be greater than 0 and less than or equal to 500"), nil
	}

	params := url.Values{}
	params.Set("p", strconv.Itoa(page))
	params.Set("ps", strconv.Itoa(pageSize))
	if q != "" {
		params.Set("q", q)
	}
	if mcputils.IsCloud() {
		org := mcputils.GetSonarQubeOrg()
		if org != "" {
			params.Set("organization", org)
		}
	}

	client := mcputils.NewSQClient()
	var resp searchProjectsResponse
	if err := client.DoGet(ctx, "/api/components/search", params, &resp); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Search projects failed: %v", err)), nil
	}

	text := fmt.Sprintf("Found %d projects (page %d/%d):\n", len(resp.Components), resp.Paging.PageIndex, (resp.Paging.Total+resp.Paging.PageSize-1)/resp.Paging.PageSize)
	for _, c := range resp.Components {
		text += fmt.Sprintf("- %s (%s)\n", c.Name, c.Key)
	}
	if len(resp.Components) == 0 {
		text = "No projects found."
	}
	return mcp.NewToolResultText(text), nil
}
