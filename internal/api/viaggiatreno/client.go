package viaggiatreno

import (
	"context"
	"fmt"
	"net/http"

	"github.com/emiliopalmerini/treni/internal/domain"
)

const baseURL = "http://www.viaggiatreno.it/infomobilita/resteasy/viaggiatreno"

type Client struct {
	httpClient *http.Client
}

func New() *Client {
	return &Client{
		httpClient: &http.Client{},
	}
}

func (c *Client) GetTrain(ctx context.Context, trainNumber string) (*domain.Train, error) {
	// TODO: implement
	return nil, fmt.Errorf("not implemented")
}

func (c *Client) GetStation(ctx context.Context, stationCode string) (*domain.Station, error) {
	// TODO: implement
	return nil, fmt.Errorf("not implemented")
}

func (c *Client) SearchStation(ctx context.Context, query string) ([]domain.Station, error) {
	// TODO: implement
	return nil, fmt.Errorf("not implemented")
}
