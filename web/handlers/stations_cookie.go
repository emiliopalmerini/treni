package handlers

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/emiliopalmerini/treni/internal/domain"
)

const (
	stationsCookieName = "stations"
	maxStations        = 5
)

// parseStations reads the stations cookie and returns the list of saved stations.
// Format: "code:name|code:name|..." where name is URL-encoded.
// Duplicates are removed (last occurrence wins).
func parseStations(r *http.Request) []domain.StationRef {
	cookie, err := r.Cookie(stationsCookieName)
	if err != nil || cookie.Value == "" {
		return nil
	}

	seen := make(map[string]bool)
	var stations []domain.StationRef

	for _, entry := range strings.Split(cookie.Value, "|") {
		parts := strings.SplitN(entry, ":", 2)
		if len(parts) != 2 || parts[0] == "" {
			continue
		}
		code := parts[0]
		if seen[code] {
			continue
		}
		seen[code] = true
		name, _ := url.QueryUnescape(parts[1])
		stations = append(stations, domain.StationRef{Code: code, Name: name})
	}

	return stations
}

// encodeStations serializes a list of domain.StationRefs into the cookie value format.
func encodeStations(stations []domain.StationRef) string {
	parts := make([]string, len(stations))
	for i, s := range stations {
		parts[i] = s.Code + ":" + url.QueryEscape(s.Name)
	}
	return strings.Join(parts, "|")
}

// addStation appends a station to the list (or moves it to the end if already present).
// Returns the updated list and false if the list is already at max capacity with a new station.
func addStation(stations []domain.StationRef, s domain.StationRef) ([]domain.StationRef, bool) {
	// Remove existing entry with same code
	filtered := make([]domain.StationRef, 0, len(stations))
	for _, existing := range stations {
		if existing.Code != s.Code {
			filtered = append(filtered, existing)
		}
	}

	if len(filtered) >= maxStations {
		return stations, false
	}

	return append(filtered, s), true
}

// removeStation removes a station by code and returns the updated list.
func removeStation(stations []domain.StationRef, code string) []domain.StationRef {
	filtered := make([]domain.StationRef, 0, len(stations))
	for _, s := range stations {
		if s.Code != code {
			filtered = append(filtered, s)
		}
	}
	return filtered
}

// writeStationsCookie writes the stations cookie to the response.
func writeStationsCookie(w http.ResponseWriter, stations []domain.StationRef) {
	value := ""
	maxAge := -1
	if len(stations) > 0 {
		value = encodeStations(stations)
		maxAge = 365 * 24 * 3600
	}
	http.SetCookie(w, &http.Cookie{
		Name:     stationsCookieName,
		Value:    value,
		Path:     "/",
		MaxAge:   maxAge,
		SameSite: http.SameSiteLaxMode,
	})
}
