package aggregate

import (
	"html"
	"regexp"
)

var htmlTagPattern = regexp.MustCompile(`<[^>]*>`)

// stripSourceHTML removes SonarQube HTML markup from source line code.
func stripSourceHTML(s string) string {
	plain := htmlTagPattern.ReplaceAllString(s, "")
	return html.UnescapeString(plain)
}
