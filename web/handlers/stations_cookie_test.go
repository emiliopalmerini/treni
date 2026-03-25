package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/emiliopalmerini/treni/internal/domain"
)

func TestParseStations_Empty(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	got := parseStations(r)
	if len(got) != 0 {
		t.Fatalf("expected empty, got %v", got)
	}
}

func TestParseStations_Single(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.AddCookie(&http.Cookie{Name: stationsCookieName, Value: "S01700:MILANO+CENTRALE"})

	got := parseStations(r)
	if len(got) != 1 {
		t.Fatalf("expected 1, got %d", len(got))
	}
	if got[0].Code != "S01700" || got[0].Name != "MILANO CENTRALE" {
		t.Fatalf("unexpected station: %+v", got[0])
	}
}

func TestParseStations_Multiple(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.AddCookie(&http.Cookie{
		Name:  stationsCookieName,
		Value: "S01700:MILANO+CENTRALE|S08409:ROMA+TERMINI|S01200:TORINO+PORTA+NUOVA",
	})

	got := parseStations(r)
	if len(got) != 3 {
		t.Fatalf("expected 3, got %d", len(got))
	}
	if got[2].Code != "S01200" || got[2].Name != "TORINO PORTA NUOVA" {
		t.Fatalf("unexpected third station: %+v", got[2])
	}
}

func TestParseStations_Deduplicates(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.AddCookie(&http.Cookie{
		Name:  stationsCookieName,
		Value: "S01700:MILANO+CENTRALE|S08409:ROMA+TERMINI|S01700:MILANO+CENTRALE",
	})

	got := parseStations(r)
	if len(got) != 2 {
		t.Fatalf("expected 2 after dedup, got %d", len(got))
	}
	// First occurrence wins
	if got[0].Code != "S01700" || got[1].Code != "S08409" {
		t.Fatalf("unexpected order: %+v", got)
	}
}

func TestParseStations_SkipsMalformed(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.AddCookie(&http.Cookie{
		Name:  stationsCookieName,
		Value: "S01700:MILANO+CENTRALE|badentry|:nocode|S08409:ROMA+TERMINI",
	})

	got := parseStations(r)
	if len(got) != 2 {
		t.Fatalf("expected 2, got %d: %+v", len(got), got)
	}
}

func TestEncodeStations(t *testing.T) {
	stations := []domain.StationRef{
		{Code: "S01700", Name: "MILANO CENTRALE"},
		{Code: "S08409", Name: "ROMA TERMINI"},
	}
	got := encodeStations(stations)
	want := "S01700:MILANO+CENTRALE|S08409:ROMA+TERMINI"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestAddStation_Appends(t *testing.T) {
	existing := []domain.StationRef{{Code: "S01700", Name: "MILANO CENTRALE"}}
	got, ok := addStation(existing, domain.StationRef{Code: "S08409", Name: "ROMA TERMINI"})
	if !ok {
		t.Fatal("expected ok")
	}
	if len(got) != 2 {
		t.Fatalf("expected 2, got %d", len(got))
	}
	if got[1].Code != "S08409" {
		t.Fatalf("expected S08409 at end, got %+v", got)
	}
}

func TestAddStation_MovesToEnd(t *testing.T) {
	existing := []domain.StationRef{
		{Code: "S01700", Name: "MILANO CENTRALE"},
		{Code: "S08409", Name: "ROMA TERMINI"},
	}
	got, ok := addStation(existing, domain.StationRef{Code: "S01700", Name: "MILANO CENTRALE"})
	if !ok {
		t.Fatal("expected ok")
	}
	if len(got) != 2 {
		t.Fatalf("expected 2, got %d", len(got))
	}
	if got[0].Code != "S08409" || got[1].Code != "S01700" {
		t.Fatalf("expected S01700 moved to end, got %+v", got)
	}
}

func TestAddStation_RejectsOverMax(t *testing.T) {
	existing := make([]domain.StationRef, maxStations)
	for i := range existing {
		existing[i] = domain.StationRef{Code: "S0000" + string(rune('0'+i)), Name: "Station"}
	}
	_, ok := addStation(existing, domain.StationRef{Code: "NEW", Name: "New Station"})
	if ok {
		t.Fatal("expected rejection when at max")
	}
}

func TestAddStation_ReplaceAtMaxIsOk(t *testing.T) {
	existing := make([]domain.StationRef, maxStations)
	for i := range existing {
		existing[i] = domain.StationRef{Code: "S0000" + string(rune('0'+i)), Name: "Station"}
	}
	// Re-adding an existing one should work (moves to end, no net increase)
	got, ok := addStation(existing, existing[0])
	if !ok {
		t.Fatal("expected ok when replacing existing at max")
	}
	if len(got) != maxStations {
		t.Fatalf("expected %d, got %d", maxStations, len(got))
	}
}

func TestRemoveStation(t *testing.T) {
	existing := []domain.StationRef{
		{Code: "S01700", Name: "MILANO CENTRALE"},
		{Code: "S08409", Name: "ROMA TERMINI"},
	}
	got := removeStation(existing, "S01700")
	if len(got) != 1 {
		t.Fatalf("expected 1, got %d", len(got))
	}
	if got[0].Code != "S08409" {
		t.Fatalf("expected S08409, got %+v", got[0])
	}
}

func TestRemoveStation_NotFound(t *testing.T) {
	existing := []domain.StationRef{{Code: "S01700", Name: "MILANO CENTRALE"}}
	got := removeStation(existing, "NOTEXIST")
	if len(got) != 1 {
		t.Fatalf("expected 1 unchanged, got %d", len(got))
	}
}

func TestRemoveStation_Last(t *testing.T) {
	existing := []domain.StationRef{{Code: "S01700", Name: "MILANO CENTRALE"}}
	got := removeStation(existing, "S01700")
	if len(got) != 0 {
		t.Fatalf("expected empty, got %d", len(got))
	}
}

func TestWriteStationsCookie_SetsValue(t *testing.T) {
	w := httptest.NewRecorder()
	stations := []domain.StationRef{
		{Code: "S01700", Name: "MILANO CENTRALE"},
		{Code: "S08409", Name: "ROMA TERMINI"},
	}
	writeStationsCookie(w, stations)

	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(cookies))
	}
	if cookies[0].Name != stationsCookieName {
		t.Fatalf("unexpected cookie name: %s", cookies[0].Name)
	}
	if cookies[0].MaxAge <= 0 {
		t.Fatal("expected positive MaxAge")
	}
}

func TestWriteStationsCookie_ClearsWhenEmpty(t *testing.T) {
	w := httptest.NewRecorder()
	writeStationsCookie(w, nil)

	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(cookies))
	}
	if cookies[0].MaxAge != -1 {
		t.Fatalf("expected MaxAge -1, got %d", cookies[0].MaxAge)
	}
}
