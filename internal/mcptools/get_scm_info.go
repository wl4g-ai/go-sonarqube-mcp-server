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

func NewGetScmInfoMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"get_scm_info",
		"Get SCM Info — Get SCM information (author, date, revision) per line of a source file.",
		json.RawMessage(
			`{
			"type": "object",
			"properties": {
				"key": {"type": "string", "description": "File key (e.g. my_project:src/foo/Bar.php)."},
				"commits_by_line": {"type": "boolean", "description": "Show commits per line."},
				"from": {"type": "integer", "description": "First line (1-based)."},
				"to": {"type": "integer", "description": "Last line (1-based)."}
			},
			"required": ["key"],
			"additionalProperties": false
}`))
}

func GetScmInfoHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	fileKey := mcputils.GetOptionalString(args, "key")
	if fileKey == "" {
		return mcp.NewToolResultError("key is required"), nil
	}

	params := url.Values{}
	params.Set("key", fileKey)
	if mcputils.GetBoolOrDefault(args, "commits_by_line", false) {
		params.Set("commits_by_line", "true")
	}
	if v, ok := args["from"]; ok {
		if f, ok := v.(float64); ok {
			params.Set("from", strconv.Itoa(int(f)))
		}
	}
	if v, ok := args["to"]; ok {
		if f, ok := v.(float64); ok {
			params.Set("to", strconv.Itoa(int(f)))
		}
	}

	client := mcputils.NewSQClient()
	raw, err := client.DoGetRaw(ctx, "/api/sources/scm", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Get SCM info failed: %v", err)), nil
	}

	return mcp.NewToolResultText(raw), nil
}
