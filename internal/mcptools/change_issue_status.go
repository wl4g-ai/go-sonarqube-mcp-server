package mcptools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/mark3labs/mcp-go/mcp"
	mcputils "sonarqube-mcp/internal/helpers"
)

func NewChangeIssueStatusMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"change_sonar_issue_status",
		`Change Sonar Issue Status — Change issue status to "accept", "falsepositive", or "reopen".`,
		json.RawMessage(`{
			"type": "object",
			"properties": {
				"key": {"type": "string", "description": "Issue key (e.g. AYxYgZQ0QHqQJQY5QHqQ)."},
				"status": {"type": "string", "enum": ["accept", "falsepositive", "reopen"], "description": "New status for the issue."}
			},
			"required": ["key", "status"],
			"additionalProperties": false
		}`),
	)
}

func ChangeIssueStatusHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if mcputils.IsReadOnly() {
		return mcp.NewToolResultError("Cannot change issue status: SONARQUBE_READ_ONLY is enabled"), nil
	}

	args := request.GetArguments()
	issueKey := mcputils.GetOptionalString(args, "key")
	status := mcputils.GetOptionalString(args, "status")

	if issueKey == "" || status == "" {
		return mcp.NewToolResultError("Both key and status are required"), nil
	}

	params := url.Values{}
	params.Set("issue", issueKey)
	params.Set("transition", status)

	client := mcputils.NewSQClient()
	if err := client.DoPost(ctx, "/api/issues/do_transition", params, nil); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Change issue status failed: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Issue %s status changed to %s.", issueKey, status)), nil
}
