package aggregate

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	mcputils "sonarqube-mcp/internal/helpers"

	"github.com/mark3labs/mcp-go/mcp"
)

type CoverageStatus string

const (
	CoverageUnknown   CoverageStatus = "unknown"
	CoverageCovered   CoverageStatus = "covered"
	CoveragePartial   CoverageStatus = "partial"
	CoverageUncovered CoverageStatus = "uncovered"
)

type aggSourceLinesResponse struct {
	Sources []aggSourceLine `json:"sources"`
}

type aggSourceLine struct {
	Line              int     `json:"line"`
	Code              string  `json:"code"`
	LineHits          *int    `json:"lineHits"`
	UtLineHits        *int    `json:"utLineHits"`
	Conditions        *int    `json:"conditions"`
	CoveredConditions *int    `json:"coveredConditions"`
	Duplicated        *bool   `json:"duplicated"`
	IsNew             *bool   `json:"isNew"`
	ScmAuthor         *string `json:"scmAuthor"`
	ScmDate           *string `json:"scmDate"`
	ScmRevision       *string `json:"scmRevision"`
}

func (s aggSourceLine) isExecutable() bool {
	return s.LineHits != nil ||
		s.UtLineHits != nil ||
		s.Conditions != nil ||
		s.CoveredConditions != nil
}

func (s aggSourceLine) isDuplicated() bool {
	return s.Duplicated != nil && *s.Duplicated
}

func (s aggSourceLine) isNewCode() bool {
	return s.IsNew != nil && *s.IsNew
}

func (s aggSourceLine) coverageStatus() CoverageStatus {
	if !s.isExecutable() {
		return CoverageUnknown
	}
	//
	// Branch coverage
	if s.Conditions != nil &&
		s.CoveredConditions != nil {
		total := *s.Conditions
		covered := *s.CoveredConditions
		if total > 0 {
			if covered == 0 {
				return CoverageUncovered
			}
			if covered < total {
				return CoveragePartial
			}
		}
	}
	//
	// Overall line coverage
	if s.LineHits != nil {
		if *s.LineHits > 0 {
			return CoverageCovered
		}
		return CoverageUncovered
	}
	//
	// Unit test coverage
	if s.UtLineHits != nil {
		if *s.UtLineHits > 0 {
			return CoverageCovered
		}
		return CoverageUncovered
	}
	return CoverageUnknown
}

// isUncovered marks executable lines with no unit-test coverage.
// SonarQube exposes lineHits/utLineHits together on runtime lines: both 0 means uncovered.
func (s aggSourceLine) isUncovered() bool {
	return s.coverageStatus() == CoverageUncovered
}

type aggIssuesSearchResponse struct {
	Paging      aggIssuePaging `json:"paging"`
	EffortTotal int            `json:"effortTotal"`
	Issues      []aggIssue     `json:"issues"`
	Components  []aggComponent `json:"components"`
	Rules       []aggRule      `json:"rules"`
}

type aggIssuePaging struct {
	PageIndex int `json:"pageIndex"`
	PageSize  int `json:"pageSize"`
	Total     int `json:"total"`
}

type aggIssue struct {
	Key                        string           `json:"key"`
	Rule                       string           `json:"rule"`
	Severity                   string           `json:"severity"`
	Component                  string           `json:"component"`
	Project                    string           `json:"project"`
	Line                       int              `json:"line"`
	Hash                       string           `json:"hash,omitempty"`
	TextRange                  *aggTextRange    `json:"textRange,omitempty"`
	Status                     string           `json:"status"`
	Message                    string           `json:"message"`
	Effort                     string           `json:"effort,omitempty"`
	Debt                       string           `json:"debt,omitempty"`
	Author                     string           `json:"author,omitempty"`
	Tags                       []string         `json:"tags,omitempty"`
	CreationDate               string           `json:"creationDate,omitempty"`
	UpdateDate                 string           `json:"updateDate,omitempty"`
	Type                       string           `json:"type,omitempty"`
	Scope                      string           `json:"scope,omitempty"`
	CleanCodeAttribute         string           `json:"cleanCodeAttribute,omitempty"`
	CleanCodeAttributeCategory string           `json:"cleanCodeAttributeCategory,omitempty"`
	Impacts                    []aggIssueImpact `json:"impacts,omitempty"`
	IssueStatus                string           `json:"issueStatus,omitempty"`
}

