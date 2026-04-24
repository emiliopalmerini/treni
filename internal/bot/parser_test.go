package bot

import "testing"

func TestParseFromTo(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantFrom string
		wantTo   string
		wantOK   bool
	}{
		{"simple", "Desio: Milano", "Desio", "Milano", true},
		{"no spaces around colon", "Desio:Milano", "Desio", "Milano", true},
		{"multi-word both sides", "Milano Centrale: Brescia Ovest", "Milano Centrale", "Brescia Ovest", true},
		{"extra whitespace", "  Desio   :   Milano  ", "Desio", "Milano", true},
		{"case preserved", "desio: milano", "desio", "milano", true},
		{"stray arrow stays in TO", "Desio : Milano > something", "Desio", "Milano > something", true},
		{"later colon lands in TO", "Desio : Milano : extra", "Desio", "Milano : extra", true},

		{"no colon", "Desio Milano", "", "", false},
		{"empty", "", "", "", false},
		{"colon only", ":", "", "", false},
		{"empty left", ": Milano", "", "", false},
		{"empty right", "Desio :", "", "", false},
		{"whitespace left", "  : Milano", "", "", false},
		{"whitespace right", "Desio :   ", "", "", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			from, to, ok := parseFromTo(tc.input)
			if ok != tc.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tc.wantOK)
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
