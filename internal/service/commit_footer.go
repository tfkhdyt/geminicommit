package service

import (
	"strings"
	"unicode"
)

// BuildIssueFooter returns a trailer line like "Refs #123" or "Refs GEN-123",
// or "" when there's nothing to add.
// ponytail: numeric-only IDs get a leading '#'; JIRA keys/titles are used verbatim.
func BuildIssueFooter(issue, keyword string) string {
	issue = strings.TrimSpace(issue)
	keyword = strings.TrimSpace(keyword)
	if issue == "" || keyword == "" {
		return ""
	}
	if strings.ContainsAny(issue, "\r\n") || strings.ContainsAny(keyword, "\r\n") {
		return ""
	}

	ref := issue
	if isNumericID(issue) {
		ref = "#" + issue
	}

	return keyword + " " + ref
}

// AppendIssueFooter adds the footer as a trailing paragraph, idempotently.
func AppendIssueFooter(message, issue, keyword string) string {
	footer := BuildIssueFooter(issue, keyword)
	if footer == "" {
		return message
	}
	if finalParagraph(message) == footer {
		return message
	}
	return strings.TrimRight(message, "\n") + "\n\n" + footer
}

func finalParagraph(message string) string {
	trimmed := strings.TrimRight(message, "\n")
	if i := strings.LastIndex(trimmed, "\n\n"); i >= 0 {
		return trimmed[i+2:]
	}
	return trimmed
}

func isNumericID(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}
