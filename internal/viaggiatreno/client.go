package viaggiatreno

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	defaultBaseURL = "http://www.viaggiatreno.it/infomobilita/resteasy/viaggiatreno"
	defaultTimeout = 10 * time.Second
)

// Client defines the interface for ViaggiaTreno API operations.
type Client interface {
	AutocompletaStazione(ctx context.Context, prefix string) ([]Station, error)
	CercaStazione(ctx context.Context, prefix string) ([]StationDetail, error)
	CercaNumeroTreno(ctx context.Context, trainNumber string) ([]TrainMatch, error)
	Partenze(ctx context.Context, stationID string, when time.Time) ([]Departure, error)
	Arrivi(ctx context.Context, stationID string, when time.Time) ([]Arrival, error)
	AndamentoTreno(ctx context.Context, originID string, trainNumber string, departureTS int64) (*TrainStatus, error)
	ElencoStazioni(ctx context.Context, regionCode int) ([]RegionStation, error)
}

// HTTPClient is an HTTP client for the ViaggiaTreno API.
type HTTPClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewHTTPClient creates a new ViaggiaTreno API client.
func NewHTTPClient() *HTTPClient {
	return &HTTPClient{
		baseURL: defaultBaseURL,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

// WithBaseURL sets a custom base URL for testing.
func (c *HTTPClient) WithBaseURL(url string) *HTTPClient {
	c.baseURL = url
	return c
}

// WithHTTPClient sets a custom HTTP client.
func (c *HTTPClient) WithHTTPClient(client *http.Client) *HTTPClient {
	c.httpClient = client
	return c
}

// AutocompletaStazione searches for stations by prefix.
// Returns a list of matching stations with their IDs.
func (c *HTTPClient) AutocompletaStazione(ctx context.Context, prefix string) ([]Station, error) {
	endpoint := fmt.Sprintf("%s/autocompletaStazione/%s", c.baseURL, url.PathEscape(prefix))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return []Station{}, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Response format: "STATION NAME|STATION_ID\n" per line
	var stations []Station
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "|")
		if len(parts) == 2 {
			stations = append(stations, Station{
				Name: strings.TrimSpace(parts[0]),
				ID:   strings.TrimSpace(parts[1]),
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	return stations, nil
}

// CercaStazione searches for stations and returns detailed info as JSON.
func (c *HTTPClient) CercaStazione(ctx context.Context, prefix string) ([]StationDetail, error) {
	endpoint := fmt.Sprintf("%s/cercaStazione/%s", c.baseURL, url.PathEscape(prefix))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return []StationDetail{}, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var stations []StationDetail
	if err := json.NewDecoder(resp.Body).Decode(&stations); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return stations, nil
}

// CercaNumeroTreno finds trains by number.
// Returns matches when multiple trains share the same number (different origins).
func (c *HTTPClient) CercaNumeroTreno(ctx context.Context, trainNumber string) ([]TrainMatch, error) {
	endpoint := fmt.Sprintf("%s/cercaNumeroTrenoTrenoAutocomplete/%s", c.baseURL, url.PathEscape(trainNumber))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return []TrainMatch{}, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Response format: "NUMBER - ORIGIN|NUMBER-ORIGIN_ID-TIMESTAMP\n" per line
	var matches []TrainMatch
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		// Split by | first
		parts := strings.Split(line, "|")
		if len(parts) != 2 {
			continue
		}

		// First part: "NUMBER - ORIGIN"
		displayParts := strings.SplitN(parts[0], " - ", 2)
		origin := ""
		if len(displayParts) == 2 {
			origin = strings.TrimSpace(displayParts[1])
		}

		// Second part: "NUMBER-ORIGIN_ID-TIMESTAMP"
		dataParts := strings.Split(parts[1], "-")
		if len(dataParts) < 3 {
			continue
		}

		timestamp, _ := strconv.ParseInt(dataParts[len(dataParts)-1], 10, 64)

		matches = append(matches, TrainMatch{
			Number:      dataParts[0],
			Origin:      origin,
			OriginID:    dataParts[1],
			DepartureTS: timestamp,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	return matches, nil
}

// Partenze gets departures from a station.
func (c *HTTPClient) Partenze(ctx context.Context, stationID string, when time.Time) ([]Departure, error) {
	timeStr := formatViaggiatrenoTime(when)
	endpoint := fmt.Sprintf("%s/partenze/%s/%s", c.baseURL, url.PathEscape(stationID), url.PathEscape(timeStr))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return []Departure{}, nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var departures []Departure
	if err := json.NewDecoder(resp.Body).Decode(&departures); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return departures, nil
}

// Arrivi gets arrivals at a station.
func (c *HTTPClient) Arrivi(ctx context.Context, stationID string, when time.Time) ([]Arrival, error) {
	timeStr := formatViaggiatrenoTime(when)
	endpoint := fmt.Sprintf("%s/arrivi/%s/%s", c.baseURL, url.PathEscape(stationID), url.PathEscape(timeStr))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return []Arrival{}, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var arrivals []Arrival
	if err := json.NewDecoder(resp.Body).Decode(&arrivals); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return arrivals, nil
}

// AndamentoTreno gets the full journey status of a train.
func (c *HTTPClient) AndamentoTreno(ctx context.Context, originID string, trainNumber string, departureTS int64) (*TrainStatus, error) {
	endpoint := fmt.Sprintf("%s/andamentoTreno/%s/%s/%d",
		c.baseURL,
		url.PathEscape(originID),
		url.PathEscape(trainNumber),
		departureTS,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var status TrainStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &status, nil
}

// ElencoStazioni gets all stations for a region.
func (c *HTTPClient) ElencoStazioni(ctx context.Context, regionCode int) ([]RegionStation, error) {
	endpoint := fmt.Sprintf("%s/elencoStazioni/%d", c.baseURL, regionCode)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return []RegionStation{}, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var stations []RegionStation
	if err := json.NewDecoder(resp.Body).Decode(&stations); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	for i := range stations {
		stations[i].Name = stations[i].Localita.LongName
	}

	return stations, nil
}

// formatViaggiatrenoTime formats a time for the ViaggiaTreno API.
// Format: "Mon Jan 02 2006 15:04:05 GMT-0700"
func formatViaggiatrenoTime(t time.Time) string {
	return t.Format("Mon Jan 02 2006 15:04:05 GMT-0700")
}
