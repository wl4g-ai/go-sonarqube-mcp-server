package mcptools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	mcputils "sonarqube-mcp/internal/helpers"
)

func NewSystemHealthMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"get_system_health",
		"Get System Health — Get health status (GREEN, YELLOW, RED) with causes and node details.",
		json.RawMessage(
			`{
			"type": "object",
			"properties": {},
			"additionalProperties": false
}`))
}

func SystemHealthHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client := mcputils.NewSQClient()
	raw, err := client.DoGetRaw(ctx, "/api/system/health", nil)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Get system health failed: %v", err)), nil
	}
	return mcp.NewToolResultText(raw), nil
}
