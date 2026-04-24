package bot

import (
	"errors"
	"testing"
)

func TestValidateFavoriteName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		ok    bool
	}{
		{"simple", "home", true},
		{"mixed case", "Home", true},
		{"digits", "home1", true},
		{"dash", "my-home", true},
		{"underscore", "my_home", true},

		{"empty", "", false},
		{"contains space", "my home", false},
		{"contains tab", "my\thome", false},
		{"contains arrow", "home>work", false},
		{"slash prefix", "/home", false},
		{"too long", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", false}, // 33 chars
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateFavoriteName(tc.input)
			if (err == nil) != tc.ok {
				t.Fatalf("ValidateFavoriteName(%q) err=%v, want ok=%v", tc.input, err, tc.ok)
			}
		})
	}
}

func TestParseSaveCommand(t *testing.T) {
	tests := []struct {
		label    string
		input    string
		wantName string
		wantFrom string
		wantTo   string
		wantErr  error
	}{
		{
			label:    "happy path",
			input:    "/save home Desio > Milano",
			wantName: "home",
			wantFrom: "Desio",
			wantTo:   "Milano",
		},
		{
			label:    "multi-word FROM and TO",
			input:    "/save office Milano Centrale > Brescia Ovest",
			wantName: "office",
			wantFrom: "Milano Centrale",
			wantTo:   "Brescia Ovest",
		},
		{
			label:    "extra whitespace",
			input:    "/save   home    Desio   >   Milano  ",
			wantName: "home",
			wantFrom: "Desio",
			wantTo:   "Milano",
		},
		{
			label:    "name lowercased",
			input:    "/save HOME Desio > Milano",
			wantName: "home",
			wantFrom: "Desio",
			wantTo:   "Milano",
		},

		{label: "no args", input: "/save", wantErr: ErrSaveUsage},
		{label: "only name", input: "/save home", wantErr: ErrSaveUsage},
		{label: "missing arrow", input: "/save home Desio Milano", wantErr: ErrSaveUsage},
		{label: "empty FROM", input: "/save home > Milano", wantErr: ErrSaveUsage},
		{label: "empty TO", input: "/save home Desio >", wantErr: ErrSaveUsage},
		{label: "invalid name with arrow", input: "/save ho>me Desio > Milano", wantErr: ErrInvalidFavoriteName},
		{label: "invalid name slash prefix", input: "/save /home Desio > Milano", wantErr: ErrInvalidFavoriteName},
	}
	for _, tc := range tests {
		t.Run(tc.label, func(t *testing.T) {
			name, from, to, err := parseSaveCommand(tc.input)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("err = %v, want %v", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if name != tc.wantName {
				t.Errorf("name = %q, want %q", name, tc.wantName)
			}
			if from != tc.wantFrom {
				t.Errorf("from = %q, want %q", from, tc.wantFrom)
			}
			if to != tc.wantTo {
				t.Errorf("to = %q, want %q", to, tc.wantTo)
			}
		})
	}
}

func TestParseUnsaveCommand(t *testing.T) {
	tests := []struct {
		label   string
		input   string
		want    string
		wantErr error
	}{
		{"happy", "/unsave home", "home", nil},
		{"lowercased", "/unsave HOME", "home", nil},
		{"whitespace", "/unsave   home  ", "home", nil},
		{"no arg", "/unsave", "", ErrUnsaveUsage},
		{"invalid name", "/unsave ho me", "", ErrInvalidFavoriteName},
	}
	for _, tc := range tests {
		t.Run(tc.label, func(t *testing.T) {
			got, err := parseUnsaveCommand(tc.input)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("err = %v, want %v", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if got != tc.want {
				t.Errorf("name = %q, want %q", got, tc.want)
			}
		})
	}
}
