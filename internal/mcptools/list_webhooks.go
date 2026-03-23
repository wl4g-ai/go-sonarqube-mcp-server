package mcptools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	mcputils "sonarqube-mcp/internal/helpers"
)

type listWebhooksResponse struct {
	Webhooks []listWebhooksEntry `json:"webhooks"`
}
type listWebhooksEntry struct {
	Key    string `json:"key"`
	Name   string `json:"name"`
	URL    string `json:"url"`
	Secret string `json:"secret"`
}

func NewListWebhooksMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"list_webhooks",
		"List Webhooks — List webhooks for the organization/instance or project. Requires Administer permission.",
		json.RawMessage(
			`{
			"type": "object",
			"properties": {
				"projectKey": {"type": "string", "description": "Optional project key to scope webhooks to a specific project."}
			},
			"additionalProperties": false
}`))
}

func ListWebhooksHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	projectKey := mcputils.OptionalProjectKey(args, "projectKey")

	client := mcputils.NewSQClient()

	var resp listWebhooksResponse
	path := "/api/webhooks/list"
	if projectKey != "" {
		path = "/api/project_webhooks/list?project=" + projectKey
	} else if mcputils.IsCloud() {
		if org := mcputils.GetSonarQubeOrg(); org != "" {
			path = "/api/webhooks/list?organization=" + org
		}
	}

	if err := client.DoGet(ctx, path, nil, &resp); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("List webhooks failed: %v", err)), nil
	}

	scope := "instance"
	if projectKey != "" {
		scope = "project " + projectKey
	}
	text := fmt.Sprintf("Webhooks for %s:\n", scope)
	for _, wh := range resp.Webhooks {
		text += fmt.Sprintf("- %s: %s → %s\n", wh.Name, wh.Key, wh.URL)
	}
	if len(resp.Webhooks) == 0 {
		text = "No webhooks found."
	}
	return mcp.NewToolResultText(text), nil
}
