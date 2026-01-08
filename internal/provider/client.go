// BIND9 API Client

package provider

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client is the BIND9 API client
type Client struct {
	endpoint   string
	apiKey     string
	token      string
	username   string
	password   string
	httpClient *http.Client
}

// NewClient creates a new BIND9 API client
func NewClient(endpoint, apiKey, username, password string, insecure bool, timeout int64) (*Client, error) {
	// Normalize endpoint
	endpoint = strings.TrimSuffix(endpoint, "/")

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure},
	}

	client := &Client{
		endpoint: endpoint,
		apiKey:   apiKey,
		username: username,
		password: password,
		httpClient: &http.Client{
			Timeout:   time.Duration(timeout) * time.Second,
			Transport: transport,
		},
	}

	// If using username/password, get initial token
	if apiKey == "" && username != "" && password != "" {
		if err := client.authenticate(); err != nil {
			return nil, fmt.Errorf("authentication failed: %w", err)
		}
	}

	return client, nil
}

// authenticate gets a JWT token using username/password
func (c *Client) authenticate() error {
	data := url.Values{}
	data.Set("username", c.username)
	data.Set("password", c.password)

	req, err := http.NewRequest("POST", c.endpoint+"/api/v1/auth/token", strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("authentication failed: %s - %s", resp.Status, string(body))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return err
	}

	c.token = tokenResp.AccessToken
	return nil
}

// doRequest performs an HTTP request with authentication
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.endpoint+path, reqBody)
	if err != nil {
		return nil, err
	}

	// Set authentication header
	if c.apiKey != "" {
		req.Header.Set("X-API-Key", c.apiKey)
	} else if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	// Re-authenticate if token expired
	if resp.StatusCode == http.StatusUnauthorized && c.username != "" {
		resp.Body.Close()
		if err := c.authenticate(); err != nil {
			return nil, err
		}
		// Retry request
		return c.doRequest(ctx, method, path, body)
	}

	return resp, nil
}

// parseResponse parses the response body into the given interface
func (c *Client) parseResponse(resp *http.Response, v interface{}) error {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	if v != nil && len(body) > 0 {
		return json.Unmarshal(body, v)
	}

	return nil
}

// ============================================================================
// Zone Operations
// ============================================================================

// Zone represents a DNS zone
type Zone struct {
	Name          string       `json:"name"`
	Type          string       `json:"zone_type"`
	File          string       `json:"file,omitempty"`
	Serial        int64        `json:"serial,omitempty"`
	Loaded        bool         `json:"loaded,omitempty"`
	DNSSECEnabled bool         `json:"dnssec_enabled,omitempty"`
	RecordCount   int64        `json:"record_count,omitempty"`
	Options       *ZoneOptions `json:"options,omitempty"`
}

// ZoneOptions contains zone configuration options
type ZoneOptions struct {
	AllowTransfer []string `json:"allow_transfer,omitempty"`
	AllowUpdate   []string `json:"allow_update,omitempty"`
	AllowQuery    []string `json:"allow_query,omitempty"`
	Notify        bool     `json:"notify,omitempty"`
}

// ZoneCreateRequest is the request body for creating a zone
type ZoneCreateRequest struct {
	Name        string            `json:"name"`
	Type        string            `json:"zone_type"`
	File        string            `json:"file,omitempty"`
	SOAMname    string            `json:"soa_mname,omitempty"`
	SOARname    string            `json:"soa_rname,omitempty"`
	SOARefresh  int               `json:"soa_refresh,omitempty"`
	SOARetry    int               `json:"soa_retry,omitempty"`
	SOAExpire   int               `json:"soa_expire,omitempty"`
	SOAMinimum  int               `json:"soa_minimum,omitempty"`
	DefaultTTL  int               `json:"default_ttl,omitempty"`
	Nameservers []string          `json:"nameservers,omitempty"`
	NSAddresses map[string]string `json:"ns_addresses,omitempty"`
	Options     *ZoneOptions      `json:"options,omitempty"`
}

