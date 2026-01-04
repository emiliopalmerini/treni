package cache

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestMemory_SetAndGet(t *testing.T) {
	c := NewMemory()
	ctx := context.Background()

	key := "test-key"
	value := []byte("test-value")
	ttl := 1 * time.Minute

	if err := c.Set(ctx, key, value, ttl); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	got, ok := c.Get(ctx, key)
	if !ok {
		t.Fatal("Get returned false, expected true")
	}

	if string(got) != string(value) {
		t.Errorf("Get returned %q, expected %q", string(got), string(value))
	}
}

func TestMemory_GetNotFound(t *testing.T) {
	c := NewMemory()
	ctx := context.Background()

	got, ok := c.Get(ctx, "nonexistent")
	if ok {
		t.Error("Get returned true for nonexistent key")
	}
	if got != nil {
		t.Errorf("Get returned %v, expected nil", got)
	}
}

func TestMemory_Expiration(t *testing.T) {
	c := NewMemory()
	ctx := context.Background()

	key := "expiring-key"
	value := []byte("expiring-value")
	ttl := 50 * time.Millisecond

	if err := c.Set(ctx, key, value, ttl); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	got, ok := c.Get(ctx, key)
	if !ok {
		t.Fatal("Get returned false before expiration")
	}
	if string(got) != string(value) {
		t.Errorf("Got %q, expected %q", string(got), string(value))
	}

	time.Sleep(60 * time.Millisecond)

	got, ok = c.Get(ctx, key)
	if ok {
		t.Error("Get returned true after expiration")
	}
	if got != nil {
		t.Errorf("Get returned %v after expiration, expected nil", got)
	}
}

func TestMemory_Delete(t *testing.T) {
	c := NewMemory()
	ctx := context.Background()

	key := "delete-key"
	value := []byte("delete-value")

	if err := c.Set(ctx, key, value, time.Minute); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	if _, ok := c.Get(ctx, key); !ok {
		t.Fatal("Key not found before delete")
	}

	if err := c.Delete(ctx, key); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	if _, ok := c.Get(ctx, key); ok {
		t.Error("Key still exists after delete")
	}
}

func TestMemory_DeleteNonexistent(t *testing.T) {
	c := NewMemory()
	ctx := context.Background()

	if err := c.Delete(ctx, "nonexistent"); err != nil {
		t.Errorf("Delete of nonexistent key returned error: %v", err)
	}
}

func TestMemory_Clear(t *testing.T) {
	c := NewMemory()
	ctx := context.Background()

	keys := []string{"key1", "key2", "key3"}
	for _, k := range keys {
		if err := c.Set(ctx, k, []byte(k), time.Minute); err != nil {
			t.Fatalf("Set failed: %v", err)
		}
	}

	for _, k := range keys {
		if _, ok := c.Get(ctx, k); !ok {
			t.Fatalf("Key %s not found before clear", k)
		}
	}

	if err := c.Clear(ctx); err != nil {
		t.Fatalf("Clear failed: %v", err)
	}

	for _, k := range keys {
		if _, ok := c.Get(ctx, k); ok {
			t.Errorf("Key %s still exists after clear", k)
		}
	}
}

func TestMemory_Overwrite(t *testing.T) {
	c := NewMemory()
	ctx := context.Background()

	key := "overwrite-key"
	value1 := []byte("value1")
	value2 := []byte("value2")

	if err := c.Set(ctx, key, value1, time.Minute); err != nil {
		t.Fatalf("First Set failed: %v", err)
	}

	if err := c.Set(ctx, key, value2, time.Minute); err != nil {
		t.Fatalf("Second Set failed: %v", err)
	}

	got, ok := c.Get(ctx, key)
	if !ok {
		t.Fatal("Get returned false")
	}
	if string(got) != string(value2) {
		t.Errorf("Got %q, expected %q", string(got), string(value2))
	}
}

func TestMemory_ConcurrentAccess(t *testing.T) {
	c := NewMemory()
	ctx := context.Background()

	var wg sync.WaitGroup
	numGoroutines := 100
	numOperations := 100

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := "concurrent-key"
				value := []byte("concurrent-value")

				_ = c.Set(ctx, key, value, time.Minute)
				_, _ = c.Get(ctx, key)

				if j%10 == 0 {
					_ = c.Delete(ctx, key)
				}
			}
		}(i)
	}

	wg.Wait()
}

func TestMemory_EmptyValue(t *testing.T) {
	c := NewMemory()
	ctx := context.Background()

	key := "empty-value-key"
	value := []byte{}

	if err := c.Set(ctx, key, value, time.Minute); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	got, ok := c.Get(ctx, key)
	if !ok {
		t.Fatal("Get returned false for empty value")
	}
	if len(got) != 0 {
		t.Errorf("Got %v, expected empty slice", got)
	}
}

func TestMemory_NilValue(t *testing.T) {
	c := NewMemory()
	ctx := context.Background()

	key := "nil-value-key"

	if err := c.Set(ctx, key, nil, time.Minute); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	got, ok := c.Get(ctx, key)
	if !ok {
		t.Fatal("Get returned false for nil value")
	}
	if got != nil {
		t.Errorf("Got %v, expected nil", got)
	}
}

func TestMemory_ZeroTTL(t *testing.T) {
	c := NewMemory()
	ctx := context.Background()

	key := "zero-ttl-key"
	value := []byte("zero-ttl-value")

	if err := c.Set(ctx, key, value, 0); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// With zero TTL, the entry expires immediately
	time.Sleep(1 * time.Millisecond)

	_, ok := c.Get(ctx, key)
	if ok {
		t.Error("Entry with zero TTL should be expired")
	}
}

func TestDefaultTTLConfig(t *testing.T) {
	cfg := DefaultTTLConfig()

	if cfg.Static != 24*time.Hour {
		t.Errorf("Static TTL = %v, expected 24h", cfg.Static)
	}
	if cfg.SemiStatic != 1*time.Hour {
		t.Errorf("SemiStatic TTL = %v, expected 1h", cfg.SemiStatic)
	}
	if cfg.Realtime != 30*time.Second {
		t.Errorf("Realtime TTL = %v, expected 30s", cfg.Realtime)
	}
}
