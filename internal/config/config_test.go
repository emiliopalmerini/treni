package config_test

import (
	"strings"
	"testing"

	"github.com/emiliopalmerini/treni/internal/config"
)

func TestLoad_success(t *testing.T) {
	t.Setenv("TELEGRAM_BOT_TOKEN", "abc:123")
	t.Setenv("TELEGRAM_ALLOWED_CHAT_IDS", "42, 7 , 9999")
	t.Setenv("STATE_FILE", "/tmp/state.json")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.BotToken != "abc:123" {
		t.Errorf("BotToken = %q, want abc:123", cfg.BotToken)
	}
	wantChats := []int64{42, 7, 9999}
	if len(cfg.AllowedChats) != len(wantChats) {
		t.Fatalf("AllowedChats len = %d, want %d", len(cfg.AllowedChats), len(wantChats))
	}
	for i, id := range wantChats {
		if cfg.AllowedChats[i] != id {
			t.Errorf("AllowedChats[%d] = %d, want %d", i, cfg.AllowedChats[i], id)
		}
	}
	if cfg.StateFilePath != "/tmp/state.json" {
		t.Errorf("StateFilePath = %q, want /tmp/state.json", cfg.StateFilePath)
	}
}

func TestLoad_defaultStateFile(t *testing.T) {
	t.Setenv("TELEGRAM_BOT_TOKEN", "abc:123")
	t.Setenv("TELEGRAM_ALLOWED_CHAT_IDS", "42")
	t.Setenv("STATE_FILE", "")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.StateFilePath != "./state.json" {
		t.Errorf("StateFilePath = %q, want ./state.json", cfg.StateFilePath)
	}
}

func TestLoad_missingToken(t *testing.T) {
	t.Setenv("TELEGRAM_BOT_TOKEN", "")
	t.Setenv("TELEGRAM_ALLOWED_CHAT_IDS", "42")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for missing token")
	}
	if !strings.Contains(err.Error(), "TELEGRAM_BOT_TOKEN") {
		t.Errorf("error = %v, want mention of TELEGRAM_BOT_TOKEN", err)
	}
}

func TestLoad_missingAllowedChats(t *testing.T) {
	t.Setenv("TELEGRAM_BOT_TOKEN", "abc:123")
	t.Setenv("TELEGRAM_ALLOWED_CHAT_IDS", "")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for missing allowed chats")
	}
	if !strings.Contains(err.Error(), "TELEGRAM_ALLOWED_CHAT_IDS") {
		t.Errorf("error = %v, want mention of TELEGRAM_ALLOWED_CHAT_IDS", err)
	}
}

func TestLoad_malformedChatID(t *testing.T) {
	t.Setenv("TELEGRAM_BOT_TOKEN", "abc:123")
	t.Setenv("TELEGRAM_ALLOWED_CHAT_IDS", "42,notanumber,7")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for malformed chat ID")
	}
	if !strings.Contains(err.Error(), "notanumber") {
		t.Errorf("error = %v, want mention of bad token", err)
	}
}