// GetZone retrieves a zone by name
func (c *Client) GetZone(ctx context.Context, name string) (*Zone, error) {
	resp, err := c.doRequest(ctx, "GET", "/api/v1/zones/"+url.PathEscape(name), nil)
	if err != nil {
		return nil, err
	}

	var zone Zone
	if err := c.parseResponse(resp, &zone); err != nil {
		return nil, err
	}

	return &zone, nil
}

// ListZones retrieves all zones, optionally filtered by parameters
func (c *Client) ListZones(ctx context.Context, params map[string]string) ([]Zone, error) {
	path := "/api/v1/zones"

	if len(params) > 0 {
		query := url.Values{}
		for k, v := range params {
			query.Set(k, v)
		}
		path += "?" + query.Encode()
	}

	resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Zones []Zone `json:"zones"`
	}
	if err := c.parseResponse(resp, &result); err != nil {
		return nil, err
	}

	return result.Zones, nil
}

// CreateZone creates a new zone
func (c *Client) CreateZone(ctx context.Context, req *ZoneCreateRequest) (*Zone, error) {
	resp, err := c.doRequest(ctx, "POST", "/api/v1/zones", req)
	if err != nil {
		return nil, err
	}

	var zone Zone
	if err := c.parseResponse(resp, &zone); err != nil {
		return nil, err
	}

	return &zone, nil
}

// DeleteZone deletes a zone
func (c *Client) DeleteZone(ctx context.Context, name string, deleteFile bool) error {
	path := "/api/v1/zones/" + url.PathEscape(name)
	if deleteFile {
		path += "?delete_file=true"
	}

	resp, err := c.doRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}

	return c.parseResponse(resp, nil)
}

// ReloadZone reloads a zone
func (c *Client) ReloadZone(ctx context.Context, name string) error {
	resp, err := c.doRequest(ctx, "POST", "/api/v1/zones/"+url.PathEscape(name)+"/reload", nil)
	if err != nil {
		return err
	}
	return c.parseResponse(resp, nil)
}

// ============================================================================
// Record Operations
// ============================================================================

// Record represents a DNS record
type Record struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	TTL   int64  `json:"ttl"`
	Class string `json:"class,omitempty"`
	RData string `json:"rdata"`
	Zone  string `json:"zone,omitempty"`
}

// RecordCreateRequest is the request for creating a record
type RecordCreateRequest struct {
	RecordType  string                 `json:"record_type"`
	Name        string                 `json:"name"`
	TTL         int                    `json:"ttl"`
	RecordClass string                 `json:"record_class,omitempty"`
	Data        map[string]interface{} `json:"data"`
}

// GetRecords retrieves records for a zone
func (c *Client) GetRecords(ctx context.Context, zone string, recordType, name string) ([]Record, error) {
	path := "/api/v1/zones/" + url.PathEscape(zone) + "/records"

	params := url.Values{}
	if recordType != "" {
		params.Set("record_type", recordType)
	}
	if name != "" {
		params.Set("name", name)
	}
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var records []Record
	if err := c.parseResponse(resp, &records); err != nil {
		return nil, err
	}

	return records, nil
}

// GetRecord retrieves a specific record
func (c *Client) GetRecord(ctx context.Context, zone, name, recordType string) (*Record, error) {
	records, err := c.GetRecords(ctx, zone, recordType, name)
	if err != nil {
		return nil, err
	}

	for _, r := range records {
		if r.Name == name && r.Type == recordType {
			return &r, nil
		}
	}

	return nil, fmt.Errorf("record not found: %s %s in zone %s", name, recordType, zone)
}

// CreateRecord creates a new record
func (c *Client) CreateRecord(ctx context.Context, zone string, req *RecordCreateRequest) (*Record, error) {
	path := "/api/v1/zones/" + url.PathEscape(zone) + "/records"

	resp, err := c.doRequest(ctx, "POST", path, req)
	if err != nil {
		return nil, err
	}

	var record Record
	if err := c.parseResponse(resp, &record); err != nil {
		return nil, err
	}

	return &record, nil
}

