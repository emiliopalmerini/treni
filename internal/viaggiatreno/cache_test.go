package viaggiatreno

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/emiliopalmerini/treni/internal/cache"
)

type mockClient struct {
	autocompletaCalls  atomic.Int32
	cercaStazioneCalls atomic.Int32
	cercaTrenoCalls    atomic.Int32
	partenzeCalls      atomic.Int32
	arriviCalls        atomic.Int32
	andamentoCalls     atomic.Int32
}

func (m *mockClient) AutocompletaStazione(_ context.Context, prefix string) ([]Station, error) {
	m.autocompletaCalls.Add(1)
	return []Station{{ID: "S01700", Name: "Milano Centrale"}}, nil
}

func (m *mockClient) CercaStazione(_ context.Context, prefix string) ([]StationDetail, error) {
	m.cercaStazioneCalls.Add(1)
	return []StationDetail{{ID: "S01700", LongName: "Milano Centrale"}}, nil
}

func (m *mockClient) CercaNumeroTreno(_ context.Context, trainNumber string) ([]TrainMatch, error) {
	m.cercaTrenoCalls.Add(1)
	return []TrainMatch{{Number: trainNumber, Origin: "Milano", OriginID: "S01700"}}, nil
}

func (m *mockClient) Partenze(_ context.Context, stationID string, when time.Time) ([]Departure, error) {
	m.partenzeCalls.Add(1)
	return []Departure{{TrainNumber: 9876, Destination: "Roma"}}, nil
}

func (m *mockClient) Arrivi(_ context.Context, stationID string, when time.Time) ([]Arrival, error) {
	m.arriviCalls.Add(1)
	return []Arrival{{TrainNumber: 9876, Origin: "Milano"}}, nil
}

func (m *mockClient) AndamentoTreno(_ context.Context, originID, trainNumber string, departureTS int64) (*TrainStatus, error) {
	m.andamentoCalls.Add(1)
	return &TrainStatus{TrainNumber: 9876, Origin: "Milano", Destination: "Roma"}, nil
}

func TestCachedClient_AutocompletaStazione(t *testing.T) {
	mock := &mockClient{}
	c := cache.NewMemory()
	ttl := cache.TTLConfig{Static: time.Minute, SemiStatic: time.Minute, Realtime: time.Minute}
	cached := NewCachedClient(mock, c, ttl)
	ctx := context.Background()

	// First call - should hit the underlying client
	result1, err := cached.AutocompletaStazione(ctx, "milano")
	if err != nil {
		t.Fatalf("First call failed: %v", err)
	}
	if len(result1) != 1 || result1[0].Name != "Milano Centrale" {
		t.Errorf("Unexpected result: %v", result1)
	}
	if mock.autocompletaCalls.Load() != 1 {
		t.Errorf("Expected 1 call, got %d", mock.autocompletaCalls.Load())
	}

	// Second call - should use cache
	result2, err := cached.AutocompletaStazione(ctx, "milano")
	if err != nil {
		t.Fatalf("Second call failed: %v", err)
	}
	if len(result2) != 1 || result2[0].Name != "Milano Centrale" {
		t.Errorf("Unexpected result: %v", result2)
	}
	if mock.autocompletaCalls.Load() != 1 {
		t.Errorf("Expected still 1 call (cached), got %d", mock.autocompletaCalls.Load())
	}
}

func TestCachedClient_CercaNumeroTreno(t *testing.T) {
	mock := &mockClient{}
	c := cache.NewMemory()
	ttl := cache.TTLConfig{Static: time.Minute, SemiStatic: time.Minute, Realtime: time.Minute}
	cached := NewCachedClient(mock, c, ttl)
	ctx := context.Background()

	// First call
	result1, err := cached.CercaNumeroTreno(ctx, "9876")
	if err != nil {
		t.Fatalf("First call failed: %v", err)
	}
	if len(result1) != 1 {
		t.Errorf("Unexpected result: %v", result1)
	}
	if mock.cercaTrenoCalls.Load() != 1 {
		t.Errorf("Expected 1 call, got %d", mock.cercaTrenoCalls.Load())
	}

	// Second call - cached
	_, err = cached.CercaNumeroTreno(ctx, "9876")
	if err != nil {
		t.Fatalf("Second call failed: %v", err)
	}
	if mock.cercaTrenoCalls.Load() != 1 {
		t.Errorf("Expected still 1 call (cached), got %d", mock.cercaTrenoCalls.Load())
	}
}

func TestCachedClient_Partenze(t *testing.T) {
	mock := &mockClient{}
	c := cache.NewMemory()
	ttl := cache.TTLConfig{Static: time.Minute, SemiStatic: time.Minute, Realtime: time.Minute}
	cached := NewCachedClient(mock, c, ttl)
	ctx := context.Background()
	now := time.Now()

	// First call
	result1, err := cached.Partenze(ctx, "S01700", now)
	if err != nil {
		t.Fatalf("First call failed: %v", err)
	}
	if len(result1) != 1 || result1[0].TrainNumber != 9876 {
		t.Errorf("Unexpected result: %v", result1)
	}
	if mock.partenzeCalls.Load() != 1 {
		t.Errorf("Expected 1 call, got %d", mock.partenzeCalls.Load())
	}

	// Second call - cached (same minute)
	_, err = cached.Partenze(ctx, "S01700", now)
	if err != nil {
		t.Fatalf("Second call failed: %v", err)
	}
	if mock.partenzeCalls.Load() != 1 {
		t.Errorf("Expected still 1 call (cached), got %d", mock.partenzeCalls.Load())
	}
}

