package tools

import (
	"errors"
	"strconv"
	"strings"
)

// Sentinel errors returned by the catalog.
var (
	ErrUnknownTool = errors.New("ferramenta desconhecida")
	ErrNoGenerator = errors.New("nenhum gerador registrado para esta ferramenta")
)

// ValidationError reports a problem with a single answer.
type ValidationError struct {
	Question string
	Reason   string
}

func (e *ValidationError) Error() string {
	return e.Question + ": " + e.Reason
}

// splitCSV splits a comma-separated value list, trimming spaces.
func splitCSV(v string) []string {
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if s := strings.TrimSpace(p); s != "" {
			out = append(out, s)
		}
	}
	return out
}

// answer returns the raw answer for a question id.
func answer(answers map[string]string, id string) string {
	return strings.TrimSpace(answers[id])
}

// boolAnswer interprets an answer as a boolean ("true"/"on"/"1"/"yes").
func boolAnswer(answers map[string]string, id string) bool {
	v := strings.ToLower(answer(answers, id))
	switch v {
	case "true", "on", "1", "yes", "sim":
		return true
	}
	return false
}

// intAnswer parses an answer as an int, falling back to fallback.
func intAnswer(answers map[string]string, id string, fallback int) int {
	v := answer(answers, id)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

// shquote wraps a value in single quotes so it survives shell interpretation,
// escaping any embedded single quotes. Empty values yield "".
func shquote(v string) string {
	if v == "" {
		return ""
	}
	return "'" + strings.ReplaceAll(v, "'", `'\''`) + "'"
}

// q returns a value safe for shell interpretation: it is wrapped in single
// quotes only when it contains whitespace or shell metacharacters.
// Tilde is deliberately excluded: wrapping "~" in quotes would prevent the
// shell from expanding paths such as ~/wordlists/rockyou.txt.
func q(v string) string {
	if v == "" {
		return ""
	}
	needsQuote := strings.ContainsAny(v, " \t\n;&|$()<>`\"'\\*?[]{}")
	if needsQuote {
		return shquote(v)
	}
	return v
}

// validHost reports whether v is a safe host-like token (IP address or
// hostname). It allows only letters, digits, dots, dashes, underscores,
// slashes, colons and IPv6 brackets, rejecting any shell metacharacter.
func validHost(v string) bool {
	if v == "" || len(v) > 253 {
		return false
	}
	for _, r := range v {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
		case r == '.', r == '-', r == '_', r == '/', r == ':', r == '[', r == ']':
		default:
			return false
		}
	}
	return true
}

// safeDQ reports whether v is safe to embed inside a double-quoted string in
// a Windows cmd context: the only breakout characters there are the double
// quote and newlines.
func safeDQ(v string) bool {
	return !strings.ContainsAny(v, "\"\n\r")
}
