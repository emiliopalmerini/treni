package viaggiatreno

type stationSearchResult struct {
	NomeLungo string `json:"nomeLungo"`
	NomeBreve string `json:"nomeBreve"`
	Label     string `json:"label"`
	ID        string `json:"id"`
}

type departureResult struct {
	NumeroTreno                           int    `json:"numeroTreno"`
	CategoriaDescrizione                  string `json:"categoriaDescrizione"`
	Destinazione                          string `json:"destinazione"`
	OrarioPartenza                        int64  `json:"orarioPartenza"`
	Ritardo                               int    `json:"ritardo"`
	BinarioProgrammatoPartenzaDescrizione string `json:"binarioProgrammatoPartenzaDescrizione"`
	BinarioEffettivoPartenzaDescrizione   string `json:"binarioEffettivoPartenzaDescrizione"`
	Provvedimento                         int    `json:"provvedimento"`
}

type arrivalResult struct {
	NumeroTreno                         int    `json:"numeroTreno"`
	CategoriaDescrizione                string `json:"categoriaDescrizione"`
	Origine                             string `json:"origine"`
	OrarioArrivo                        int64  `json:"orarioArrivo"`
	Ritardo                             int    `json:"ritardo"`
	BinarioProgrammatoArrivoDescrizione string `json:"binarioProgrammatoArrivoDescrizione"`
	BinarioEffettivoArrivoDescrizione   string `json:"binarioEffettivoArrivoDescrizione"`
	Provvedimento                       int    `json:"provvedimento"`
}

type trainResult struct {
	NumeroTreno          int          `json:"numeroTreno"`
	Categoria            string       `json:"categoria"`
	Origine              string       `json:"origine"`
	Destinazione         string       `json:"destinazione"`
	OrarioPartenza       int64        `json:"orarioPartenza"`
	OrarioArrivo         int64        `json:"orarioArrivo"`
	Ritardo              int          `json:"ritardo"`
	Provvedimento        int          `json:"provvedimento"`
	OraUltimoRilevamento int64        `json:"oraUltimoRilevamento"`
	Fermate              []stopResult `json:"fermate"`
}

type stopResult struct {
	ID                                    string `json:"id"`
	Stazione                              string `json:"stazione"`
	ArrivoTeorico                         int64  `json:"arrivo_teorico"`
	ArrivoReale                           int64  `json:"arrivoReale"`
	PartenzaTeorica                       int64  `json:"partenza_teorica"`
	PartenzaReale                         int64  `json:"partenzaReale"`
	RitardoArrivo                         int    `json:"ritardoArrivo"`
	RitardoPartenza                       int    `json:"ritardoPartenza"`
	BinarioProgrammatoPartenzaDescrizione string `json:"binarioProgrammatoPartenzaDescrizione"`
	BinarioEffettivoPartenzaDescrizione   string `json:"binarioEffettivoPartenzaDescrizione"`
	TipoFermata                           string `json:"tipoFermata"` // P=origin, A=destination, F=intermediate
}
