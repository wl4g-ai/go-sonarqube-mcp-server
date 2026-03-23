package mcptools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	mcputils "sonarqube-mcp/internal/helpers"
)

type listQGsResponse struct {
	QualityGates []listQGsEntry `json:"qualitygates"`
}
type listQGsEntry struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func NewListQualityGatesMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"list_quality_gates",
		"List Quality Gates — List all quality gates defined on the SonarQube instance.",
		json.RawMessage(
			`{
			"type": "object",
			"properties": {},
			"additionalProperties": false
}`))
}

func ListQualityGatesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client := mcputils.NewSQClient()
	var resp listQGsResponse
	if err := client.DoGet(ctx, "/api/qualitygates/list", nil, &resp); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("List quality gates failed: %v", err)), nil
	}

	text := "Quality Gates:\n"
	for _, qg := range resp.QualityGates {
		text += fmt.Sprintf("- %s (id=%d)\n", qg.Name, qg.ID)
	}
	return mcp.NewToolResultText(text), nil
}
