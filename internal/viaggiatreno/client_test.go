package viaggiatreno_test

import (
	"context"
	"testing"
	"time"

	"github.com/emiliopalmerini/treni/internal/viaggiatreno"
)

func TestAutocompletaStazione(t *testing.T) {
	client := viaggiatreno.NewClient()
	ctx := context.Background()

	stations, err := client.AutocompletaStazione(ctx, "MILANO")
	if err != nil {
		t.Fatalf("AutocompletaStazione failed: %v", err)
	}

	if len(stations) == 0 {
		t.Fatal("expected at least one station")
	}

	found := false
	for _, s := range stations {
		if s.Name == "MILANO CENTRALE" && s.ID == "S01700" {
			found = true
			break
		}
	}

	if !found {
		t.Error("expected to find MILANO CENTRALE with ID S01700")
	}
}

func TestCercaNumeroTreno(t *testing.T) {
	client := viaggiatreno.NewClient()
	ctx := context.Background()

	matches, err := client.CercaNumeroTreno(ctx, "9600")
	if err != nil {
		t.Fatalf("CercaNumeroTreno failed: %v", err)
	}

	if len(matches) == 0 {
		t.Skip("no trains found for number 9600, skipping")
	}

	t.Logf("Found %d matches for train 9600", len(matches))
	for _, m := range matches {
		t.Logf("  Train %s from %s (ID: %s)", m.Number, m.Origin, m.OriginID)
	}
}

func TestPartenze(t *testing.T) {
	client := viaggiatreno.NewClient()
	ctx := context.Background()

	departures, err := client.Partenze(ctx, "S01700", time.Now())
	if err != nil {
		t.Fatalf("Partenze failed: %v", err)
	}

	t.Logf("Found %d departures from Milano Centrale", len(departures))
	for i, d := range departures {
		if i >= 3 {
			break
		}
		t.Logf("  Train %d to %s, delay: %d min", d.TrainNumber, d.Destination, d.Delay)
	}
}
