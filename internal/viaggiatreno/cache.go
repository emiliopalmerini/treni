package viaggiatreno

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/emiliopalmerini/treni/internal/cache"
)

// CachedClient wraps a Client with caching.
type CachedClient struct {
	client Client
	cache  cache.Cache
	ttl    cache.TTLConfig
}

// NewCachedClient creates a new caching wrapper.
func NewCachedClient(client Client, c cache.Cache, ttl cache.TTLConfig) *CachedClient {
	return &CachedClient{
		client: client,
		cache:  c,
		ttl:    ttl,
	}
}

func (c *CachedClient) AutocompletaStazione(ctx context.Context, prefix string) ([]Station, error) {
	key := fmt.Sprintf("autocompleta:%s", prefix)

	if data, ok := c.cache.Get(ctx, key); ok {
		var stations []Station
		if err := json.Unmarshal(data, &stations); err == nil {
			return stations, nil
		}
	}

	stations, err := c.client.AutocompletaStazione(ctx, prefix)
	if err != nil {
		return nil, err
	}

	if data, err := json.Marshal(stations); err == nil {
		c.cache.Set(ctx, key, data, c.ttl.Static)
	}

	return stations, nil
}

func (c *CachedClient) CercaStazione(ctx context.Context, prefix string) ([]StationDetail, error) {
	key := fmt.Sprintf("cerca_stazione:%s", prefix)

	if data, ok := c.cache.Get(ctx, key); ok {
		var stations []StationDetail
		if err := json.Unmarshal(data, &stations); err == nil {
			return stations, nil
		}
	}

	stations, err := c.client.CercaStazione(ctx, prefix)
	if err != nil {
		return nil, err
	}

	if data, err := json.Marshal(stations); err == nil {
		c.cache.Set(ctx, key, data, c.ttl.Static)
	}

	return stations, nil
}

func (c *CachedClient) CercaNumeroTreno(ctx context.Context, trainNumber string) ([]TrainMatch, error) {
	key := fmt.Sprintf("cerca_treno:%s", trainNumber)

	if data, ok := c.cache.Get(ctx, key); ok {
		var matches []TrainMatch
		if err := json.Unmarshal(data, &matches); err == nil {
			return matches, nil
		}
	}

	matches, err := c.client.CercaNumeroTreno(ctx, trainNumber)
	if err != nil {
		return nil, err
	}

	if data, err := json.Marshal(matches); err == nil {
		c.cache.Set(ctx, key, data, c.ttl.SemiStatic)
	}

	return matches, nil
}

func (c *CachedClient) Partenze(ctx context.Context, stationID string, when time.Time) ([]Departure, error) {
	key := fmt.Sprintf("partenze:%s:%s", stationID, normalizeTime(when))

	if data, ok := c.cache.Get(ctx, key); ok {
		var departures []Departure
		if err := json.Unmarshal(data, &departures); err == nil {
			return departures, nil
		}
	}

	departures, err := c.client.Partenze(ctx, stationID, when)
	if err != nil {
		return nil, err
	}

	if data, err := json.Marshal(departures); err == nil {
		c.cache.Set(ctx, key, data, c.ttl.Realtime)
	}

	return departures, nil
}

func (c *CachedClient) Arrivi(ctx context.Context, stationID string, when time.Time) ([]Arrival, error) {
	key := fmt.Sprintf("arrivi:%s:%s", stationID, normalizeTime(when))

	if data, ok := c.cache.Get(ctx, key); ok {
		var arrivals []Arrival
		if err := json.Unmarshal(data, &arrivals); err == nil {
			return arrivals, nil
		}
	}

	arrivals, err := c.client.Arrivi(ctx, stationID, when)
	if err != nil {
		return nil, err
	}

	if data, err := json.Marshal(arrivals); err == nil {
		c.cache.Set(ctx, key, data, c.ttl.Realtime)
	}

	return arrivals, nil
}

func (c *CachedClient) AndamentoTreno(ctx context.Context, originID string, trainNumber string, departureTS int64) (*TrainStatus, error) {
	key := fmt.Sprintf("andamento:%s:%s:%d", originID, trainNumber, departureTS)

	if data, ok := c.cache.Get(ctx, key); ok {
		var status TrainStatus
		if err := json.Unmarshal(data, &status); err == nil {
			return &status, nil
		}
	}

	status, err := c.client.AndamentoTreno(ctx, originID, trainNumber, departureTS)
	if err != nil {
		return nil, err
	}

	if status != nil {
		if data, err := json.Marshal(status); err == nil {
			c.cache.Set(ctx, key, data, c.ttl.Realtime)
		}
	}

	return status, nil
}

func (c *CachedClient) ElencoStazioni(ctx context.Context, regionCode int) ([]RegionStation, error) {
	key := fmt.Sprintf("elenco_stazioni:%d", regionCode)

	if data, ok := c.cache.Get(ctx, key); ok {
		var stations []RegionStation
		if err := json.Unmarshal(data, &stations); err == nil {
			return stations, nil
		}
	}

	stations, err := c.client.ElencoStazioni(ctx, regionCode)
	if err != nil {
		return nil, err
	}

	if data, err := json.Marshal(stations); err == nil {
		c.cache.Set(ctx, key, data, c.ttl.Static)
	}

	return stations, nil
}

// normalizeTime truncates time to minute resolution for cache key generation.
func normalizeTime(t time.Time) string {
	return t.Truncate(time.Minute).Format("200601021504")
}
