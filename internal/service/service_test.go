package service_test

import (
	"context"
	"errors"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/emiliopalmerini/treni/internal/domain"
	"github.com/emiliopalmerini/treni/internal/service"
)

type fakeClient struct {
	departures map[string][]domain.Departure
	err        error
	gotAt      time.Time

	trains    map[string]*domain.Train
	trainErrs map[string]error

	mu          sync.Mutex
	trainCalled []string
}

func (f *fakeClient) GetTrain(ctx context.Context, n string) (*domain.Train, error) {
	f.mu.Lock()
	f.trainCalled = append(f.trainCalled, n)
	f.mu.Unlock()
	if err, ok := f.trainErrs[n]; ok {
		return nil, err
	}
	if t, ok := f.trains[n]; ok {
		return t, nil
	}
	return nil, errors.New("train not found")
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

func dep(number string, minutesFromNow int, dest string) domain.Departure {
	return domain.Departure{
		TrainNumber:   number,
		ScheduledTime: fixedNow().Add(time.Duration(minutesFromNow) * time.Minute),
		Destination:   dest,
	}
}

func trainWithStops(number string, stops ...string) *domain.Train {
	s := make([]domain.Stop, len(stops))
	for i, name := range stops {
		s[i] = domain.Stop{StationCode: name, StationName: name}
	}
	return &domain.Train{Number: number, Stops: s}
}


func TestDeparturesVia_matchesIntermediateStop(t *testing.T) {
	// Train terminates at Albairate, but stops at Milano Porta Garibaldi
	// AFTER Desio. User queries `Desio > Milano` — should match.
	fc := &fakeClient{
		departures: map[string][]domain.Departure{
			"DESIO": {dep("24001", 10, "Albairate")},
		},
		trains: map[string]*domain.Train{
			"24001": trainWithStops("24001", "DESIO", "Milano Porta Garibaldi", "Milano Bovisa", "Albairate"),
		},
	}
	svc := service.NewWithClock(fc, fixedNow)

	got, err := svc.DeparturesVia(context.Background(), "DESIO", "Milano", 60*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d, want 1 (train stops at Milano)", len(got))
	}
}

func TestDeparturesVia_excludesReverseDirection(t *testing.T) {
	// Milano appears BEFORE Desio in the stop list: train is heading
	// *away* from Milano. Must NOT match `Desio > Milano`.
	fc := &fakeClient{
		departures: map[string][]domain.Departure{
			"DESIO": {dep("24003", 10, "Saronno")},
		},
		trains: map[string]*domain.Train{
			"24003": trainWithStops("24003",
				"Albairate", "Milano Bovisa", "Milano Porta Garibaldi",
				"DESIO", "Seregno", "Saronno"),
		},
	}
	svc := service.NewWithClock(fc, fixedNow)

	got, err := svc.DeparturesVia(context.Background(), "DESIO", "Milano", 60*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("got %d, want 0 (Milano is before Desio; reverse direction)", len(got))
	}
}

func TestDeparturesVia_fromNotInStopsFallsBackToTerminus(t *testing.T) {
	// GetTrain succeeds but stops don't include the FROM code (odd API).
	// Must fall back to terminus-only match, NOT accept any stop.
	fc := &fakeClient{
		departures: map[string][]domain.Departure{
			"DESIO": {
				dep("A", 10, "Milano Centrale"), // terminus matches
				dep("B", 20, "Saronno"),          // terminus doesn't match
			},
		},
		trains: map[string]*domain.Train{
			"A": trainWithStops("A", "Seregno", "Milano Centrale"), // no DESIO
			"B": trainWithStops("B", "Milano", "Seregno", "Saronno"), // Milano before unrelated stop
		},
	}
	svc := service.NewWithClock(fc, fixedNow)

	got, err := svc.DeparturesVia(context.Background(), "DESIO", "Milano", 60*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].TrainNumber != "A" {
		t.Fatalf("want only A (terminus match); got %v", got)
	}
}

func TestDeparturesVia_matchesTerminus(t *testing.T) {
	fc := &fakeClient{
		departures: map[string][]domain.Departure{
			"DESIO": {dep("99999", 10, "Milano Centrale")},
		},
		trains: map[string]*domain.Train{
			"99999": trainWithStops("99999", "Desio", "Milano Centrale"),
		},
	}
	svc := service.NewWithClock(fc, fixedNow)

	got, err := svc.DeparturesVia(context.Background(), "DESIO", "Milano", 60*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d, want 1", len(got))
	}
}

func TestDeparturesVia_excludesNonMatchingTrain(t *testing.T) {
	fc := &fakeClient{
		departures: map[string][]domain.Departure{
			"DESIO": {
				dep("24001", 10, "Saronno"),  // goes the other way
				dep("24002", 20, "Albairate"), // stops at Milano
			},
		},
		trains: map[string]*domain.Train{
			"24001": trainWithStops("24001", "DESIO", "Seveso", "Saronno"),
			"24002": trainWithStops("24002", "DESIO", "Milano Bovisa", "Albairate"),
		},
	}
	svc := service.NewWithClock(fc, fixedNow)

	got, err := svc.DeparturesVia(context.Background(), "DESIO", "Milano", 60*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d, want 1 (only the Albairate train passes through Milano)", len(got))
	}
	if got[0].TrainNumber != "24002" {
		t.Errorf("wrong train: %s", got[0].TrainNumber)
	}
}

func TestDeparturesVia_getTrainFailsFallsBackToTerminus(t *testing.T) {
	fc := &fakeClient{
		departures: map[string][]domain.Departure{
			"DESIO": {
				dep("24001", 5, "Milano Centrale"), // GetTrain fails; terminus matches
				dep("24002", 10, "Saronno"),        // GetTrain fails; terminus doesn't match
			},
		},
		trainErrs: map[string]error{
			"24001": errors.New("upstream 500"),
			"24002": errors.New("upstream 500"),
		},
	}
	svc := service.NewWithClock(fc, fixedNow)

	got, err := svc.DeparturesVia(context.Background(), "DESIO", "Milano", 60*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d, want 1 (terminus fallback for failed GetTrain)", len(got))
	}
	if got[0].TrainNumber != "24001" {
		t.Errorf("wrong train: %s", got[0].TrainNumber)
	}
}

func TestDeparturesVia_preservesAscendingOrder(t *testing.T) {
	fc := &fakeClient{
		departures: map[string][]domain.Departure{
			"DESIO": {
				dep("A", 30, "Milano"),
				dep("B", 5, "Milano"),
				dep("C", 15, "Milano"),
			},
		},
		trains: map[string]*domain.Train{
			"A": trainWithStops("A", "DESIO", "Milano"),
			"B": trainWithStops("B", "DESIO", "Milano"),
			"C": trainWithStops("C", "DESIO", "Milano"),
		},
	}
	svc := service.NewWithClock(fc, fixedNow)

	got, err := svc.DeparturesVia(context.Background(), "DESIO", "Milano", 60*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if !sort.SliceIsSorted(got, func(i, j int) bool {
		return got[i].ScheduledTime.Before(got[j].ScheduledTime)
	}) {
		t.Errorf("not sorted ascending: %v", got)
	}
}

func TestDeparturesVia_windowFilter(t *testing.T) {
	fc := &fakeClient{
		departures: map[string][]domain.Departure{
			"DESIO": {
				dep("A", -5, "Milano"),
				dep("B", 10, "Milano"),
				dep("C", 65, "Milano"),
			},
		},
		trains: map[string]*domain.Train{
			"A": trainWithStops("A", "DESIO", "Milano"),
			"B": trainWithStops("B", "DESIO", "Milano"),
			"C": trainWithStops("C", "DESIO", "Milano"),
		},
	}
	svc := service.NewWithClock(fc, fixedNow)

	got, err := svc.DeparturesVia(context.Background(), "DESIO", "Milano", 60*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d, want 1 (only in-window)", len(got))
	}
}

func TestDeparturesVia_propagatesDeparturesAPIError(t *testing.T) {
	want := errors.New("boom")
	fc := &fakeClient{err: want}
	svc := service.NewWithClock(fc, fixedNow)

	_, err := svc.DeparturesVia(context.Background(), "DESIO", "Milano", 60*time.Minute)
	if !errors.Is(err, want) {
		t.Fatalf("err = %v, want wraps %v", err, want)
	}
}

func TestDeparturesVia_passesClockToAPI(t *testing.T) {
	fc := &fakeClient{departures: map[string][]domain.Departure{"DESIO": {}}}
	svc := service.NewWithClock(fc, fixedNow)

	_, _ = svc.DeparturesVia(context.Background(), "DESIO", "M", 60*time.Minute)

	if !fc.gotAt.Equal(fixedNow()) {
		t.Errorf("client got at=%v, want %v", fc.gotAt, fixedNow())
	}
}
