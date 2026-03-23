package mcptools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/mark3labs/mcp-go/mcp"
	mcputils "sonarqube-mcp/internal/helpers"
)

func NewShowSecurityHotspotMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"show_security_hotspot",
		"Show Security Hotspot — Get detailed information about a specific security hotspot (rule details, code context, flows, comments).",
		json.RawMessage(
			`{
			"type": "object",
			"properties": {
				"hotspotKey": {"type": "string", "description": "Security hotspot key."}
			},
			"required": ["hotspotKey"],
			"additionalProperties": false
}`))
}

func ShowSecurityHotspotHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	hotspotKey := mcputils.GetOptionalString(args, "hotspotKey")
	if hotspotKey == "" {
		return mcp.NewToolResultError("hotspotKey is required"), nil
	}

	params := url.Values{}
	params.Set("hotspot", hotspotKey)

	client := mcputils.NewSQClient()
	raw, err := client.DoGetRaw(ctx, "/api/hotspots/show", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Show hotspot failed: %v", err)), nil
	}

	return mcp.NewToolResultText(raw), nil
}
