package mcptools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/mark3labs/mcp-go/mcp"
	mcputils "sonarqube-mcp/internal/helpers"
)

func NewGetFileCoverageDetailsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"get_file_coverage_details",
		"Get File Coverage Details — Get line-by-line coverage for a file: uncovered lines, partially covered branches.",
		json.RawMessage(
			`{
			"type": "object",
			"properties": {
				"key": {"type": "string", "description": "File key (e.g. my_project:src/foo/Bar.java)."},
				"branch": {"type": "string", "description": "Branch name."},
				"pullRequest": {"type": "string", "description": "Pull request key."}
			},
			"required": ["key"],
			"additionalProperties": false
}`))
}

func GetFileCoverageDetailsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
	raw, err := client.DoGetRaw(ctx, "/api/sources/coverage", params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Get file coverage failed: %v", err)), nil
	}

	return mcp.NewToolResultText(raw), nil
}
