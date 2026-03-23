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

type searchCoverageResponse struct {
	Measures []searchCoverageMeasure `json:"measures"`
	Paging   searchCoveragePaging    `json:"paging"`
}
type searchCoverageMeasure struct {
	Component string `json:"component"`
	Metric    string `json:"metric"`
	Value     string `json:"value"`
}
type searchCoveragePaging struct {
	PageIndex int `json:"pageIndex"`
	PageSize  int `json:"pageSize"`
	Total     int `json:"total"`
}

func NewSearchFilesByCoverageMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"search_files_by_coverage",
		"Search Files by Coverage — Search files sorted by coverage ascending (worst first). Helps identify files needing test improvements.",
		json.RawMessage(
			`{
			"type": "object",
			"properties": {
				"projectKey": {"type": "string", "description": "SonarQube project key. Required unless a default is configured via SONARQUBE_PROJECT_KEY."},
				"branch": {"type": "string", "description": "Branch name."},
				"pullRequest": {"type": "string", "description": "Pull request key."},
				"maxCoverage": {"type": "number", "description": "Maximum coverage percentage (0-100). Only files below this threshold."},
				"pageIndex": {"type": "integer", "description": "1-based page index. Defaults to 1.", "default": 1},
				"pageSize": {"type": "integer", "description": "Page size. Max 500.", "default": 100}
			},
			"additionalProperties": false
}`))
}

func SearchFilesByCoverageHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	projectKey, err := mcputils.ResolveProjectKey(args, "projectKey")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	params := url.Values{}
	params.Set("component", projectKey)
	params.Set("metricKeys", "coverage")
	params.Set("s", "metricPeriod")
	params.Set("asc", "true") // worst coverage first

	pageIndex := mcputils.GetIntOrDefault(args, "pageIndex", 1)
	pageSize := mcputils.GetIntOrDefault(args, "pageSize", 100)
	params.Set("p", strconv.Itoa(pageIndex))
	params.Set("ps", strconv.Itoa(pageSize))

	if branch := mcputils.GetOptionalString(args, "branch"); branch != "" {
		params.Set("branch", branch)
	}
	if pr := mcputils.GetOptionalString(args, "pullRequest"); pr != "" {
		params.Set("pullRequest", pr)
	}

	client := mcputils.NewSQClient()
	var resp searchCoverageResponse
	if err := client.DoGet(ctx, "/api/measures/component_tree", params, &resp); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Search files by coverage failed: %v", err)), nil
	}

	maxCov := -1.0
	if v, ok := args["maxCoverage"]; ok {
		if f, ok := v.(float64); ok {
			maxCov = f
		}
	}

	text := fmt.Sprintf("Files by coverage (worst first) for %s:\n", projectKey)
	count := 0
	for _, m := range resp.Measures {
		cov := 0.0
		if _, err := fmt.Sscanf(m.Value, "%f", &cov); err == nil {
			if maxCov >= 0 && cov > maxCov {
				continue
			}
		}
		text += fmt.Sprintf("- %s: %s%%\n", m.Component, m.Value)
		count++
	}
	if count == 0 {
		text += "(no files found)\n"
	}
	return mcp.NewToolResultText(text), nil
}
