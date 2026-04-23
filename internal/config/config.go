package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	BotToken      string
	AllowedChats  []int64
	StateFilePath string
	// TimeOverride, if non-zero, pins the service clock to this moment.
	// Useful for local dev (e.g. query tomorrow morning from tonight).
	TimeOverride time.Time
}

const (
	envBotToken     = "TELEGRAM_BOT_TOKEN"
	envAllowedChats = "TELEGRAM_ALLOWED_CHAT_IDS"
	envStateFile    = "STATE_FILE"
	envTimeOverride = "TIME_OVERRIDE"

	defaultStateFile = "./state.json"
)

func Load() (*Config, error) {
	if err := loadEnvFile(".env"); err != nil {
		return nil, fmt.Errorf("load .env: %w", err)
	}
	token := strings.TrimSpace(os.Getenv(envBotToken))
	if token == "" {
		return nil, fmt.Errorf("%s is required", envBotToken)
	}

	rawChats := strings.TrimSpace(os.Getenv(envAllowedChats))
	if rawChats == "" {
		return nil, fmt.Errorf("%s is required", envAllowedChats)
	}
	chats, err := parseChatIDs(rawChats)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", envAllowedChats, err)
	}
	if len(chats) == 0 {
		return nil, errors.New("at least one allowed chat ID is required")
	}

	stateFile := strings.TrimSpace(os.Getenv(envStateFile))
	if stateFile == "" {
		stateFile = defaultStateFile
	}

	var override time.Time
	if raw := strings.TrimSpace(os.Getenv(envTimeOverride)); raw != "" {
		parsed, err := parseTimeOverride(raw)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", envTimeOverride, err)
		}
		override = parsed
	}

	return &Config{
		BotToken:      token,
		AllowedChats:  chats,
		StateFilePath: stateFile,
		TimeOverride:  override,
	}, nil
}

// parseTimeOverride accepts RFC3339 (e.g. 2026-04-24T08:00:00+02:00) or a
// looser "2006-01-02 15:04" form in the local timezone.
func parseTimeOverride(s string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	if t, err := time.ParseInLocation("2006-01-02 15:04", s, time.Local); err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("unrecognized time format %q (want RFC3339 or 2006-01-02 15:04)", s)
}

func parseChatIDs(s string) ([]int64, error) {
	parts := strings.Split(s, ",")
	ids := make([]int64, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		id, err := strconv.ParseInt(p, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid chat ID %q: %w", p, err)
		}
		ids = append(ids, id)
	}
	return ids, nil
}
