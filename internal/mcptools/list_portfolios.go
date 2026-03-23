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

type listViewsResponse struct {
	Views  []listViewsEntry `json:"views"`
	Paging listViewsPaging  `json:"paging"`
}
type listViewsEntry struct {
	Key  string `json:"key"`
	Name string `json:"name"`
	Desc string `json:"desc"`
}
type listViewsPaging struct {
	PageIndex int `json:"pageIndex"`
	PageSize  int `json:"pageSize"`
	Total     int `json:"total"`
}

func NewListPortfoliosMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"list_portfolios",
		"List Portfolios — List portfolios/views. On SonarQube Cloud, provides enterprise portfolios. On Server, provides standard views.",
		json.RawMessage(
			`{
			"type": "object",
			"properties": {
				"q": {"type": "string", "description": "Optional search query to filter by name."},
				"favorite": {"type": "boolean", "description": "Only show favorite portfolios."},
				"pageIndex": {"type": "integer", "description": "1-based page index.", "default": 1},
				"pageSize": {"type": "integer", "description": "Page size.", "default": 100}
			},
			"additionalProperties": false
}`))
}

func ListPortfoliosHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	params := url.Values{}
	if q := mcputils.GetOptionalString(args, "q"); q != "" {
		params.Set("q", q)
	}
	if mcputils.GetBoolOrDefault(args, "favorite", false) {
		params.Set("favorite", "true")
	}
	pageIndex := mcputils.GetIntOrDefault(args, "pageIndex", 1)
	pageSize := mcputils.GetIntOrDefault(args, "pageSize", 100)
	params.Set("p", strconv.Itoa(pageIndex))
	params.Set("ps", strconv.Itoa(pageSize))

	client := mcputils.NewSQClient()
	var resp listViewsResponse
	path := "/api/views/list"
	if mcputils.IsCloud() {
		path = "/api/v2/views/list"
		if org := mcputils.GetSonarQubeOrg(); org != "" {
			params.Set("organization", org)
		}
	}
	if err := client.DoGet(ctx, path, params, &resp); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("List portfolios failed: %v", err)), nil
	}

	text := fmt.Sprintf("Portfolios (total %d):\n", resp.Paging.Total)
	for _, v := range resp.Views {
		text += fmt.Sprintf("- %s (%s): %s\n", v.Name, v.Key, v.Desc)
	}
	if len(resp.Views) == 0 {
		text = "No portfolios found."
	}
	return mcp.NewToolResultText(text), nil
}
