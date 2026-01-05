package pipfi

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/henomis/umbra/internal/content"
	"github.com/henomis/umbra/internal/provider"
)

const (
	defaultBaseURL = "http://p.ip.fi"
	defaultTimeout = 15 * time.Second
	maxSizeBytes   = 10 * 1024 * 1024
	formFieldName  = "paste"
	userAgent      = "Wget/1.21.1 (linux-gnu)"
)

var _ provider.Provider = (*Pipfi)(nil)

// Pipfi implements the Provider interface.
type Pipfi struct {
	client  *http.Client
	baseURL string
}

// Meta holds the metadata for Pipfi uploads.
type Meta struct {
	URL string `json:"url"`
}

// New creates a new Pipfi provider instance.
func New() *Pipfi {
	return &Pipfi{
		baseURL: defaultBaseURL,
		client: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

// Name returns the provider name.
func (p *Pipfi) Name() string {
	return provider.PIPFI
}

// Upload sends data to p.ip.fi.
func (p *Pipfi) Upload(ctx context.Context, payload []byte) (json.RawMessage, error) {
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
		p.baseURL,
		&buf,
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("User-Agent", userAgent)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	url := strings.TrimSpace(string(body))

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
func (p *Pipfi) Download(ctx context.Context, meta json.RawMessage) ([]byte, error) {
	m := Meta{}
	if err := json.Unmarshal(meta, &m); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, m.URL, http.NoBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := p.client.Do(req)
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

// MaxSize returns the maximum allowed size for uploads.
func (p *Pipfi) MaxSize() int64 {
	// Not documented; conservative guess
	return maxSizeBytes
}

// Expire returns the default expiration duration for uploads.
func (p *Pipfi) Expire() time.Duration {
	// Retention not guaranteed
	return 0
}
