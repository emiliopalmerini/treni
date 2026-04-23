package bot

import "strings"

// parseFromTo parses a user message of the form "<FROM> > <TO>".
// Both sides are free text, trimmed. Returns ok=false if the message
// doesn't contain a `>` or either side is empty.
func parseFromTo(text string) (from, to string, ok bool) {
	idx := strings.Index(text, ">")
	if idx < 0 {
		return "", "", false
	}
	from = strings.TrimSpace(text[:idx])
	to = strings.TrimSpace(text[idx+1:])
	if from == "" || to == "" {
		return "", "", false
	}
	return from, to, true
}
