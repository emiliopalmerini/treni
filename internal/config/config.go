package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	BotToken      string
	AllowedChats  []int64
	StateFilePath string
}

const (
	envBotToken     = "TELEGRAM_BOT_TOKEN"
	envAllowedChats = "TELEGRAM_ALLOWED_CHAT_IDS"
	envStateFile    = "STATE_FILE"

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

	return &Config{
		BotToken:      token,
		AllowedChats:  chats,
		StateFilePath: stateFile,
	}, nil
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
