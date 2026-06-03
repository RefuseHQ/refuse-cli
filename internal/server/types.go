// Package server is the typed HTTP client for the refuse REST API.
package server

// CheckPackageRequest is the body for POST /api/v1/check/package.
type CheckPackageRequest struct {
	Ecosystem string `json:"ecosystem"`
	Name      string `json:"name"`
	Version   string `json:"version"`
}

// BatchCheckRequest is the body for POST /api/v1/check/batch.
type BatchCheckRequest struct {
	Packages []CheckPackageRequest `json:"packages"`
}

// SuggestedFix mirrors the server-side SuggestedFix in @refuse/shared.
type SuggestedFix struct {
	Version        string `json:"version"`
	Type           string `json:"type"`
	BreakingChange bool   `json:"breaking_change"`
	Rationale      string `json:"rationale"`
}

// VulnerabilityRef is one advisory entry from a check response.
type VulnerabilityRef struct {
	RefuseID       string   `json:"refuse_id"`
	CVE            *string  `json:"cve,omitempty"`
	GHSA           *string  `json:"ghsa,omitempty"`
	SeverityScore  float64  `json:"severity_score"`
	SeverityLabel  string   `json:"severity_label"`
	Summary        string   `json:"summary"`
	References     []string `json:"references,omitempty"`
	IsMalicious    bool     `json:"is_malicious,omitempty"`
	KevListed      bool     `json:"kev_listed,omitempty"`
	KevAddedAt     *string  `json:"kev_added_at,omitempty"`
	RansomwareUse  bool     `json:"ransomware_use,omitempty"`
	EpssScore      float64  `json:"epss_score,omitempty"`
	EpssPercentile float64  `json:"epss_percentile,omitempty"`
}

// CheckPackageResponse is the body of POST /api/v1/check/package.
type CheckPackageResponse struct {
	Vulnerable      bool               `json:"vulnerable"`
	Package         string             `json:"package"`
	Version         string             `json:"version"`
	Vulnerabilities []VulnerabilityRef `json:"vulnerabilities"`
	SuggestedFixes  []SuggestedFix     `json:"suggested_fixes"`
	Freshness       string             `json:"freshness"`
	Malicious       bool               `json:"malicious,omitempty"`
	Error           string             `json:"error,omitempty"`
}

// BatchSummary is the aggregate row from a batch check response.
type BatchSummary struct {
	Total      int            `json:"total"`
	Vulnerable int            `json:"vulnerable"`
	BySeverity map[string]int `json:"by_severity"`
	Malicious  int            `json:"malicious,omitempty"`
}

// BatchCheckResponse is the body of POST /api/v1/check/batch.
type BatchCheckResponse struct {
	Results []CheckPackageResponse `json:"results"`
	Summary BatchSummary           `json:"summary"`
}

// CheckLockfileRequest is the body for POST /api/v1/check/lockfile.
type CheckLockfileRequest struct {
	Filename string `json:"filename"`
	Content  string `json:"content"`
}

// CheckDockerfileRequest is the body for POST /api/v1/check/dockerfile.
type CheckDockerfileRequest struct {
	Filename string `json:"filename"`
	Content  string `json:"content"`
}

// CheckWorkflowRequest is the body for POST /api/v1/check/workflow.
type CheckWorkflowRequest struct {
	Filename string `json:"filename"`
	Content  string `json:"content"`
}
