package viaggiatreno

import "time"

// Station represents a train station from the autocomplete endpoint.
type Station struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// StationDetail represents detailed station info from cercaStazione.
type StationDetail struct {
	ID        string  `json:"id"`
	LongName  string  `json:"nomeLungo"`
	ShortName string  `json:"nomeBreve"`
	Label     string  `json:"label"`
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lon"`
	Region    int     `json:"codiceRegione"`
}

// TrainMatch represents a train search result from cercaNumeroTrenoTrenoAutocomplete.
type TrainMatch struct {
	Number      string `json:"number"`
	Origin      string `json:"origin"`
	OriginID    string `json:"originId"`
	DepartureTS int64  `json:"departureTs"`
}

// Departure represents a train departure from partenze endpoint.
type Departure struct {
	TrainNumber      int    `json:"numeroTreno"`
	Category         string `json:"categoria"`
	CategoryDesc     string `json:"categoriaDescrizione"`
	Origin           string `json:"origine"`
	OriginID         string `json:"codOrigine"`
	Destination      string `json:"destinazione"`
	DepartureTime    int64  `json:"orarioPartenza"`
	Delay            int    `json:"ritardo"`
	Platform         string `json:"binarioProgrammatoPartenzaDescrizione"`
	ActualPlatform   string `json:"binarioEffettivoPartenzaDescrizione"`
	CirculationState int    `json:"provvedimento"` // 0=normal, 1=cancelled, 2=partially cancelled
	InStation        bool   `json:"inStazione"`
	Departed         bool   `json:"nonPartito"`
}

// Arrival represents a train arrival from arrivi endpoint.
type Arrival struct {
	TrainNumber      int    `json:"numeroTreno"`
	Category         string `json:"categoria"`
	CategoryDesc     string `json:"categoriaDescrizione"`
	Origin           string `json:"origine"`
	Destination      string `json:"destinazione"`
	ArrivalTime      int64  `json:"orarioArrivo"`
	Delay            int    `json:"ritardo"`
	Platform         string `json:"binarioProgrammatoArrivoDescrizione"`
	ActualPlatform   string `json:"binarioEffettivoArrivoDescrizione"`
	CirculationState int    `json:"provvedimento"`
}

// TrainStatus represents the full train journey from andamentoTreno.
type TrainStatus struct {
	TrainNumber       int    `json:"numeroTreno"`
	Category          string `json:"categoria"`
	Origin            string `json:"origine"`
	OriginID          string `json:"idOrigine"`
	Destination       string `json:"destinazione"`
	DestinationID     string `json:"idDestinazione"`
	DepartureTime     int64  `json:"orarioPartenza"`
	ArrivalTime       int64  `json:"orarioArrivo"`
	Delay             int    `json:"ritardo"`
	CirculationState  int    `json:"provvedimento"`
	LastDetection     string `json:"stazioneUltimoRilevamento"`
	LastDetectionTime int64  `json:"oraUltimoRilevamento"`
	Stops             []Stop `json:"fermate"`
	TrainType         string `json:"tipoTreno"`
	Cancelled         bool   `json:"-"`
}

// Stop represents a stop in the train journey.
type Stop struct {
	StationID          string `json:"id"`
	StationName        string `json:"stazione"`
	StopType           string `json:"tipoFermata"` // P=origin, A=destination, F=intermediate
	ScheduledArrival   int64  `json:"arrivo_teorico"`
	ScheduledDeparture int64  `json:"partenza_teorica"`
	ActualArrival      int64  `json:"arrivoReale"`
	ActualDeparture    int64  `json:"partenzaReale"`
	ArrivalDelay       int    `json:"ritardoArrivo"`
	DepartureDelay     int    `json:"ritardoPartenza"`
	Platform           string `json:"binarioProgrammatoArrivoDescrizione"`
	ActualPlatform     string `json:"binarioEffettivoArrivoDescrizione"`
	ActualStopType     int    `json:"actualFermataType"` // 1=regular, 2=unscheduled, 3=suppressed, 0=unavailable
}

// Helper methods

// DepartureTimeUTC converts the departure timestamp to time.Time.
func (d *Departure) DepartureTimeUTC() time.Time {
	return time.UnixMilli(d.DepartureTime)
}

// ArrivalTimeUTC converts the arrival timestamp to time.Time.
func (a *Arrival) ArrivalTimeUTC() time.Time {
	return time.UnixMilli(a.ArrivalTime)
}

// IsCancelled returns true if the train is cancelled.
func (d *Departure) IsCancelled() bool {
	return d.CirculationState == 1
}

// IsPartiallyCancelled returns true if the train is partially cancelled.
func (d *Departure) IsPartiallyCancelled() bool {
	return d.CirculationState == 2
}

// EffectivePlatform returns the actual platform if available, otherwise the scheduled one.
func (d *Departure) EffectivePlatform() string {
	if d.ActualPlatform != "" {
		return d.ActualPlatform
	}
	return d.Platform
}

// EffectivePlatform returns the actual platform if available, otherwise the scheduled one.
func (a *Arrival) EffectivePlatform() string {
	if a.ActualPlatform != "" {
		return a.ActualPlatform
	}
	return a.Platform
}

// EffectivePlatform returns the actual platform if available, otherwise the scheduled one.
func (s *Stop) EffectivePlatform() string {
	if s.ActualPlatform != "" {
		return s.ActualPlatform
	}
	return s.Platform
}

// RegionStation represents a station from the elencoStazioni endpoint.
type RegionStation struct {
	ID        string  `json:"codiceStazione"`
	Name      string  `json:"-"`
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lon"`
	Region    int     `json:"codReg"`
	Localita  struct {
		LongName  string `json:"nomeLungo"`
		ShortName string `json:"nomeBreve"`
		ID        string `json:"id"`
	} `json:"localita"`
}
