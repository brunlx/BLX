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
// ID is the question id (empty when the error is not tied to one question).
type ValidationError struct {
	ID       string
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
// "#" and "!" are treated as metacharacters: unquoted, "#" would truncate the
// line as a comment and "!" triggers history expansion in interactive bash.
func q(v string) string {
	if v == "" {
		return ""
	}
	needsQuote := strings.ContainsAny(v, " \t\n\r;&|$()<>`\"'\\*?[]{}#!")
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

// validIface reports whether v is a safe network interface name. Unlike
// validHost it rejects slashes, colons and brackets (path/flag injection).
func validIface(v string) bool {
	if v == "" || len(v) > 32 {
		return false
	}
	for _, r := range v {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
		case r == '-', r == '_', r == '.':
		default:
			return false
		}
	}
	return true
}

// validPort reports whether p is within the TCP/UDP port range.
func validPort(p int) bool { return p >= 1 && p <= 65535 }

// validEncoder reports whether v is a safe msfvenom encoder name
// (e.g. x86/shikata_ga_nai). It allows only module-path characters, never
// whitespace or shell metacharacters.
func validEncoder(v string) bool {
	if v == "" || len(v) > 64 {
		return false
	}
	for _, r := range v {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
		case r == '/', r == '_', r == '-', r == '.', r == '+':
		default:
			return false
		}
	}
	return true
}

// validModulePath reports whether v is a safe Metasploit module path
// (e.g. exploit/windows/smb/ms17_010_eternalblue). It must contain at least
// one "/" and only path-safe characters.
func validModulePath(v string) bool {
	if v == "" || len(v) > 200 || !strings.Contains(v, "/") {
		return false
	}
	for _, r := range v {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
		case r == '/', r == '_', r == '-', r == '.', r == '+':
		default:
			return false
		}
	}
	return true
}

// safeDQ reports whether v is safe to embed inside a double-quoted string.
// The generated mimikatz command line is double-quoted and may run under
// cmd, PowerShell or a POSIX shell, so the block list covers all three:
// double quotes and newlines break out everywhere, "$", backtick and "%"
// enable variable/command substitution in PowerShell/cmd/POSIX shells.
// Backslashes are allowed on purpose: Windows ticket paths (C:\...\x.kirbi)
// require them, and a bare "\" is inert inside double quotes in every shell.
func safeDQ(v string) bool {
	return !strings.ContainsAny(v, "\"$\n\r`%")
}
