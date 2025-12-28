package umbra

import (
	"github.com/vbauerster/mpb/v8"

	"github.com/henomis/umbra/config"
	"github.com/henomis/umbra/internal/provider"
)

// Umbra is the main struct that holds the configuration, providers, and logger.
type Umbra struct {
	config    *config.Config
	providers []provider.Provider
	progress  *mpb.Progress
}

// New creates a configured Umbra instance, validating the given configuration
// and initializing the logging and provider stack according to its settings.
func New(config *config.Config) (*Umbra, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	u := &Umbra{
		config:   config,
		progress: mpb.New(),
	}

	err := u.buildProviders()
	if err != nil {
		return nil, err
	}

	if config.Copies > len(u.providers) {
		return nil, ErrCopiesExceedProviders
	}

	return u, nil
}
