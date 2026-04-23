package bot

import "testing"

func TestParseQuery(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantLine string
		wantStn  string
		wantOK   bool
	}{
		{"simple", "S9 Desio", "S9", "Desio", true},
		{"lowercase line", "s9 desio", "S9", "desio", true},
		{"mixed case line", "rV Milano", "RV", "Milano", true},
		{"multi-word station", "RV Milano Centrale", "RV", "Milano Centrale", true},
		{"extra inner whitespace", "S9  Desio", "S9", "Desio", true},
		{"trailing space trimmed", "S9 Desio   ", "S9", "Desio", true},
		{"leading space trimmed", "   S9 Desio", "S9", "Desio", true},
		{"tab separator", "S9\tDesio", "S9", "Desio", true},

		{"no space", "S9", "", "", false},
		{"only spaces after line", "S9   ", "", "", false},
		{"empty", "", "", "", false},
		{"digits-only first", "2419", "", "", false},
		{"digits-only with station", "2419 Milano", "", "", false},
		{"punctuation line", "S-9 Desio", "", "", false},
		{"line with inner digit-letter mix", "R2V Desio", "", "", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			line, station, ok := parseQuery(tc.input)
			if ok != tc.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tc.wantOK)
			}
			if line != tc.wantLine {
				t.Errorf("line = %q, want %q", line, tc.wantLine)
			}
			if station != tc.wantStn {
				t.Errorf("station = %q, want %q", station, tc.wantStn)
			}
		})
	}
}
