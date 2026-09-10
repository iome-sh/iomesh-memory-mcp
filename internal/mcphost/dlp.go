package mcphost

import (
	"regexp"
)

// RedactPlaceholder replaces matched secret-shaped tokens on ingest.
const RedactPlaceholder = "[REDACTED]"

// Host-side DLP is residual-honest prefix/shape heuristics only:
//
// Covered (common paste shapes before palace write):
//   - GitHub PATs: ghp_ / gho_ / ghu_ / ghs_ / ghr_ / github_pat_
//   - sk- tokens (OpenAI-style / sk-ant- / sk-proj- with enough entropy)
//   - AWS access-key ids (AKIA…)
//   - Slack xox[baprs]- tokens
//   - PEM private key blocks
//   - Compact JWT-shaped triples (eyJ…eyJ…)
//
// Not covered (do not invent): hardware-bound keys · default envelope
// encryption · full commercial DLP · OCR / binary · Memory GA.
// healthz honesty is unchanged (no DLP field).
// dual_write OFF · not Memory GA.
var (
	reGitHubFine = regexp.MustCompile(`github_pat_[A-Za-z0-9_]{20,}`)
	reGitHubPAT  = regexp.MustCompile(`gh[pousr]_[A-Za-z0-9]{20,}`)
	reSKToken    = regexp.MustCompile(`sk-[A-Za-z0-9_-]{20,}`)
	reAWSKey     = regexp.MustCompile(`AKIA[0-9A-Z]{16}`)
	reSlackTok   = regexp.MustCompile(`xox[baprs]-[A-Za-z0-9-]{10,}`)
	reJWT        = regexp.MustCompile(`eyJ[A-Za-z0-9_-]{10,}\.eyJ[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}`)
	rePEMBlock   = regexp.MustCompile(`-----BEGIN [A-Z ]*PRIVATE KEY-----[\s\S]*?-----END [A-Z ]*PRIVATE KEY-----`)
)

// RedactSecrets replaces common secret-shaped tokens with RedactPlaceholder.
// Empty input is returned unchanged.
func RedactSecrets(s string) string {
	if s == "" {
		return s
	}
	out := s
	out = rePEMBlock.ReplaceAllString(out, RedactPlaceholder)
	out = reGitHubFine.ReplaceAllString(out, RedactPlaceholder)
	out = reGitHubPAT.ReplaceAllString(out, RedactPlaceholder)
	out = reSKToken.ReplaceAllString(out, RedactPlaceholder)
	out = reAWSKey.ReplaceAllString(out, RedactPlaceholder)
	out = reSlackTok.ReplaceAllString(out, RedactPlaceholder)
	out = reJWT.ReplaceAllString(out, RedactPlaceholder)
	return out
}

// SecretsWereRedacted reports whether RedactSecrets changed s.
func SecretsWereRedacted(s string) bool {
	return RedactSecrets(s) != s
}
