package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/emiliopalmerini/treni/internal/domain"
	"github.com/emiliopalmerini/treni/internal/service"
)

type fakeClient struct {
	departures map[string][]domain.Departure
	err        error

	gotAt time.Time
}

func (f *fakeClient) GetTrain(ctx context.Context, n string) (*domain.Train, error) {
	return nil, errors.New("not used")
}
func (f *fakeClient) GetStation(ctx context.Context, c string) (*domain.Station, error) {
	return nil, errors.New("not used")
}
func (f *fakeClient) SearchStation(ctx context.Context, q string) ([]domain.Station, error) {
	return nil, errors.New("not used")
}
func (f *fakeClient) GetDepartures(ctx context.Context, stationCode string, at time.Time) ([]domain.Departure, error) {
	f.gotAt = at
	if f.err != nil {
		return nil, f.err
	}
	return f.departures[stationCode], nil
}

func fixedNow() time.Time {
	return time.Date(2026, 4, 24, 14, 0, 0, 0, time.UTC)
}

func dep(category string, minutesFromNow int, dest string) domain.Departure {
	return domain.Departure{
		TrainCategory: category,
		ScheduledTime: fixedNow().Add(time.Duration(minutesFromNow) * time.Minute),
		Destination:   dest,
	}
}

func TestQueryDepartures_filtersByLineCaseInsensitive(t *testing.T) {
	fc := &fakeClient{
		departures: map[string][]domain.Departure{
			"S123": {
				dep("S9", 10, "Saronno"),
				dep("RV", 12, "Milano Centrale"),
				dep("s9", 15, "Albairate"),
			},
		},
	}
	svc := service.NewWithClock(fc, fixedNow)

	got, err := svc.QueryDepartures(context.Background(), "s9", "S123", 60*time.Minute)
	if err != nil {
		t.Fatalf("QueryDepartures: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d departures, want 2", len(got))
	}
}

func TestQueryDepartures_windowExcludesPastAndFuture(t *testing.T) {
	fc := &fakeClient{
		departures: map[string][]domain.Departure{
			"S123": {
				dep("S9", -5, "Saronno"),   // past: excluded
				dep("S9", 10, "Saronno"),   // inside: included
				dep("S9", 55, "Albairate"), // inside: included
				dep("S9", 65, "Saronno"),   // outside: excluded
			},
		},
	}
	svc := service.NewWithClock(fc, fixedNow)

	got, err := svc.QueryDepartures(context.Background(), "S9", "S123", 60*time.Minute)
	if err != nil {
		t.Fatalf("QueryDepartures: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d, want 2 (only in-window)", len(got))
	}
}

func TestQueryDepartures_sortedAscending(t *testing.T) {
	fc := &fakeClient{
		departures: map[string][]domain.Departure{
			"S123": {
				dep("S9", 40, "Saronno"),
				dep("S9", 5, "Albairate"),
				dep("S9", 20, "Saronno"),
			},
		},
	}
	svc := service.NewWithClock(fc, fixedNow)

	got, err := svc.QueryDepartures(context.Background(), "S9", "S123", 60*time.Minute)
	if err != nil {
		t.Fatalf("QueryDepartures: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("got %d, want 3", len(got))
	}
	for i := 1; i < len(got); i++ {
		if got[i].ScheduledTime.Before(got[i-1].ScheduledTime) {
			t.Errorf("not sorted ascending at index %d", i)
		}
	}
}

func TestQueryDepartures_noMatchesReturnsEmpty(t *testing.T) {
	fc := &fakeClient{
		departures: map[string][]domain.Departure{
			"S123": {dep("RV", 10, "Milano")},
		},
	}
	svc := service.NewWithClock(fc, fixedNow)

	got, err := svc.QueryDepartures(context.Background(), "S9", "S123", 60*time.Minute)
	if err != nil {
		t.Fatalf("QueryDepartures: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("got %d, want 0", len(got))
	}
}

func TestQueryDepartures_propagatesAPIError(t *testing.T) {
	want := errors.New("boom")
	fc := &fakeClient{err: want}
	svc := service.NewWithClock(fc, fixedNow)

	_, err := svc.QueryDepartures(context.Background(), "S9", "S123", 60*time.Minute)
	if !errors.Is(err, want) {
		t.Fatalf("err = %v, want wraps %v", err, want)
	}
}

func TestQueryDepartures_passesClockThroughToAPI(t *testing.T) {
	fc := &fakeClient{departures: map[string][]domain.Departure{"S123": {}}}
	svc := service.NewWithClock(fc, fixedNow)

	_, err := svc.QueryDepartures(context.Background(), "S9", "S123", 60*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if !fc.gotAt.Equal(fixedNow()) {
		t.Errorf("client got at=%v, want %v", fc.gotAt, fixedNow())
	}
}
