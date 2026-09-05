// Package getmyip is a tiny, dependency-free Go client for the getmyip.pro
// IP geolocation API — privacy-first (no logs, no trackers), ip-api-compatible
// JSON. See https://getmyip.pro/docs (API at https://api.getmyip.pro).
package getmyip

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// DefaultBaseURL is the public API endpoint.
const DefaultBaseURL = "https://api.getmyip.pro"

// Result is a single geolocation lookup response.
type Result struct {
	IP          string  `json:"ip"`
	IPVersion   int     `json:"ip_version"`
	IsPrivate   bool    `json:"is_private"`
	CountryCode string  `json:"country_code"`
	Country     string  `json:"country"`
	Region      string  `json:"region"`
	City        string  `json:"city"`
	Postal      string  `json:"postal"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Timezone    string  `json:"timezone"`
	ASN         int     `json:"asn"`
	Org         string  `json:"org"`
	Hostname    string  `json:"hostname,omitempty"`
}

// Client talks to the getmyip API. The zero value is not usable; call New.
type Client struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

// Option configures a Client.
type Option func(*Client)

// WithAPIKey sets the X-API-Key sent with each request (optional — the free
// anonymous tier needs no key; a key raises rate limits).
func WithAPIKey(key string) Option { return func(c *Client) { c.apiKey = key } }

// WithBaseURL overrides the API base URL (useful for testing).
func WithBaseURL(u string) Option { return func(c *Client) { c.baseURL = strings.TrimRight(u, "/") } }

// WithHTTPClient sets a custom *http.Client (timeouts, transport, proxy…).
func WithHTTPClient(h *http.Client) Option { return func(c *Client) { c.http = h } }

// New creates a Client. With no options it talks to the public API over HTTPS
// with a 10s timeout.
func New(opts ...Option) *Client {
	c := &Client{baseURL: DefaultBaseURL, http: &http.Client{Timeout: 10 * time.Second}}
	for _, o := range opts {
		o(c)
	}
	return c
}

// Me returns geolocation for the caller's own public IP.
func (c *Client) Me(ctx context.Context) (*Result, error) {
	var r Result
	if err := c.get(ctx, "/v1/json", &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// Lookup returns geolocation for the given IPv4 or IPv6 address.
func (c *Client) Lookup(ctx context.Context, ip string) (*Result, error) {
	var r Result
	if err := c.get(ctx, "/v1/"+url.PathEscape(ip), &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// Batch looks up many addresses in a single request.
func (c *Client) Batch(ctx context.Context, ips []string) ([]Result, error) {
	body, err := json.Marshal(map[string][]string{"ips": ips})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/batch", strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	c.auth(req)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, apiError(resp)
	}
	var out struct {
		Results []Result `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out.Results, nil
}

func (c *Client) get(ctx context.Context, path string, v any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return err
	}
	c.auth(req)
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return apiError(resp)
	}
	return json.NewDecoder(resp.Body).Decode(v)
}

func (c *Client) auth(req *http.Request) {
	req.Header.Set("Accept", "application/json")
	if c.apiKey != "" {
		req.Header.Set("X-API-Key", c.apiKey)
	}
}

// apiError turns a non-200 response into an error, using the API's
// {"error": "..."} body when present.
func apiError(resp *http.Response) error {
	var e struct {
		Error string `json:"error"`
	}
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<16))
	_ = json.Unmarshal(body, &e)
	if e.Error != "" {
		return fmt.Errorf("getmyip: %s (status %d)", e.Error, resp.StatusCode)
	}
	return fmt.Errorf("getmyip: unexpected status %d", resp.StatusCode)
}
