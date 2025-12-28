package provider

import (
	"context"
	"time"

	"github.com/henomis/umbra/internal/content"
)

// Provider defines the interface that all providers must implement.
type Provider interface {
	Name() string
	Upload(context.Context, []byte) (content.Meta, error)
	Download(context.Context, content.Meta) ([]byte, error)
	MaxSize() int64
	Expire() time.Duration
}

// Options represents provider-specific configuration options.
type Options = map[string]string

// DefaultProviders is the list of providers used if none are specified in the configuration.
var DefaultProviders = []string{
	// "dummy", for testing purposes only
	"termbin",
	"clbin",
	"pipfi",
	"pastecnetorg",
}

// Provider names.
const (
	DUMMY        = "dummy"
	TERMBIN      = "termbin"
	CLBIN        = "clbin"
	PIPFI        = "pipfi"
	PASTECNETORG = "pastecnetorg"
)