type aggTextRange struct {
	StartLine   int `json:"startLine"`
	EndLine     int `json:"endLine"`
	StartOffset int `json:"startOffset"`
	EndOffset   int `json:"endOffset"`
}

type aggIssueImpact struct {
	SoftwareQuality string `json:"softwareQuality"`
	Severity        string `json:"severity"`
}

type aggComponent struct {
	Key       string `json:"key"`
	Qualifier string `json:"qualifier"`
	Name      string `json:"name"`
	LongName  string `json:"longName"`
	Path      string `json:"path"`
}

type aggRule struct {
	Key      string `json:"key"`
	Name     string `json:"name"`
	Lang     string `json:"lang"`
	LangName string `json:"langName"`
}

// AggSearchIssueSummaryResponse is the aggregated file issue + source summary.
type AggSearchIssueSummaryResponse struct {
	Branch      string         `json:"branch,omitempty"`
	Key         string         `json:"key"`
	PullRequest string         `json:"pullRequest,omitempty"`
	NewCode     CodeBucket     `json:"newCode"`
	OverallCode CodeBucket     `json:"overallCode"`
	Paging      aggIssuePaging `json:"paging,omitempty"`
}

// CodeBucket groups issue snippets and filtered source lines for a scope.
type CodeBucket struct {
	IssueContent      []IssueContentEntry `json:"issueContent"`
	UncoverageContent []SourceLineSnippet `json:"uncoverageContent"`
	DuplicatedContent []SourceLineSnippet `json:"duplicatedContent"`
}

// IssueContentEntry is an issue enriched with source lines from /api/sources/lines.
type IssueContentEntry struct {
	Key                        string              `json:"key"`
	Rule                       string              `json:"rule"`
	RuleName                   string              `json:"ruleName,omitempty"`
	Severity                   string              `json:"severity"`
	Type                       string              `json:"type,omitempty"`
	Status                     string              `json:"status"`
	IssueStatus                string              `json:"issueStatus,omitempty"`
	Message                    string              `json:"message"`
	Effort                     string              `json:"effort,omitempty"`
	Author                     string              `json:"author,omitempty"`
	Tags                       []string            `json:"tags,omitempty"`
	StartLine                  int                 `json:"startLine"`
	EndLine                    int                 `json:"endLine"`
	TextRange                  *aggTextRange       `json:"textRange,omitempty"`
	CleanCodeAttribute         string              `json:"cleanCodeAttribute,omitempty"`
	CleanCodeAttributeCategory string              `json:"cleanCodeAttributeCategory,omitempty"`
	Impacts                    []aggIssueImpact    `json:"impacts,omitempty"`
	CreationDate               string              `json:"creationDate,omitempty"`
	SourceLines                []SourceLineSnippet `json:"sourceLines"`
}

// SourceLineSnippet is a compact source line for agent consumption.
// Scope (new/overall, uncovered/duplicated/issue) is implied by the parent JSON path.
type SourceLineSnippet struct {
	Line int    `json:"line"`
	Code string `json:"code"`
}

func NewAggSearchIssueSummaryMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"agg_search_issue_summary",
		"Aggregate File Issue Summary — Calls /api/sources/lines first, then /api/issues/search. Builds issueContent from issue textRange lines, duplicatedContent from duplicated lines, and uncoverageContent from executable lines where lineHits and utLineHits are both 0. Splits results into newCode (isNew lines) and overallCode (full file).",
		json.RawMessage(`{
			"type": "object",
			"properties": {
				"key": {"type": "string", "description": "File component key (e.g. my_project:src/foo/Bar.java)."},
				"branch": {"type": "string", "description": "Branch name."},
				"pullRequest": {"type": "string", "description": "Pull request ID."},
				"issueStatuses": {
					"type": "array",
					"items": {"type": "string", "enum": ["OPEN", "CONFIRMED", "FALSE_POSITIVE", "ACCEPTED", "FIXED", "IN_SANDBOX"]},
					"description": "Issue statuses to include. Defaults to OPEN and CONFIRMED."
				},
				"p": {"type": "number", "description": "Issue page number. Defaults to 1.", "default": 1},
				"ps": {"type": "number", "description": "Issue page size (max 500). Defaults to 500.", "default": 500}
			},
			"required": ["key"],
			"additionalProperties": false
		}`),
	)
}

func AggSearchIssueSummaryHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	fileKey := mcputils.GetOptionalString(args, "key")
	if fileKey == "" {
		return mcp.NewToolResultError("key is required"), nil
	}

	branch := mcputils.GetOptionalString(args, "branch")
	pr := mcputils.GetOptionalString(args, "pullRequest")
	if branch != "" && pr != "" {
		return mcp.NewToolResultError("branch and pullRequest cannot both be specified"), nil
	}

	// API 1: /api/sources/lines — must succeed before issues search.
	sourceParams := url.Values{}
	sourceParams.Set("key", fileKey)
	if branch != "" {
		sourceParams.Set("branch", branch)
	}
	if pr != "" {
		sourceParams.Set("pullRequest", pr)
	}

	client := mcputils.NewSQClient()
	var sourceResp aggSourceLinesResponse
	if err := client.DoGet(ctx, "/api/sources/lines", sourceParams, &sourceResp); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to retrieve source lines: %v", err)), nil
	}

	// API 2: /api/issues/search
	issueStatuses := mcputils.GetStringArray(args, "issueStatuses")
	if len(issueStatuses) == 0 {
		issueStatuses = []string{"OPEN", "CONFIRMED"}
	}

	page := mcputils.GetIntOrDefault(args, "p", 1)
	pageSize := mcputils.GetIntOrDefault(args, "ps", 500)
	if pageSize > 500 {
		pageSize = 500
	}
	if pageSize < 1 {
		pageSize = 500
	}

	issueParams := url.Values{}
	issueParams.Set("components", fileKey)
	issueParams.Set("issueStatuses", strings.Join(issueStatuses, ","))
	issueParams.Set("additionalFields", "_all")
	issueParams.Set("s", "FILE_LINE")
	issueParams.Set("p", strconv.Itoa(page))
	issueParams.Set("ps", strconv.Itoa(pageSize))
	if branch != "" {
		issueParams.Set("branch", branch)
	}
	if pr != "" {
		issueParams.Set("pullRequest", pr)
	}
	if mcputils.IsCloud() {
		if org := mcputils.GetSonarQubeOrg(); org != "" {
			issueParams.Set("organization", org)
		}
	}

	var issuesResp aggIssuesSearchResponse
	if err := client.DoGet(ctx, "/api/issues/search", issueParams, &issuesResp); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to search issues: %v", err)), nil
	}

	response := buildAggSearchIssueSummary(fileKey, branch, pr, sourceResp, issuesResp)
	respJSON, _ := json.MarshalIndent(response, "", "  ")
	return mcp.NewToolResultText(string(respJSON)), nil
}

func buildAggSearchIssueSummary(fileKey, branch, pr string, sourceResp aggSourceLinesResponse, issuesResp aggIssuesSearchResponse) AggSearchIssueSummaryResponse {
	sourceLines := indexSourceLines(sourceResp.Sources)
	ruleNames := indexRuleNames(issuesResp.Rules)

	return AggSearchIssueSummaryResponse{
		Branch:      branch,
		Key:         fileKey,
		PullRequest: pr,
		NewCode:     buildCodeBucket(sourceLines, issuesResp.Issues, ruleNames, true),
		OverallCode: buildCodeBucket(sourceLines, issuesResp.Issues, ruleNames, false),
		Paging:      issuesResp.Paging,
	}
}

func indexSourceLines(sources []aggSourceLine) map[int]aggSourceLine {
	byLine := make(map[int]aggSourceLine, len(sources))
	for _, src := range sources {
		byLine[src.Line] = src
	}
	return byLine
}

func indexRuleNames(rules []aggRule) map[string]string {
	names := make(map[string]string, len(rules))
	for _, rule := range rules {
		names[rule.Key] = rule.Name
	}
	return names
}

func buildCodeBucket(sourceByLine map[int]aggSourceLine, issues []aggIssue, ruleNames map[string]string, newCodeOnly bool) CodeBucket {
	return CodeBucket{
		IssueContent:      buildIssueContent(sourceByLine, issues, ruleNames, newCodeOnly),
		UncoverageContent: buildUncoverageContent(sourceByLine, newCodeOnly),
		DuplicatedContent: buildDuplicatedContent(sourceByLine, newCodeOnly),
	}
}

