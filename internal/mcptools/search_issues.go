package mcptools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	mcputils "sonarqube-mcp/internal/helpers"
)

type searchIssuesResponse struct {
	Paging searchIssuesPaging  `json:"paging"`
	Issues []searchIssuesIssue `json:"issues"`
}
type searchIssuesPaging struct {
	PageIndex int `json:"pageIndex"`
	PageSize  int `json:"pageSize"`
	Total     int `json:"total"`
}
type searchIssuesIssue struct {
	Key       string `json:"key"`
	Rule      string `json:"rule"`
	Severity  string `json:"severity"`
	Status    string `json:"status"`
	Component string `json:"component"`
	Message   string `json:"message"`
	Line      int    `json:"line"`
}

func NewSearchIssuesMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"search_sonar_issues_in_projects",
		"Search Sonar Issues in Projects — Search for issues (bugs, vulnerabilities, code smells). Filter by severity, software quality impact, status, files, projects.",
		json.RawMessage(
			`{
			"type": "object",
			"properties": {
				"projects": {"type": "array", "items": {"type": "string"}, "description": "Project keys to search in."},
				"files": {"type": "array", "items": {"type": "string"}, "description": "Component keys (files/directories) to filter."},
				"branch": {"type": "string", "description": "Long-lived branch name."},
				"pullRequest": {"type": "string", "description": "Pull request key."},
				"severities": {"type": "array", "items": {"type": "string", "enum": ["INFO", "MINOR", "MAJOR", "CRITICAL", "BLOCKER"]}, "description": "Issue severities to filter."},
				"impactSoftwareQualities": {"type": "array", "items": {"type": "string", "enum": ["MAINTAINABILITY", "RELIABILITY", "SECURITY"]}, "description": "Software quality impacts."},
				"issueStatuses": {"type": "array", "items": {"type": "string", "enum": ["OPEN", "CONFIRMED", "FALSE_POSITIVE", "ACCEPTED", "FIXED", "IN_SANDBOX"]}, "description": "Issue statuses to filter."},
				"issueKey": {"type": "array", "items": {"type": "string"}, "description": "Specific issue keys."},
				"p": {"type": "integer", "description": "Page number. Defaults to 1.", "default": 1},
				"ps": {"type": "integer", "description": "Page size. Max 500. Defaults to 100.", "default": 100}
			},
			"additionalProperties": false
}`))
}

func SearchIssuesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	params := url.Values{}

	if projects := mcputils.GetStringArray(args, "projects"); len(projects) > 0 {
		params.Set("projects", strings.Join(projects, ","))
	}
	if files := mcputils.GetStringArray(args, "files"); len(files) > 0 {
		params.Set("files", strings.Join(files, ","))
	}
	if branch := mcputils.GetOptionalString(args, "branch"); branch != "" {
		params.Set("branch", branch)
	}
	if pr := mcputils.GetOptionalString(args, "pullRequest"); pr != "" {
		params.Set("pullRequest", pr)
	}
	if severities := mcputils.GetStringArray(args, "severities"); len(severities) > 0 {
		params.Set("severities", strings.Join(severities, ","))
	}
	if qualities := mcputils.GetStringArray(args, "impactSoftwareQualities"); len(qualities) > 0 {
		params.Set("impactSoftwareQualities", strings.Join(qualities, ","))
	}
	if statuses := mcputils.GetStringArray(args, "issueStatuses"); len(statuses) > 0 {
		params.Set("issueStatuses", strings.Join(statuses, ","))
	}
	if keys := mcputils.GetStringArray(args, "issueKey"); len(keys) > 0 {
		params.Set("issues", strings.Join(keys, ","))
	}
	if mcputils.IsCloud() {
		if org := mcputils.GetSonarQubeOrg(); org != "" {
			params.Set("organization", org)
		}
	}

	page := mcputils.GetIntOrDefault(args, "p", 1)
	pageSize := mcputils.GetIntOrDefault(args, "ps", 100)
	params.Set("p", strconv.Itoa(page))
	params.Set("ps", strconv.Itoa(pageSize))

	client := mcputils.NewSQClient()
	var resp searchIssuesResponse
	if err := client.DoGet(ctx, "/api/issues/search", params, &resp); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Search issues failed: %v", err)), nil
	}

	text := fmt.Sprintf("Found %d issues:\n", resp.Paging.Total)
	for _, issue := range resp.Issues {
		line := ""
		if issue.Line > 0 {
			line = fmt.Sprintf(":%d", issue.Line)
		}
		text += fmt.Sprintf("- %s [%s] %s%s — %s (%s)\n", issue.Key, issue.Severity, issue.Component, line, issue.Message, issue.Status)
	}
	if len(resp.Issues) == 0 {
		text = "No issues found."
	}
	return mcp.NewToolResultText(text), nil
}
