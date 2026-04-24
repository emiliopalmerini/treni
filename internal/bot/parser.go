package bot

import "strings"

// parseFromTo parses a user message of the form "<FROM>: <TO>".
// Both sides are free text, trimmed. Returns ok=false if the message
// doesn't contain a `:` or either side is empty. A stray `:` appearing
// later in the message lands in TO (split on the first `:`).
func parseFromTo(text string) (from, to string, ok bool) {
	before, after, found := strings.Cut(text, ":")
	if !found {
		return "", "", false
	}
	from = strings.TrimSpace(before)
	to = strings.TrimSpace(after)
	if from == "" || to == "" {
		return "", "", false
	}
	return from, to, true
}
