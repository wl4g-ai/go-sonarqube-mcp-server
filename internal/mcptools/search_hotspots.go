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

type searchHotspotsResponse struct {
	Paging   searchHotspotsPaging    `json:"paging"`
	Hotspots []searchHotspotsHotspot `json:"hotspots"`
}
type searchHotspotsPaging struct {
	PageIndex int `json:"pageIndex"`
	PageSize  int `json:"pageSize"`
	Total     int `json:"total"`
}
type searchHotspotsHotspot struct {
	Key        string `json:"key"`
	Component  string `json:"component"`
	Project    string `json:"project"`
	Status     string `json:"status"`
	Resolution string `json:"resolution"`
	Message    string `json:"message"`
	Line       int    `json:"line"`
}

func NewSearchSecurityHotspotsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"search_security_hotspots",
		"Search Security Hotspots — Search security hotspots in a project. Either projectKey or hotspotKeys must be provided.",
		json.RawMessage(
			`{
			"type": "object",
			"properties": {
				"projectKey": {"type": "string", "description": "SonarQube project key."},
				"hotspotKeys": {"type": "array", "items": {"type": "string"}, "description": "Specific hotspot keys."},
				"branch": {"type": "string", "description": "Long-lived branch name."},
				"pullRequest": {"type": "string", "description": "Pull request key."},
				"files": {"type": "array", "items": {"type": "string"}, "description": "Component keys to filter."},
				"status": {"type": "string", "enum": ["TO_REVIEW", "REVIEWED"], "description": "Hotspot status."},
				"resolution": {"type": "string", "enum": ["FIXED", "SAFE", "ACKNOWLEDGED"], "description": "Hotspot resolution (for REVIEWED)."},
				"sinceLeakPeriod": {"type": "boolean", "description": "Only hotspots since leak period."},
				"onlyMine": {"type": "boolean", "description": "Only hotspots assigned to current user."},
				"p": {"type": "integer", "description": "Page number. Defaults to 1.", "default": 1},
				"ps": {"type": "integer", "description": "Page size. Max 500. Defaults to 100.", "default": 100}
			},
			"additionalProperties": false
}`))
}

func SearchSecurityHotspotsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	params := url.Values{}

	if projectKey := mcputils.OptionalProjectKey(args, "projectKey"); projectKey != "" {
		params.Set("project", projectKey)
	}
	if keys := mcputils.GetStringArray(args, "hotspotKeys"); len(keys) > 0 {
		params.Set("hotspots", strings.Join(keys, ","))
	}
	if branch := mcputils.GetOptionalString(args, "branch"); branch != "" {
		params.Set("branch", branch)
	}
	if pr := mcputils.GetOptionalString(args, "pullRequest"); pr != "" {
		params.Set("pullRequest", pr)
	}
	if files := mcputils.GetStringArray(args, "files"); len(files) > 0 {
		params.Set("files", strings.Join(files, ","))
	}
	if status := mcputils.GetOptionalString(args, "status"); status != "" {
		params.Set("status", status)
	}
	if resolution := mcputils.GetOptionalString(args, "resolution"); resolution != "" {
		params.Set("resolution", resolution)
	}
	if mcputils.GetBoolOrDefault(args, "sinceLeakPeriod", false) {
		params.Set("sinceLeakPeriod", "true")
	}
	if mcputils.GetBoolOrDefault(args, "onlyMine", false) {
		params.Set("onlyMine", "true")
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
	var resp searchHotspotsResponse
	if err := client.DoGet(ctx, "/api/hotspots/search", params, &resp); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Search hotspots failed: %v", err)), nil
	}

	text := fmt.Sprintf("Found %d security hotspots:\n", resp.Paging.Total)
	for _, h := range resp.Hotspots {
		line := ""
		if h.Line > 0 {
			line = fmt.Sprintf(":%d", h.Line)
		}
		text += fmt.Sprintf("- %s [%s] %s%s — %s", h.Key, h.Status, h.Component, line, h.Message)
		if h.Resolution != "" {
			text += fmt.Sprintf(" (%s)", h.Resolution)
		}
		text += "\n"
	}
	if len(resp.Hotspots) == 0 {
		text = "No security hotspots found."
	}
	return mcp.NewToolResultText(text), nil
}
