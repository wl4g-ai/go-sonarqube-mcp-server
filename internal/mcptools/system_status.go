package mcptools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	mcputils "sonarqube-mcp/internal/helpers"
)

func NewSystemStatusMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"get_system_status",
		"Get System Status — Get SonarQube Server status (STARTING, UP, DOWN, RESTARTING, DB_MIGRATION_NEEDED, DB_MIGRATION_RUNNING), version, and ID.",
		json.RawMessage(
			`{
			"type": "object",
			"properties": {},
			"additionalProperties": false
}`))
}

func SystemStatusHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client := mcputils.NewSQClient()
	raw, err := client.DoGetRaw(ctx, "/api/system/status", nil)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Get system status failed: %v", err)), nil
	}
	return mcp.NewToolResultText(raw), nil
}
