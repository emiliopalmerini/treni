package viaggiatreno

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/emiliopalmerini/treni/internal/domain"
)

const baseURL = "http://www.viaggiatreno.it/infomobilita/resteasy/viaggiatreno"

type Client struct {
	httpClient *http.Client
	baseURL    string
}

func New() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    baseURL,
	}
}

func (c *Client) SearchStation(ctx context.Context, query string) ([]domain.Station, error) {
	endpoint := fmt.Sprintf("%s/cercaStazione/%s", c.baseURL, url.PathEscape(query))

	body, err := c.doRequest(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("search station: %w", err)
	}

	var results []stationSearchResult
	if err := json.Unmarshal(body, &results); err != nil {
		return nil, fmt.Errorf("parse station search: %w", err)
	}

	stations := make([]domain.Station, len(results))
	for i, r := range results {
		stations[i] = domain.Station{
			Code: r.ID,
			Name: r.NomeLungo,
		}
	}
	return stations, nil
}

func (c *Client) GetStationRegion(ctx context.Context, stationCode string) (int, error) {
	endpoint := fmt.Sprintf("%s/regione/%s", c.baseURL, url.PathEscape(stationCode))

	body, err := c.doRequest(ctx, endpoint)
	if err != nil {
		return 0, fmt.Errorf("get station region: %w", err)
	}

	region, err := strconv.Atoi(strings.TrimSpace(string(body)))
	if err != nil {
		return 0, fmt.Errorf("parse region: %w", err)
	}
	return region, nil
}

func (c *Client) GetDepartures(ctx context.Context, stationCode string) ([]domain.Departure, error) {
	timestamp := formatTimestamp(time.Now())
	endpoint := fmt.Sprintf("%s/partenze/%s/%s", c.baseURL, url.PathEscape(stationCode), url.PathEscape(timestamp))

	body, err := c.doRequest(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("get departures: %w", err)
	}

	var results []departureResult
	if err := json.Unmarshal(body, &results); err != nil {
		return nil, fmt.Errorf("parse departures: %w", err)
	}

	departures := make([]domain.Departure, len(results))
	for i, r := range results {
		departures[i] = domain.Departure{
			TrainNumber:   strconv.Itoa(r.NumeroTreno),
			TrainCategory: r.CategoriaDescrizione,
			Destination:   r.Destinazione,
			ScheduledTime: parseMillisTimestamp(r.OrarioPartenza),
			Delay:         r.Ritardo,
			Platform:      r.BinarioProgrammatoPartenzaDescrizione,
			Status:        mapTrainStatus(r.Provvedimento),
		}
	}
	return departures, nil
}

func (c *Client) GetArrivals(ctx context.Context, stationCode string) ([]domain.Arrival, error) {
	timestamp := formatTimestamp(time.Now())
	endpoint := fmt.Sprintf("%s/arrivi/%s/%s", c.baseURL, url.PathEscape(stationCode), url.PathEscape(timestamp))

	body, err := c.doRequest(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("get arrivals: %w", err)
	}

	var results []arrivalResult
	if err := json.Unmarshal(body, &results); err != nil {
		return nil, fmt.Errorf("parse arrivals: %w", err)
	}

	arrivals := make([]domain.Arrival, len(results))
	for i, r := range results {
		arrivals[i] = domain.Arrival{
			TrainNumber:   strconv.Itoa(r.NumeroTreno),
			TrainCategory: r.CategoriaDescrizione,
			Origin:        r.Origine,
			ScheduledTime: parseMillisTimestamp(r.OrarioArrivo),
			Delay:         r.Ritardo,
			Platform:      r.BinarioProgrammatoArrivoDescrizione,
			Status:        mapTrainStatus(r.Provvedimento),
		}
	}
	return arrivals, nil
}

func (c *Client) GetStation(ctx context.Context, stationCode string) (*domain.Station, error) {
	arrivals, err := c.GetArrivals(ctx, stationCode)
	if err != nil {
		return nil, err
	}

	departures, err := c.GetDepartures(ctx, stationCode)
	if err != nil {
		return nil, err
	}

	return &domain.Station{
		Code:       stationCode,
		Arrivals:   arrivals,
		Departures: departures,
	}, nil
}

