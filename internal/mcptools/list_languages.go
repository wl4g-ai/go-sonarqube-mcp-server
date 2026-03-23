package mcptools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	mcputils "sonarqube-mcp/internal/helpers"
)

type listLanguagesResponse struct {
	Languages []listLanguagesEntry `json:"languages"`
}
type listLanguagesEntry struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

func NewListLanguagesMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"list_languages",
		"List Languages — List all programming languages supported by this SonarQube instance.",
		json.RawMessage(
			`{
			"type": "object",
			"properties": {
				"q": {"type": "string", "description": "Optional pattern to match language keys or names."}
			},
			"additionalProperties": false
}`))
}

func ListLanguagesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	q := mcputils.GetOptionalString(args, "q")

	client := mcputils.NewSQClient()
	var resp listLanguagesResponse
	if err := client.DoGet(ctx, "/api/languages/list", nil, &resp); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("List languages failed: %v", err)), nil
	}

	text := fmt.Sprintf("Supported languages (%d):\n", len(resp.Languages))
	for _, l := range resp.Languages {
		if q != "" && l.Key != q && l.Name != q {
			continue
		}
		text += fmt.Sprintf("- %s: %s\n", l.Key, l.Name)
	}
	return mcp.NewToolResultText(text), nil
}
