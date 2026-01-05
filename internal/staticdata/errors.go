package staticdata

import "errors"

var (
	ErrNotFound          = errors.New("data not found")
	ErrSourceUnavailable = errors.New("data source unavailable")
)
