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

type searchMetricsResponse struct {
	Metrics []searchMetricsEntry `json:"metrics"`
	Paging  searchMetricsPaging  `json:"paging"`
}
type searchMetricsEntry struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
}
type searchMetricsPaging struct {
	PageIndex int `json:"pageIndex"`
	PageSize  int `json:"pageSize"`
	Total     int `json:"total"`
}

func NewSearchMetricsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"search_metrics",
		"Search Metrics — Search for available metrics on the SonarQube instance.",
		json.RawMessage(
			`{
			"type": "object",
			"properties": {
				"p": {"type": "integer", "description": "Page number. Defaults to 1.", "default": 1},
				"ps": {"type": "integer", "description": "Page size. Max 500. Defaults to 100.", "default": 100}
			},
			"additionalProperties": false
}`))
}

func SearchMetricsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	page := mcputils.GetIntOrDefault(args, "p", 1)
	pageSize := mcputils.GetIntOrDefault(args, "ps", 100)

	params := url.Values{}
	params.Set("p", strconv.Itoa(page))
	params.Set("ps", strconv.Itoa(pageSize))

	client := mcputils.NewSQClient()
	var resp searchMetricsResponse
	if err := client.DoGet(ctx, "/api/metrics/search", params, &resp); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Search metrics failed: %v", err)), nil
	}

	text := fmt.Sprintf("Metrics (page %d, total %d):\n", page, resp.Paging.Total)
	for _, m := range resp.Metrics {
		text += fmt.Sprintf("- %s: %s [%s] — %s\n", m.Key, m.Name, m.Type, m.Description)
	}
	return mcp.NewToolResultText(text), nil
}