func (c *Client) FindTrainOrigin(ctx context.Context, trainNumber string) (string, int64, error) {
	endpoint := fmt.Sprintf("%s/cercaNumeroTrenoTrenoAutocomplete/%s", c.baseURL, url.PathEscape(trainNumber))

	body, err := c.doRequest(ctx, endpoint)
	if err != nil {
		return "", 0, fmt.Errorf("find train origin: %w", err)
	}

	// Response format: "trainNumber - StationName|trainNumber-stationCode-timestamp\n"
	lines := strings.Split(strings.TrimSpace(string(body)), "\n")
	if len(lines) == 0 || lines[0] == "" {
		return "", 0, fmt.Errorf("train not found: %s", trainNumber)
	}

	// Parse first result: "trainNumber - StationName|trainNumber-stationCode-timestamp"
	parts := strings.Split(lines[0], "|")
	if len(parts) < 2 {
		return "", 0, fmt.Errorf("invalid train response format")
	}

	// Parse "trainNumber-stationCode-timestamp"
	dataParts := strings.Split(parts[1], "-")
	if len(dataParts) < 3 {
		return "", 0, fmt.Errorf("invalid train data format")
	}

	stationCode := dataParts[1]
	timestamp, _ := strconv.ParseInt(dataParts[2], 10, 64)

	return stationCode, timestamp, nil
}

func (c *Client) GetTrain(ctx context.Context, trainNumber string) (*domain.Train, error) {
	// First find the origin station and departure timestamp
	originCode, departureTimestamp, err := c.FindTrainOrigin(ctx, trainNumber)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("%s/andamentoTreno/%s/%s/%d",
		c.baseURL,
		url.PathEscape(originCode),
		url.PathEscape(trainNumber),
		departureTimestamp,
	)

	body, err := c.doRequest(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("get train: %w", err)
	}

	var result trainResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse train: %w", err)
	}

	train := &domain.Train{
		Number:        strconv.Itoa(result.NumeroTreno),
		Category:      result.Categoria,
		Origin:        result.Origine,
		Destination:   result.Destinazione,
		DepartureTime: parseMillisTimestamp(result.OrarioPartenza),
		ArrivalTime:   parseMillisTimestamp(result.OrarioArrivo),
		Delay:         result.Ritardo,
		Status:        mapTrainStatus(result.Provvedimento),
		LastUpdate:    parseMillisTimestamp(result.OraUltimoRilevamento),
	}

	train.Stops = make([]domain.Stop, len(result.Fermate))
	for i, f := range result.Fermate {
		train.Stops[i] = domain.Stop{
			StationCode:       f.ID,
			StationName:       f.Stazione,
			ScheduledArrival:  parseMillisTimestamp(f.ArrivoTeorico),
			ActualArrival:     parseMillisTimestamp(f.ArrivoReale),
			ScheduledDepart:   parseMillisTimestamp(f.PartenzaTeorica),
			ActualDepart:      parseMillisTimestamp(f.PartenzaReale),
			ArrivalDelay:      f.RitardoArrivo,
			DepartureDelay:    f.RitardoPartenza,
			Platform:          f.BinarioProgrammatoPartenzaDescrizione,
			PlatformConfirmed: f.BinarioEffettivoPartenzaDescrizione != "",
		}
		if f.BinarioEffettivoPartenzaDescrizione != "" {
			train.Stops[i].Platform = f.BinarioEffettivoPartenzaDescrizione
		}
	}

	return train, nil
}

func (c *Client) doRequest(ctx context.Context, endpoint string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func formatTimestamp(t time.Time) string {
	// Format: "Mon Jan 02 2006 15:04:05 GMT+0100"
	return t.Format("Mon Jan 02 2006 15:04:05 GMT-0700")
}

func parseMillisTimestamp(ms int64) time.Time {
	if ms == 0 {
		return time.Time{}
	}
	return time.UnixMilli(ms)
}

func mapTrainStatus(provvedimento int) domain.TrainStatus {
	switch provvedimento {
	case 0:
		return domain.TrainStatusOnTime
	case 1:
		return domain.TrainStatusCancelled
	case 2:
		return domain.TrainStatusCancelled
	default:
		return domain.TrainStatusUnknown
	}
}
