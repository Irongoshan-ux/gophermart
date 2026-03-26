package accrual

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type OrderInfo struct {
	Order   string   `json:"order"`
	Status  string   `json:"status"` // REGISTERED, INVALID, PROCESSING, PROCESSED
	Accrual *float64 `json:"accrual,omitempty"`
}

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) GetOrder(ctx context.Context, number string) (*OrderInfo, error) {
	url := c.baseURL + "/api/orders/" + number
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
		var info OrderInfo
		if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
			return nil, fmt.Errorf("decode response: %w", err)
		}
		return &info, nil
	case http.StatusNoContent:
		return nil, nil
	case http.StatusTooManyRequests:
		retryAfter := 60
		if s := resp.Header.Get("Retry-After"); s != "" {
			if n, err := strconv.Atoi(s); err == nil {
				retryAfter = n
			}
		}
		return nil, &RateLimitError{RetryAfter: retryAfter}
	default:
		return nil, fmt.Errorf("accrual API: %s", resp.Status)
	}
}

type RateLimitError struct {
	RetryAfter int
}

func (e *RateLimitError) Error() string {
	return "accrual rate limit exceeded"
}
