package mcptools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/mark3labs/mcp-go/mcp"
	mcputils "sonarqube-mcp/internal/helpers"
)

func NewGetRawSourceMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"get_raw_source",
		`Get Raw Source — Get source code as raw text. Requires "See Source Code" permission.`,
		json.RawMessage(`{
			"type": "object",
			"properties": {
				"key": {"type": "string", "description": "File key (e.g. my_project:src/foo/Bar.php)."},
				"branch": {"type": "string", "description": "Branch name."},
				"pullRequest": {"type": "string", "description": "Pull request key."}
			},
			"required": ["key"],
			"additionalProperties": false
		}`),
	)
}

func GetRawSourceHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	fileKey := mcputils.GetOptionalString(args, "key")
	if fileKey == "" {
		return mcp.NewToolResultError("key is required"), nil
	}

	params := url.Values{}
	params.Set("key", fileKey)
	if branch := mcputils.GetOptionalString(args, "branch"); branch != "" {
		params.Set("branch", branch)
	}
	if pr := mcputils.GetOptionalString(args, "pullRequest"); pr != "" {
		params.Set("pullRequest", pr)
	}

	client := mcputils.NewSQClient()
	raw, err := client.DoGetRaw(ctx, "/api/sources/raw", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Get raw source failed: %v", err)), nil
	}

	return mcp.NewToolResultText(raw), nil
}
