package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadEnvFile_basic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	const body = `# a comment
FOO=bar
BAZ=qux quux
EMPTY=
# trailing comment
QUOTED="hello world"
SINGLE='single quoted'
`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	for _, k := range []string{"FOO", "BAZ", "EMPTY", "QUOTED", "SINGLE"} {
		t.Setenv(k, "")
		os.Unsetenv(k)
	}

	if err := loadEnvFile(path); err != nil {
		t.Fatalf("loadEnvFile: %v", err)
	}

	wants := map[string]string{
		"FOO":    "bar",
		"BAZ":    "qux quux",
		"EMPTY":  "",
		"QUOTED": "hello world",
		"SINGLE": "single quoted",
	}
	for k, want := range wants {
		if got := os.Getenv(k); got != want {
			t.Errorf("%s = %q, want %q", k, got, want)
		}
	}
}

func TestLoadEnvFile_doesNotOverrideExisting(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	if err := os.WriteFile(path, []byte("PREEXISTING=from_file\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("PREEXISTING", "from_env")
	if err := loadEnvFile(path); err != nil {
		t.Fatalf("loadEnvFile: %v", err)
	}
	if got := os.Getenv("PREEXISTING"); got != "from_env" {
		t.Errorf("PREEXISTING = %q, want from_env (existing must win)", got)
	}
}

func TestLoadEnvFile_missingFileIsOK(t *testing.T) {
	if err := loadEnvFile(filepath.Join(t.TempDir(), "does-not-exist")); err != nil {
		t.Errorf("missing file should not error: %v", err)
	}
}

func TestLoadEnvFile_malformedLineErrors(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	if err := os.WriteFile(path, []byte("not a key value line\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := loadEnvFile(path); err == nil {
		t.Error("expected error for malformed line")
	}
}

func TestLoad_readsDotEnvInCWD(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte(
		"TELEGRAM_BOT_TOKEN=tok\nTELEGRAM_ALLOWED_CHAT_IDS=1,2\n"),
		0o644); err != nil {
		t.Fatal(err)
	}

	oldCWD, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(oldCWD) })
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	os.Unsetenv("TELEGRAM_BOT_TOKEN")
	os.Unsetenv("TELEGRAM_ALLOWED_CHAT_IDS")
	os.Unsetenv("STATE_FILE")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.BotToken != "tok" {
		t.Errorf("BotToken = %q, want tok", cfg.BotToken)
	}
	if len(cfg.AllowedChats) != 2 {
		t.Errorf("AllowedChats = %v, want 2 entries", cfg.AllowedChats)
	}
}
