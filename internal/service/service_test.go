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
	return time.Date(2026, 4, 25, 8, 0, 0, 0, time.UTC)
}

func dep(minutesFromNow int, dest string) domain.Departure {
	return domain.Departure{
		ScheduledTime: fixedNow().Add(time.Duration(minutesFromNow) * time.Minute),
		Destination:   dest,
	}
}

func TestDeparturesFromTo_substringCaseInsensitive(t *testing.T) {
	fc := &fakeClient{
		departures: map[string][]domain.Departure{
			"DESIO": {
				dep(10, "Milano Centrale"),
				dep(15, "Saronno"),
				dep(20, "Milano Porta Garibaldi"),
			},
		},
	}
	svc := service.NewWithClock(fc, fixedNow)

	got, err := svc.DeparturesFromTo(context.Background(), "DESIO", "milano", 60*time.Minute)
	if err != nil {
		t.Fatalf("DeparturesFromTo: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d, want 2 (both Milano destinations)", len(got))
	}
}

func TestDeparturesFromTo_windowExcludesOutOfRange(t *testing.T) {
	fc := &fakeClient{
		departures: map[string][]domain.Departure{
			"DESIO": {
				dep(-5, "Milano"),
				dep(10, "Milano"),
				dep(55, "Milano"),
				dep(65, "Milano"),
			},
		},
	}
	svc := service.NewWithClock(fc, fixedNow)

	got, err := svc.DeparturesFromTo(context.Background(), "DESIO", "Milano", 60*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d, want 2 (only in-window)", len(got))
	}
}

func TestDeparturesFromTo_sortedAscending(t *testing.T) {
	fc := &fakeClient{
		departures: map[string][]domain.Departure{
			"DESIO": {
				dep(40, "Milano"),
				dep(5, "Milano"),
				dep(20, "Milano"),
			},
		},
	}
	svc := service.NewWithClock(fc, fixedNow)

	got, err := svc.DeparturesFromTo(context.Background(), "DESIO", "Milano", 60*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	for i := 1; i < len(got); i++ {
		if got[i].ScheduledTime.Before(got[i-1].ScheduledTime) {
			t.Errorf("not sorted ascending at index %d", i)
		}
	}
}

func TestDeparturesFromTo_noMatchesReturnsEmpty(t *testing.T) {
	fc := &fakeClient{
		departures: map[string][]domain.Departure{
			"DESIO": {dep(10, "Saronno")},
		},
	}
	svc := service.NewWithClock(fc, fixedNow)

	got, err := svc.DeparturesFromTo(context.Background(), "DESIO", "Milano", 60*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("got %d, want 0", len(got))
	}
}

func TestDeparturesFromTo_propagatesAPIError(t *testing.T) {
	want := errors.New("boom")
	fc := &fakeClient{err: want}
	svc := service.NewWithClock(fc, fixedNow)

	_, err := svc.DeparturesFromTo(context.Background(), "DESIO", "Milano", 60*time.Minute)
	if !errors.Is(err, want) {
		t.Fatalf("err = %v, want wraps %v", err, want)
	}
}

func TestDeparturesFromTo_passesClockToAPI(t *testing.T) {
	fc := &fakeClient{departures: map[string][]domain.Departure{"DESIO": {}}}
	svc := service.NewWithClock(fc, fixedNow)

	_, _ = svc.DeparturesFromTo(context.Background(), "DESIO", "M", 60*time.Minute)

	if !fc.gotAt.Equal(fixedNow()) {
		t.Errorf("client got at=%v, want %v", fc.gotAt, fixedNow())
	}
}
