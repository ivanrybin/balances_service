package currconv

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

const RUB = "RUB"

type ratesResponse struct {
	Success   bool               `json:"success"`
	Timestamp int                `json:"timestamp"`
	Base      string             `json:"base"`
	Date      string             `json:"date"`
	Rates     map[string]float64 `json:"rates"`
}

type Converter struct {
	ctx context.Context

	rates      sync.Map
	token      string
	httpClient *http.Client
}

func New(ctx context.Context, token string) (*Converter, error) {
	if token == "" {
		return nil, fmt.Errorf("empty converter token")
	}

	c := &Converter{
		ctx:   ctx,
		token: token,
		httpClient: &http.Client{
			Timeout: time.Second * 10,
		},
	}

	return c, nil
}

func (c *Converter) baseURL() string {
	return "http://api.exchangeratesapi.io/v1/latest?access_key=" + c.token
}

func (c *Converter) ratesURL(from, to string) string {
	return c.baseURL() + "&format=1" + "&base=" + from + "&symbols=" + to
}

func (c *Converter) getRate(from, to string) (ratesResponse, error) {
	req, err := http.NewRequestWithContext(c.ctx, http.MethodGet, c.ratesURL(from, to), nil)
	if err != nil {
		return ratesResponse{}, fmt.Errorf("cannot create request")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return ratesResponse{}, fmt.Errorf("cannot make request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		errMsg, _ := io.ReadAll(resp.Body)
		return ratesResponse{}, fmt.Errorf("bad response status: %v: %v", resp.StatusCode, string(errMsg))
	}

	ratesResp := ratesResponse{}
	if err = json.NewDecoder(resp.Body).Decode(&ratesResp); err != nil {
		return ratesResponse{}, fmt.Errorf("cannot decode response: %w", err)
	}
	if !ratesResp.Success {
		return ratesResponse{}, fmt.Errorf("bad response")
	}
	return ratesResp, nil
}

func (c *Converter) updateRate(from string) error {
	rates, err := c.getRate(from, RUB)
	if err != nil {
		return fmt.Errorf("cannot update currency rates: %w", err)
	}
	for _, rate := range rates.Rates {
		c.rates.Store(from, rate)
	}
	return nil
}

func (c *Converter) FromRUB(sum int64, to string) (int64, error) {
	switch to {
	case "", RUB:
		return sum, nil
	}

	rate, ok := c.rates.Load(strings.ToUpper(to))
	if !ok {
		if err := c.updateRate(to); err != nil {
			return 0, err
		}
		rate, ok = c.rates.Load(strings.ToUpper(to))
		if !ok {
			return 0, fmt.Errorf("no such currency pair")
		}
	}
	return int64(float64(sum) / rate.(float64)), nil
}
