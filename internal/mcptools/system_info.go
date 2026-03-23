package mcptools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	mcputils "sonarqube-mcp/internal/helpers"
)

func NewSystemInfoMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"get_system_info",
		"Get System Info — Get detailed system configuration (JVM, DB, search indexes, settings). Requires Administer permission.",
		json.RawMessage(
			`{
			"type": "object",
			"properties": {},
			"additionalProperties": false
}`))
}

func SystemInfoHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client := mcputils.NewSQClient()
	raw, err := client.DoGetRaw(ctx, "/api/system/info", nil)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Get system info failed: %v", err)), nil
	}
	return mcp.NewToolResultText(raw), nil
}
