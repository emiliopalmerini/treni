package bot

import (
	"regexp"
	"strings"
)

var lineRegex = regexp.MustCompile(`^[A-Za-z]+[0-9]*$`)

// parseQuery parses a user message of the form "<LINE> <STATION...>".
// LINE is uppercased; STATION is trimmed. Returns ok=false if the message
// does not match this shape.
func parseQuery(text string) (line, station string, ok bool) {
	text = strings.TrimSpace(text)
	if text == "" {
		return "", "", false
	}
	idx := strings.IndexAny(text, " \t")
	if idx < 0 {
		return "", "", false
	}
	first := text[:idx]
	rest := strings.TrimSpace(text[idx+1:])
	if !lineRegex.MatchString(first) {
		return "", "", false
	}
	if rest == "" {
		return "", "", false
	}
	return strings.ToUpper(first), rest, true
}
