package transport

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"agent/internal/model"
)

const (
	defaultMaxErrorBodyBytes = 8192
	gzipMinBodyBytes         = 32 * 1024
)

type Client struct {
	baseURL           string
	token             string
	httpClient        *http.Client
	maxErrorBodyBytes int64
}

func New(baseURL, token string, timeout time.Duration) *Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}

	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		httpClient: &http.Client{
			Timeout:   timeout,
			Transport: tr,
		},
		maxErrorBodyBytes: defaultMaxErrorBodyBytes,
	}
}

func (c *Client) PushMetrics(ctx context.Context, req model.PushMetricsRequest) error {
	headers := map[string]string{
		"X-Metrics-Source": req.Source,
	}

	if !req.BatchCollectedAt.IsZero() {
		headers["X-Batch-Collected-At"] = req.BatchCollectedAt.UTC().Format(time.RFC3339Nano)
	}

	return c.postJSON(ctx, "/agent/metrics", req, headers)
}

func (c *Client) PushInventory(ctx context.Context, req model.PushInventoryRequest) error {
	headers := map[string]string{}

	if req.RevisionHash != "" {
		headers["X-Inventory-Revision"] = req.RevisionHash
	}
	if !req.CollectedAt.IsZero() {
		headers["X-Inventory-Collected-At"] = req.CollectedAt.UTC().Format(time.RFC3339Nano)
	}

	return c.postJSON(ctx, "/agent/inventory", req, headers)
}

func (c *Client) Heartbeat(ctx context.Context) error {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/agent/heartbeat", nil)
	if err != nil {
		return err
	}

	c.applyCommonHeaders(httpReq.Header)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := c.checkResponse(resp, http.MethodPost, "/agent/heartbeat"); err != nil {
		return err
	}

	return nil
}

func (c *Client) postJSON(ctx context.Context, path string, payload any, extraHeaders map[string]string) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	reader := bytes.NewReader(body)
	var requestBody io.Reader = reader
	useGzip := len(body) >= gzipMinBodyBytes

	if useGzip {
		compressed, err := gzipBytes(body)
		if err != nil {
			return err
		}
		requestBody = bytes.NewReader(compressed)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, requestBody)
	if err != nil {
		return err
	}

	c.applyCommonHeaders(httpReq.Header)
	httpReq.Header.Set("Content-Type", "application/json")
	if useGzip {
		httpReq.Header.Set("Content-Encoding", "gzip")
	}

	for k, v := range extraHeaders {
		if strings.TrimSpace(k) == "" || strings.TrimSpace(v) == "" {
			continue
		}
		httpReq.Header.Set(k, v)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := c.checkResponse(resp, http.MethodPost, path); err != nil {
		return err
	}

	return nil
}

func (c *Client) applyCommonHeaders(h http.Header) {
	h.Set("Authorization", "Bearer "+c.token)
	h.Set("Accept", "application/json")
	h.Set("User-Agent", "kubeevalhub-agent/1.0")
}

func (c *Client) checkResponse(resp *http.Response, method, path string) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	bodySnippet, readErr := readBodySnippet(resp.Body, c.maxErrorBodyBytes)
	if readErr != nil {
		return fmt.Errorf("%s %s failed: status=%d, could not read error body: %w", method, path, resp.StatusCode, readErr)
	}

	if bodySnippet == "" {
		return fmt.Errorf("%s %s failed: status=%d", method, path, resp.StatusCode)
	}

	return fmt.Errorf("%s %s failed: status=%d body=%q", method, path, resp.StatusCode, bodySnippet)
}

func readBodySnippet(r io.Reader, maxBytes int64) (string, error) {
	if r == nil {
		return "", nil
	}
	if maxBytes <= 0 {
		maxBytes = defaultMaxErrorBodyBytes
	}

	limited := io.LimitReader(r, maxBytes)
	b, err := io.ReadAll(limited)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(b)), nil
}

func gzipBytes(src []byte) ([]byte, error) {
	var buf bytes.Buffer

	zw := gzip.NewWriter(&buf)
	if _, err := zw.Write(src); err != nil {
		_ = zw.Close()
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