// DeleteRecord deletes a record
func (c *Client) DeleteRecord(ctx context.Context, zone, name, recordType, rdata string) error {
	path := "/api/v1/zones/" + url.PathEscape(zone) + "/records/" +
		url.PathEscape(name) + "/" + url.PathEscape(recordType)

	if rdata != "" {
		path += "?rdata=" + url.QueryEscape(rdata)
	}

	resp, err := c.doRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}

	return c.parseResponse(resp, nil)
}

// ============================================================================
// DNSSEC Operations
// ============================================================================

// DNSSECKey represents a DNSSEC key
type DNSSECKey struct {
	Zone      string   `json:"zone"`
	KeyTag    int      `json:"key_tag"`
	Algorithm int      `json:"algorithm"`
	KeyType   string   `json:"key_type"`
	Bits      int      `json:"bits"`
	State     string   `json:"state"`
	Flags     int      `json:"flags"`
	Protocol  int      `json:"protocol"`
	PublicKey string   `json:"public_key,omitempty"`
	DSRecords []string `json:"ds_records,omitempty"`
}

// DNSSECKeyCreateRequest is the request for creating a DNSSEC key
type DNSSECKeyCreateRequest struct {
	KeyType   string `json:"key_type"`
	Algorithm int    `json:"algorithm"`
	Bits      int    `json:"bits,omitempty"`
	TTL       int    `json:"ttl,omitempty"`
}

// ListDNSSECKeys lists DNSSEC keys for a zone
func (c *Client) ListDNSSECKeys(ctx context.Context, zone string) ([]DNSSECKey, error) {
	resp, err := c.doRequest(ctx, "GET", "/api/v1/dnssec/zones/"+url.PathEscape(zone)+"/keys", nil)
	if err != nil {
		return nil, err
	}

	var keys []DNSSECKey
	if err := c.parseResponse(resp, &keys); err != nil {
		return nil, err
	}

	return keys, nil
}

// CreateDNSSECKey creates a new DNSSEC key
func (c *Client) CreateDNSSECKey(ctx context.Context, zone string, req *DNSSECKeyCreateRequest) (*DNSSECKey, error) {
	resp, err := c.doRequest(ctx, "POST", "/api/v1/dnssec/zones/"+url.PathEscape(zone)+"/keys", req)
	if err != nil {
		return nil, err
	}

	var key DNSSECKey
	if err := c.parseResponse(resp, &key); err != nil {
		return nil, err
	}

	return &key, nil
}

// DeleteDNSSECKey deletes a DNSSEC key
func (c *Client) DeleteDNSSECKey(ctx context.Context, zone string, keyTag int) error {
	resp, err := c.doRequest(ctx, "DELETE", fmt.Sprintf("/api/v1/dnssec/zones/%s/keys/%d",
		url.PathEscape(zone), keyTag), nil)
	if err != nil {
		return err
	}
	return c.parseResponse(resp, nil)
}

// SignZone signs a zone
func (c *Client) SignZone(ctx context.Context, zone string) error {
	resp, err := c.doRequest(ctx, "POST", "/api/v1/dnssec/zones/"+url.PathEscape(zone)+"/sign", nil)
	if err != nil {
		return err
	}
	return c.parseResponse(resp, nil)
}

// ListRecords retrieves records for a zone with optional filters
func (c *Client) ListRecords(ctx context.Context, zone string, params map[string]string) ([]Record, error) {
	path := "/api/v1/zones/" + url.PathEscape(zone) + "/records"

	if len(params) > 0 {
		query := url.Values{}
		for k, v := range params {
			if k == "type" {
				query.Set("record_type", v)
			} else {
				query.Set(k, v)
			}
		}
		path += "?" + query.Encode()
	}

	resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var records []Record
	if err := c.parseResponse(resp, &records); err != nil {
		return nil, err
	}

	return records, nil
}
