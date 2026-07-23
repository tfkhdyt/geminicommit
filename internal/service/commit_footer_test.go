package service

import "testing"

func TestBuildIssueFooter_numericID(t *testing.T) {
	got := BuildIssueFooter("123", "Refs")
	want := "Refs #123"
	if got != want {
		t.Fatalf("BuildIssueFooter() = %q, want %q", got, want)
	}
}

func TestBuildIssueFooter_jiraKey(t *testing.T) {
	got := BuildIssueFooter("GEN-123", "Refs")
	want := "Refs GEN-123"
	if got != want {
		t.Fatalf("BuildIssueFooter() = %q, want %q", got, want)
	}
}

func TestBuildIssueFooter_emptyIssue(t *testing.T) {
	if got := BuildIssueFooter("", "Refs"); got != "" {
		t.Fatalf("BuildIssueFooter() = %q, want empty", got)
	}
}

func TestBuildIssueFooter_emptyKeyword(t *testing.T) {
	if got := BuildIssueFooter("123", ""); got != "" {
		t.Fatalf("BuildIssueFooter() = %q, want empty", got)
	}
}

func TestBuildIssueFooter_rejectsLineBreaks(t *testing.T) {
	cases := []struct {
		name    string
		issue   string
		keyword string
	}{
		{name: "newline in issue", issue: "12\n3", keyword: "Refs"},
		{name: "cr in issue", issue: "12\r3", keyword: "Refs"},
		{name: "newline in keyword", issue: "123", keyword: "Re\nfs"},
		{name: "cr in keyword", issue: "123", keyword: "Re\rfs"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := BuildIssueFooter(tc.issue, tc.keyword); got != "" {
				t.Fatalf("BuildIssueFooter() = %q, want empty", got)
			}
		})
	}
}

func TestAppendIssueFooter_appends(t *testing.T) {
	msg := "feat(api): add profile endpoint"
	got := AppendIssueFooter(msg, "128", "Closes")
	want := "feat(api): add profile endpoint\n\nCloses #128"
	if got != want {
		t.Fatalf("AppendIssueFooter() = %q, want %q", got, want)
	}
}

func TestAppendIssueFooter_idempotent(t *testing.T) {
	msg := "feat(api): add profile endpoint\n\nRefs #128"
	got := AppendIssueFooter(msg, "128", "Refs")
	if got != msg {
		t.Fatalf("AppendIssueFooter() = %q, want unchanged %q", got, msg)
	}
}

func TestAppendIssueFooter_idempotentOnlyFinalParagraph(t *testing.T) {
	msg := "feat(api): mention Refs #128 in body\n\nmore detail"
	got := AppendIssueFooter(msg, "128", "Refs")
	want := "feat(api): mention Refs #128 in body\n\nmore detail\n\nRefs #128"
	if got != want {
		t.Fatalf("AppendIssueFooter() = %q, want %q", got, want)
	}
}
