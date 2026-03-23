package mcptools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	mcputils "sonarqube-mcp/internal/helpers"
)

type componentMeasuresResponse struct {
	Component componentMeasuresComp `json:"component"`
}
type componentMeasuresComp struct {
	Key      string                     `json:"key"`
	Name     string                     `json:"name"`
	Measures []componentMeasuresMeasure `json:"measures"`
}
type componentMeasuresMeasure struct {
	Metric    string `json:"metric"`
	Value     string `json:"value"`
	BestValue bool   `json:"bestValue"`
}

func NewGetComponentMeasuresMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"get_component_measures",
		"Get Component Measures — Get project measures like ncloc, complexity, violations, coverage, etc.",
		json.RawMessage(
			`{
			"type": "object",
			"properties": {
				"projectKey": {"type": "string", "description": "SonarQube project key. Required unless a default is configured via SONARQUBE_PROJECT_KEY."},
				"branch": {"type": "string", "description": "Long-lived branch name."},
				"pullRequest": {"type": "string", "description": "Pull request key."},
				"metricKeys": {"type": "array", "items": {"type": "string"}, "description": "Metric keys (e.g. ncloc, complexity, violations, coverage)."}
			},
			"additionalProperties": false
}`))
}

func GetComponentMeasuresHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	projectKey, err := mcputils.ResolveProjectKey(args, "projectKey")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	metricKeys := mcputils.GetStringArray(args, "metricKeys")

	params := url.Values{}
	params.Set("component", projectKey)
	if len(metricKeys) > 0 {
		params.Set("metricKeys", strings.Join(metricKeys, ","))
	}
	if branch := mcputils.GetOptionalString(args, "branch"); branch != "" {
		params.Set("branch", branch)
	}
	if pr := mcputils.GetOptionalString(args, "pullRequest"); pr != "" {
		params.Set("pullRequest", pr)
	}

	client := mcputils.NewSQClient()
	var resp componentMeasuresResponse
	if err := client.DoGet(ctx, "/api/measures/component", params, &resp); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Get component measures failed: %v", err)), nil
	}

	text := fmt.Sprintf("Measures for %s:\n", resp.Component.Name)
	for _, m := range resp.Component.Measures {
		best := ""
		if m.BestValue {
			best = " [best]"
		}
		text += fmt.Sprintf("- %s: %s%s\n", m.Metric, m.Value, best)
	}
	if len(resp.Component.Measures) == 0 {
		text += "(no measures returned)\n"
	}
	return mcp.NewToolResultText(text), nil
}
