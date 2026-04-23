package aggregate

import (
	"testing"
)

func TestStripSourceHTML(t *testing.T) {
	input := `MongoTemplate <span class="sym-74 sym">mongoTemplate</span>) {`
	want := "MongoTemplate mongoTemplate) {"
	if got := stripSourceHTML(input); got != want {
		t.Fatalf("stripSourceHTML() = %q, want %q", got, want)
	}
}

func TestIsUncovered(t *testing.T) {
	covered := aggSourceLine{LineHits: intPtr(1), UtLineHits: intPtr(1)}
	if covered.isUncovered() {
		t.Fatal("covered line should not be uncovered")
	}

	uncovered := aggSourceLine{LineHits: intPtr(0), UtLineHits: intPtr(0)}
	if !uncovered.isUncovered() {
		t.Fatal("both zero should be uncovered")
	}

	comment := aggSourceLine{Line: 1, Code: "// comment"}
	if comment.isUncovered() {
		t.Fatal("comment line without lineHits should not be uncovered")
	}
}

func TestBuildAggSearchIssueSummary(t *testing.T) {
	isNew := true
	notNew := false
	sourceResp := aggSourceLinesResponse{
		Sources: []aggSourceLine{
			{Line: 36, Code: `// old comment`, IsNew: &notNew},
			{Line: 105, Code: `covered()`, LineHits: intPtr(1), UtLineHits: intPtr(1), Duplicated: boolPtr(false), IsNew: &isNew},
			{Line: 115, Code: `uncovered()`, LineHits: intPtr(0), UtLineHits: intPtr(0), Duplicated: boolPtr(false), IsNew: &isNew},
			{Line: 139, Code: `dup()`, Duplicated: boolPtr(true), IsNew: &notNew},
		},
	}
	issuesResp := aggIssuesSearchResponse{
		Paging: aggIssuePaging{PageIndex: 1, PageSize: 500, Total: 1},
		Issues: []aggIssue{
			{
				Key:       "issue-1",
				Rule:      "java:S125",
				Severity:  "MAJOR",
				Line:      36,
				TextRange: &aggTextRange{StartLine: 36, EndLine: 36},
				Status:    "OPEN",
				Message:   "Remove commented code",
				Type:      "CODE_SMELL",
			},
		},
		Rules: []aggRule{{Key: "java:S125", Name: "Sections of code should not be commented out"}},
	}

	resp := buildAggSearchIssueSummary("proj:src/Foo.java", "main", "", sourceResp, issuesResp)

	if resp.Key != "proj:src/Foo.java" || resp.Branch != "main" {
		t.Fatalf("unexpected metadata: %+v", resp)
	}
	if len(resp.OverallCode.IssueContent) != 1 {
		t.Fatalf("overall issueContent = %d, want 1", len(resp.OverallCode.IssueContent))
	}
	if len(resp.OverallCode.IssueContent[0].SourceLines) != 1 {
		t.Fatalf("overall issue source lines = %d, want 1", len(resp.OverallCode.IssueContent[0].SourceLines))
	}
	if len(resp.NewCode.IssueContent) != 0 {
		t.Fatalf("new issueContent = %d, want 0", len(resp.NewCode.IssueContent))
	}
	if len(resp.OverallCode.UncoverageContent) != 1 || resp.OverallCode.UncoverageContent[0].Line != 115 {
		t.Fatalf("overall uncoverage = %+v", resp.OverallCode.UncoverageContent)
	}
	if len(resp.NewCode.UncoverageContent) != 1 || resp.NewCode.UncoverageContent[0].Line != 115 {
		t.Fatalf("new uncoverage = %+v", resp.NewCode.UncoverageContent)
	}
	if len(resp.OverallCode.DuplicatedContent) != 1 || resp.OverallCode.DuplicatedContent[0].Line != 139 {
		t.Fatalf("overall duplicated = %+v", resp.OverallCode.DuplicatedContent)
	}
	if len(resp.NewCode.DuplicatedContent) != 0 {
		t.Fatalf("new duplicated = %d, want 0", len(resp.NewCode.DuplicatedContent))
	}

	snippet := resp.OverallCode.UncoverageContent[0]
	if snippet.Line != 115 || snippet.Code != "uncovered()" {
		t.Fatalf("uncoverage snippet = %+v", snippet)
	}
}

func intPtr(v int) *int    { return &v }
func boolPtr(v bool) *bool { return &v }