func TestCachedClient_Arrivi(t *testing.T) {
	mock := &mockClient{}
	c := cache.NewMemory()
	ttl := cache.TTLConfig{Static: time.Minute, SemiStatic: time.Minute, Realtime: time.Minute}
	cached := NewCachedClient(mock, c, ttl)
	ctx := context.Background()
	now := time.Now()

	// First call
	result1, err := cached.Arrivi(ctx, "S01700", now)
	if err != nil {
		t.Fatalf("First call failed: %v", err)
	}
	if len(result1) != 1 {
		t.Errorf("Unexpected result: %v", result1)
	}
	if mock.arriviCalls.Load() != 1 {
		t.Errorf("Expected 1 call, got %d", mock.arriviCalls.Load())
	}

	// Second call - cached
	_, err = cached.Arrivi(ctx, "S01700", now)
	if err != nil {
		t.Fatalf("Second call failed: %v", err)
	}
	if mock.arriviCalls.Load() != 1 {
		t.Errorf("Expected still 1 call (cached), got %d", mock.arriviCalls.Load())
	}
}

func TestCachedClient_AndamentoTreno(t *testing.T) {
	mock := &mockClient{}
	c := cache.NewMemory()
	ttl := cache.TTLConfig{Static: time.Minute, SemiStatic: time.Minute, Realtime: time.Minute}
	cached := NewCachedClient(mock, c, ttl)
	ctx := context.Background()

	// First call
	result1, err := cached.AndamentoTreno(ctx, "S01700", "9876", 1234567890)
	if err != nil {
		t.Fatalf("First call failed: %v", err)
	}
	if result1 == nil || result1.TrainNumber != 9876 {
		t.Errorf("Unexpected result: %v", result1)
	}
	if mock.andamentoCalls.Load() != 1 {
		t.Errorf("Expected 1 call, got %d", mock.andamentoCalls.Load())
	}

	// Second call - cached
	result2, err := cached.AndamentoTreno(ctx, "S01700", "9876", 1234567890)
	if err != nil {
		t.Fatalf("Second call failed: %v", err)
	}
	if result2 == nil || result2.TrainNumber != 9876 {
		t.Errorf("Unexpected result: %v", result2)
	}
	if mock.andamentoCalls.Load() != 1 {
		t.Errorf("Expected still 1 call (cached), got %d", mock.andamentoCalls.Load())
	}
}

func TestCachedClient_CacheExpiration(t *testing.T) {
	mock := &mockClient{}
	c := cache.NewMemory()
	ttl := cache.TTLConfig{Static: 50 * time.Millisecond, SemiStatic: 50 * time.Millisecond, Realtime: 50 * time.Millisecond}
	cached := NewCachedClient(mock, c, ttl)
	ctx := context.Background()

	// First call
	_, err := cached.AutocompletaStazione(ctx, "test")
	if err != nil {
		t.Fatalf("First call failed: %v", err)
	}
	if mock.autocompletaCalls.Load() != 1 {
		t.Errorf("Expected 1 call, got %d", mock.autocompletaCalls.Load())
	}

	// Wait for expiration
	time.Sleep(60 * time.Millisecond)

	// Third call - cache expired, should call client again
	_, err = cached.AutocompletaStazione(ctx, "test")
	if err != nil {
		t.Fatalf("Third call failed: %v", err)
	}
	if mock.autocompletaCalls.Load() != 2 {
		t.Errorf("Expected 2 calls after expiration, got %d", mock.autocompletaCalls.Load())
	}
}

func TestCachedClient_DifferentKeys(t *testing.T) {
	mock := &mockClient{}
	c := cache.NewMemory()
	ttl := cache.TTLConfig{Static: time.Minute, SemiStatic: time.Minute, Realtime: time.Minute}
	cached := NewCachedClient(mock, c, ttl)
	ctx := context.Background()

	// Call with different prefixes
	_, _ = cached.AutocompletaStazione(ctx, "milano")
	_, _ = cached.AutocompletaStazione(ctx, "roma")
	_, _ = cached.AutocompletaStazione(ctx, "milano") // cached

	if mock.autocompletaCalls.Load() != 2 {
		t.Errorf("Expected 2 calls (different keys), got %d", mock.autocompletaCalls.Load())
	}
}

func TestNormalizeTime(t *testing.T) {
	t1 := time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC)
	t2 := time.Date(2024, 1, 15, 10, 30, 15, 0, time.UTC)
	t3 := time.Date(2024, 1, 15, 10, 31, 0, 0, time.UTC)

	n1 := normalizeTime(t1)
	n2 := normalizeTime(t2)
	n3 := normalizeTime(t3)

	if n1 != n2 {
		t.Errorf("Same minute should produce same key: %s != %s", n1, n2)
	}
	if n1 == n3 {
		t.Errorf("Different minutes should produce different keys: %s == %s", n1, n3)
	}
}
