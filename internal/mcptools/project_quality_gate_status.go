package mcptools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/mark3labs/mcp-go/mcp"
	mcputils "sonarqube-mcp/internal/helpers"
)

type projectStatusResponse struct {
	ProjectStatus projectStatusEntry `json:"projectStatus"`
}
type projectStatusEntry struct {
	Status     string                   `json:"status"`
	Conditions []projectStatusCondition `json:"conditions"`
	Periods    []projectStatusPeriod    `json:"periods"`
}
type projectStatusCondition struct {
	MetricKey      string `json:"metricKey"`
	Comparator     string `json:"comparator"`
	ErrorThreshold string `json:"errorThreshold"`
	Status         string `json:"status"`
	ActualValue    string `json:"actualValue"`
}
type projectStatusPeriod struct {
	Index int    `json:"index"`
	Mode  string `json:"mode"`
	Date  string `json:"date"`
}

func NewProjectQualityGateStatusMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"get_project_quality_gate_status",
		"Get Project Quality Gate Status — Get the quality gate status for a project. Requires analysisId, projectId, or projectKey.",
		json.RawMessage(
			`{
			"type": "object",
			"properties": {
				"analysisId": {"type": "string", "description": "Analysis ID."},
				"projectId": {"type": "string", "description": "Project ID."},
				"projectKey": {"type": "string", "description": "SonarQube project key."},
				"branch": {"type": "string", "description": "Branch name."},
				"pullRequest": {"type": "string", "description": "Pull request key."}
			},
			"additionalProperties": false
}`))
}

func ProjectQualityGateStatusHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	params := url.Values{}
	if analysisID := mcputils.GetOptionalString(args, "analysisId"); analysisID != "" {
		params.Set("analysisId", analysisID)
	}
	if projectID := mcputils.GetOptionalString(args, "projectId"); projectID != "" {
		params.Set("projectId", projectID)
	}
	if projectKey := mcputils.OptionalProjectKey(args, "projectKey"); projectKey != "" {
		params.Set("projectKey", projectKey)
	}
	if branch := mcputils.GetOptionalString(args, "branch"); branch != "" {
		params.Set("branch", branch)
	}
	if pr := mcputils.GetOptionalString(args, "pullRequest"); pr != "" {
		params.Set("pullRequest", pr)
	}

	client := mcputils.NewSQClient()
	var resp projectStatusResponse
	if err := client.DoGet(ctx, "/api/qualitygates/project_status", params, &resp); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Get project quality gate status failed: %v", err)), nil
	}

	ps := resp.ProjectStatus
	text := fmt.Sprintf("Quality Gate Status: %s\n", ps.Status)
	for _, c := range ps.Conditions {
		text += fmt.Sprintf("- %s: %s %s %s (actual=%s)\n", c.MetricKey, c.Comparator, c.ErrorThreshold, c.Status, c.ActualValue)
	}
	return mcp.NewToolResultText(text), nil
}
