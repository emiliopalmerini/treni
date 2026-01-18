package viaggiatreno

import (
	"context"
	"testing"
	"time"
)

func TestIntegrationSearchStation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	client := New()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	stations, err := client.SearchStation(ctx, "Roma")
	if err != nil {
		t.Fatalf("SearchStation failed: %v", err)
	}

	if len(stations) == 0 {
		t.Fatal("expected at least one station")
	}

	found := false
	for _, s := range stations {
		if s.Code == "S08409" && s.Name == "ROMA TERMINI" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find Roma Termini (S08409)")
	}
}

func TestIntegrationGetStationRegion(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	client := New()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	region, err := client.GetStationRegion(ctx, "S08409")
	if err != nil {
		t.Fatalf("GetStationRegion failed: %v", err)
	}

	if region == 0 {
		t.Error("expected non-zero region")
	}
}

func TestIntegrationGetDepartures(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	client := New()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	departures, err := client.GetDepartures(ctx, "S08409")
	if err != nil {
		t.Fatalf("GetDepartures failed: %v", err)
	}

	if len(departures) == 0 {
		t.Fatal("expected at least one departure")
	}

	d := departures[0]
	if d.TrainNumber == "" {
		t.Error("expected non-empty train number")
	}
	if d.Destination == "" {
		t.Error("expected non-empty destination")
	}
}

func TestIntegrationGetArrivals(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	client := New()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	arrivals, err := client.GetArrivals(ctx, "S08409")
	if err != nil {
		t.Fatalf("GetArrivals failed: %v", err)
	}

	if len(arrivals) == 0 {
		t.Fatal("expected at least one arrival")
	}

	a := arrivals[0]
	if a.TrainNumber == "" {
		t.Error("expected non-empty train number")
	}
	if a.Origin == "" {
		t.Error("expected non-empty origin")
	}
}

func TestFormatTimestamp(t *testing.T) {
	loc, _ := time.LoadLocation("Europe/Rome")
	tm := time.Date(2025, 1, 18, 14, 30, 0, 0, loc)

	result := formatTimestamp(tm)

	// Should produce something like "Sat Jan 18 2025 14:30:00 GMT+0100"
	if result == "" {
		t.Error("expected non-empty timestamp")
	}
}

func TestParseMillisTimestamp(t *testing.T) {
	tests := []struct {
		name     string
		ms       int64
		wantZero bool
	}{
		{"zero", 0, true},
		{"valid", 1705582200000, false}, // 2024-01-18 14:30:00 UTC
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseMillisTimestamp(tt.ms)
			if tt.wantZero && !result.IsZero() {
				t.Error("expected zero time")
			}
			if !tt.wantZero && result.IsZero() {
				t.Error("expected non-zero time")
			}
		})
	}
}

func TestMapTrainStatus(t *testing.T) {
	tests := []struct {
		provvedimento int
		want          string
	}{
		{0, "on_time"},
		{1, "cancelled"},
		{2, "cancelled"},
		{99, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			result := mapTrainStatus(tt.provvedimento)
			if string(result) != tt.want {
				t.Errorf("got %s, want %s", result, tt.want)
			}
		})
	}
}
