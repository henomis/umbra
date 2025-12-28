package pastecnetorg

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/henomis/umbra/internal/content"
	"github.com/henomis/umbra/internal/provider"
)

// Pastecnetorg implements the Provider interface.
type Pastecnetorg struct{}

var _ provider.Provider = (*Pastecnetorg)(nil)

const (
	tcpNetwork         = "tcp"
	termbinEndpoint    = "paste.c-net.org:9999"
	responseTrimCutset = "\x00\r\n "
)

// Meta holds the metadata for Pastecnetorg uploads.
type Meta struct {
	URL string `json:"url"`
}

// New creates a new Pastecnetorg provider instance.
func New(_ provider.Options) *Pastecnetorg {
	return &Pastecnetorg{}
}

// MaxSize returns the maximum allowed size for uploads.
func (p *Pastecnetorg) MaxSize() int64 {
	return 10 * 1024 * 1024 // 10 MB
}

// Name returns the provider name.
func (p *Pastecnetorg) Name() string {
	return provider.PASTECNETORG
}

// Expire returns the default expiration duration for uploads.
func (p *Pastecnetorg) Expire() time.Duration {
	// 1 week
	return 180 * 24 * time.Hour
}

// Upload sends data to termbin.com.
func (p *Pastecnetorg) Upload(ctx context.Context, payload []byte) (content.Meta, error) {
	// 1. Encode in Base64 to safely handle binary data
	encoded := base64.StdEncoding.EncodeToString(payload)

	// 2. TCP connection to Termbin
	dialer := &net.Dialer{}
	conn, err := dialer.DialContext(ctx, tcpNetwork, termbinEndpoint)
	if err != nil {
		return nil, wrapConnectionErr(err)
	}
	defer conn.Close()

	// 3. Send the data
	if _, err = conn.Write([]byte(encoded)); err != nil {
		return nil, wrapPayloadErr(err)
	}

	// 4. Read the response (the paste URL)
	response, err := io.ReadAll(conn)
	if err != nil {
		return nil, wrapResponseReadErr(err)
	}

	// Clean the URL (remove null bytes or spaces)
	url := string(bytes.TrimRight(response, responseTrimCutset))
	url = strings.TrimSpace(url)

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
func (p *Pastecnetorg) Download(ctx context.Context, meta content.Meta) ([]byte, error) {
	m := Meta{}
	if err := json.Unmarshal(meta, &m); err != nil {
		return nil, err
	}
	url := strings.TrimSpace(m.URL)
	if url == "" {
		return nil, ErrMetaURLMissing
	}

	// 1. HTTP GET request to retrieve the text
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, wrapDataFetchErr(url, err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, wrapDataFetchErr(url, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, wrapBodyReadErr(err)
	}

	// 2. Decode from Base64 back to the original binary
	decoded, err := base64.StdEncoding.DecodeString(string(body))
	if err != nil {
		return nil, wrapDecodeErr(err)
	}

	return decoded, nil
}
