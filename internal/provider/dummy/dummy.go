package dummy

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/henomis/umbra/internal/content"
	"github.com/henomis/umbra/internal/provider"
)

// Dummy implements the Provider interface for testing purposes.
type Dummy struct {
	basePath string
}

// Meta holds the metadata for Dummy uploads.
type Meta struct {
	Path string `json:"path"`
}

const (
	basePath = "dummy-path"
)

var _ provider.Provider = (*Dummy)(nil)

// New creates a new Dummy provider instance.
func New(options map[string]string) *Dummy {
	path := "/tmp"
	if p, ok := options[basePath]; ok {
		path = p
	}

	return &Dummy{
		basePath: path,
	}
}

// Name returns the provider name.
func (d *Dummy) Name() string {
	return provider.DUMMY
}

// Expire returns the default expiration duration for uploads.
func (d *Dummy) Expire() time.Duration {
	return 0
}

// MaxSize returns the maximum allowed size for uploads.
func (d *Dummy) MaxSize() int64 {
	return 100 * 1024 * 1024 // 100 MB
}

// Upload saves data to a temporary file and returns its path in the metadata.
func (d *Dummy) Upload(_ context.Context, payload []byte) (content.Meta, error) {
	tmpFile, err := os.CreateTemp(d.basePath, "umbra-dummy-*")
	if err != nil {
		return nil, err
	}

	if _, err := tmpFile.Write(payload); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return nil, err
	}

	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpFile.Name())
		return nil, err
	}

	meta := Meta{
		Path: tmpFile.Name(),
	}

	metaBytes, err := json.Marshal(meta)
	if err != nil {
		os.Remove(tmpFile.Name())
		return nil, err
	}

	return content.Meta(metaBytes), nil
}

// Download retrieves data from a file specified in the metadata.
func (d *Dummy) Download(_ context.Context, meta content.Meta) ([]byte, error) {
	m := Meta{}
	if err := json.Unmarshal(meta, &m); err != nil {
		return nil, err
	}

	if m.Path == "" {
		return nil, ErrPathRequired
	}

	return os.ReadFile(m.Path)
}
