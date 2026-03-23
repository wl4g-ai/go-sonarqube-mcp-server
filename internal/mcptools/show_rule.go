package mcptools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/mark3labs/mcp-go/mcp"
	mcputils "sonarqube-mcp/internal/helpers"
)

func NewShowRuleMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"show_rule",
		"Show Rule — Show detailed information about a SonarQube rule.",
		json.RawMessage(
			`{
			"type": "object",
			"properties": {
				"key": {"type": "string", "description": "Rule key (e.g. javascript:EmptyBlock)."}
			},
			"required": ["key"],
			"additionalProperties": false
}`))
}

func ShowRuleHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	ruleKey := mcputils.GetOptionalString(args, "key")
	if ruleKey == "" {
		return mcp.NewToolResultError("key is required"), nil
	}

	params := url.Values{}
	params.Set("key", ruleKey)

	client := mcputils.NewSQClient()
	raw, err := client.DoGetRaw(ctx, "/api/rules/show", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Show rule failed: %v", err)), nil
	}

	return mcp.NewToolResultText(raw), nil
}
