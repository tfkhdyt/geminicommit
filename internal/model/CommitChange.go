package model

// CommitChange represents the structure for an atomic code change.
// It includes fields for the commit message, the reason for the change,
// and a list of file identifiers affected by the change.
type CommitChange struct {
	// CommitMessage holds the main message for the commit.
	CommitMessage string `json:"commitMessage"`

	// Reason explains the rationale behind the atomic change.
	Reason string `json:"reason"`

	// FileIdentifiers is an array of strings representing the names
	// of the files included in this change.
	FileIdentifiers []string `json:"fileIdentifiers"`

	// Error is set if anything went wrong on the way
	Error error
}
