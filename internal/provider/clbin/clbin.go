package clbin

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/henomis/umbra/internal/content"
	"github.com/henomis/umbra/internal/provider"
)

const (
	defaultBaseURL = "https://clbin.com"
	defaultTimeout = 15 * time.Second
	formFieldName  = "clbin"
	maxSizeBytes   = 10 * 1024 * 1024
)

// Clbin implements provider.Provider using clbin.com.
type Clbin struct {
	client  *http.Client
	baseURL string
}

// Meta holds the metadata for Clbin uploads.
type Meta struct {
	URL string `json:"url"`
}

var _ provider.Provider = (*Clbin)(nil)

// New creates a new Clbin provider instance.
func New(_ provider.Options) *Clbin {
	return &Clbin{
		baseURL: defaultBaseURL,
		client: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

// Name returns the provider name.
func (c *Clbin) Name() string {
	return provider.CLBIN
}

// MaxSize returns the maximum allowed size for uploads.
func (c *Clbin) MaxSize() int64 {
	return maxSizeBytes
}

// Expire returns the default expiration duration for uploads.
func (c *Clbin) Expire() time.Duration {
	return 0
}

// Upload sends data to clbin.com.
func (c *Clbin) Upload(ctx context.Context, payload []byte) (content.Meta, error) {
	encoded := base64.StdEncoding.EncodeToString(payload)

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormField(formFieldName)
	if err != nil {
		return nil, err
	}

	if _, err := part.Write([]byte(encoded)); err != nil {
		return nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL,
		&buf,
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	url, err := extractURL(string(body))
	if err != nil {
		return nil, err
	}

	meta := Meta{
		URL: url,
	}

	metaBytes, err := json.Marshal(meta)
	if err != nil {
		return nil, err
	}

	return content.Meta(metaBytes), nil
}

// Download fetches the data from the URL stored in Meta.
func (c *Clbin) Download(ctx context.Context, meta content.Meta) ([]byte, error) {
	m := Meta{}
	if err := json.Unmarshal(meta, &m); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, m.URL, http.NoBody)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	encoded, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	data, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(encoded)))
	if err != nil {
		return nil, err
	}

	return data, nil
}

func extractURL(body string) (string, error) {
	lines := strings.Split(strings.TrimSpace(body), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if strings.HasPrefix(line, "hxxps://") || strings.HasPrefix(line, "https://") {
			return strings.Replace(line, "hxxps://", "https://", 1), nil
		}
	}
	return "", errors.New("clbin: no URL found in response")
}
