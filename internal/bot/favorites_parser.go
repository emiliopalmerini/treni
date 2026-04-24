package bot

import (
	"errors"
	"strings"
)

// favoriteNameMaxLen caps nicknames at 32 chars. Keeps callback_data
// (max 64B) comfortably in range and keeps list output readable.
const favoriteNameMaxLen = 32

var (
	ErrSaveUsage           = errors.New("usage: /save <name> <FROM>: <TO>")
	ErrUnsaveUsage         = errors.New("usage: /unsave <name>")
	ErrInvalidFavoriteName = errors.New("invalid favorite name")
)

// ValidateFavoriteName enforces the nickname rules from ADR-013, as
// amended by ADR-014: 1–32 chars, no whitespace, no ':' (the query
// separator; would collide with the alias shim), no '>', does not
// start with '/'.
func ValidateFavoriteName(name string) error {
	if name == "" || len(name) > favoriteNameMaxLen {
		return ErrInvalidFavoriteName
	}
	if strings.HasPrefix(name, "/") {
		return ErrInvalidFavoriteName
	}
	if strings.ContainsAny(name, " \t\n\r:>") {
		return ErrInvalidFavoriteName
	}
	return nil
}

// parseSaveCommand parses "/save <name> <FROM>: <TO>" into its parts.
// Name is lowercased. Returns ErrSaveUsage when the grammar is off,
// ErrInvalidFavoriteName when the name fails validation.
func parseSaveCommand(text string) (name, from, to string, err error) {
	rest, ok := stripCommand(text, "/save")
	if !ok || rest == "" {
		return "", "", "", ErrSaveUsage
	}
	nameRaw, routeRaw, split := splitFirstWhitespace(rest)
	if !split || routeRaw == "" {
		return "", "", "", ErrSaveUsage
	}
	if vErr := ValidateFavoriteName(nameRaw); vErr != nil {
		return "", "", "", vErr
	}
	from, to, ok = parseFromTo(routeRaw)
	if !ok {
		return "", "", "", ErrSaveUsage
	}
	return strings.ToLower(nameRaw), from, to, nil
}

// parseUnsaveCommand parses "/unsave <name>".
func parseUnsaveCommand(text string) (string, error) {
	rest, ok := stripCommand(text, "/unsave")
	if !ok || rest == "" {
		return "", ErrUnsaveUsage
	}
	if err := ValidateFavoriteName(rest); err != nil {
		return "", err
	}
	return strings.ToLower(rest), nil
}

func stripCommand(text, cmd string) (string, bool) {
	text = strings.TrimSpace(text)
	if text == cmd {
		return "", true
	}
	if !strings.HasPrefix(text, cmd+" ") && !strings.HasPrefix(text, cmd+"\t") {
		return "", false
	}
	return strings.TrimSpace(text[len(cmd):]), true
}

// splitFirstWhitespace splits at the first whitespace run, returning
// the first token and the trimmed remainder.
func splitFirstWhitespace(s string) (first, rest string, ok bool) {
	s = strings.TrimSpace(s)
	idx := strings.IndexAny(s, " \t\n")
	if idx < 0 {
		return s, "", false
	}
	return s[:idx], strings.TrimSpace(s[idx:]), true
}
