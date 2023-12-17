package livechess

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	endpoint = "localhost:1982"
	path     = "api/v1.0"
)

var (
	DefaultClient = NewClient()
)

// Client is a LiveChess client.
type Client interface {
	EBoards(ctx context.Context) ([]EBoardResponse, error)
}

func NewClient() Client {
	return &httpClient{
		client: http.DefaultClient,
		base:   fmt.Sprintf("http://%v/%v", endpoint, path),
	}
}

type httpClient struct {
	client *http.Client
	base   string
}

func (c *httpClient) EBoards(ctx context.Context) ([]EBoardResponse, error) {
	url := fmt.Sprintf("%v/eboards", c.base)

	resp, err := c.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}
	// logw.Debugf(ctx, "GET %v -> %v", c.base, string(buf))

	var ret []EBoardResponse
	if err := json.Unmarshal(buf, &ret); err != nil {
		return nil, err
	}
	return ret, nil
}