func buildIssueContent(sourceByLine map[int]aggSourceLine, issues []aggIssue, ruleNames map[string]string, newCodeOnly bool) []IssueContentEntry {
	result := make([]IssueContentEntry, 0, len(issues))
	for _, issue := range issues {
		startLine, endLine := issueLineRange(issue)
		if startLine == 0 {
			continue
		}

		snippets := extractSourceLineSnippets(sourceByLine, startLine, endLine)
		if len(snippets) == 0 {
			continue
		}
		if newCodeOnly && !lineRangeHasNewCode(sourceByLine, startLine, endLine) {
			continue
		}

		result = append(result, IssueContentEntry{
			Key:                        issue.Key,
			Rule:                       issue.Rule,
			RuleName:                   ruleNames[issue.Rule],
			Severity:                   issue.Severity,
			Type:                       issue.Type,
			Status:                     issue.Status,
			IssueStatus:                issue.IssueStatus,
			Message:                    issue.Message,
			Effort:                     issue.Effort,
			Author:                     issue.Author,
			Tags:                       issue.Tags,
			StartLine:                  startLine,
			EndLine:                    endLine,
			TextRange:                  issue.TextRange,
			CleanCodeAttribute:         issue.CleanCodeAttribute,
			CleanCodeAttributeCategory: issue.CleanCodeAttributeCategory,
			Impacts:                    issue.Impacts,
			CreationDate:               issue.CreationDate,
			SourceLines:                snippets,
		})
	}
	return result
}

func buildUncoverageContent(sourceByLine map[int]aggSourceLine, newCodeOnly bool) []SourceLineSnippet {
	result := make([]SourceLineSnippet, 0)
	for lineNum := 1; lineNum <= maxSourceLine(sourceByLine); lineNum++ {
		src, ok := sourceByLine[lineNum]
		if !ok || !src.isUncovered() {
			continue
		}
		if newCodeOnly && !src.isNewCode() {
			continue
		}
		result = append(result, toSourceLineSnippet(src))
	}
	return result
}

func buildDuplicatedContent(sourceByLine map[int]aggSourceLine, newCodeOnly bool) []SourceLineSnippet {
	result := make([]SourceLineSnippet, 0)
	for lineNum := 1; lineNum <= maxSourceLine(sourceByLine); lineNum++ {
		src, ok := sourceByLine[lineNum]
		if !ok || !src.isDuplicated() {
			continue
		}
		if newCodeOnly && !src.isNewCode() {
			continue
		}
		result = append(result, toSourceLineSnippet(src))
	}
	return result
}

func extractSourceLineSnippets(sourceByLine map[int]aggSourceLine, startLine, endLine int) []SourceLineSnippet {
	if endLine < startLine {
		endLine = startLine
	}
	result := make([]SourceLineSnippet, 0, endLine-startLine+1)
	for lineNum := startLine; lineNum <= endLine; lineNum++ {
		src, ok := sourceByLine[lineNum]
		if !ok {
			continue
		}
		result = append(result, toSourceLineSnippet(src))
	}
	return result
}

func lineRangeHasNewCode(sourceByLine map[int]aggSourceLine, startLine, endLine int) bool {
	for lineNum := startLine; lineNum <= endLine; lineNum++ {
		if src, ok := sourceByLine[lineNum]; ok && src.isNewCode() {
			return true
		}
	}
	return false
}

func toSourceLineSnippet(src aggSourceLine) SourceLineSnippet {
	return SourceLineSnippet{
		Line: src.Line,
		Code: strings.TrimSpace(stripSourceHTML(src.Code)),
	}
}

func maxSourceLine(sourceByLine map[int]aggSourceLine) int {
	maxLine := 0
	for lineNum := range sourceByLine {
		if lineNum > maxLine {
			maxLine = lineNum
		}
	}
	return maxLine
}

func issueLineRange(issue aggIssue) (start, end int) {
	if issue.TextRange != nil {
		start = issue.TextRange.StartLine
		end = issue.TextRange.EndLine
		if end == 0 {
			end = start
		}
		return start, end
	}
	if issue.Line > 0 {
		return issue.Line, issue.Line
	}
	return 0, 0
}
