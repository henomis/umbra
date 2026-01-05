package umbra

import (
	"crypto/rand"
	"math/big"
	"time"

	"github.com/henomis/umbra/internal/provider"
	"github.com/henomis/umbra/internal/provider/clbin"
	"github.com/henomis/umbra/internal/provider/pastecnetorg"
	"github.com/henomis/umbra/internal/provider/pipfi"
	"github.com/henomis/umbra/internal/provider/termbin"
)

// buildProviders initializes the configured providers list, applying defaults
// when none are specified and returning an error for unknown providers.
func (u *Umbra) buildProviders() error {
	var providers []provider.Provider

	if len(u.config.Providers) == 0 {
		u.config.Providers = provider.DefaultProviders
	}

	for _, p := range u.config.Providers {
		switch p {
		case provider.TERMBIN:
			providers = append(providers, termbin.New())
		case provider.CLBIN:
			providers = append(providers, clbin.New())
		case provider.PIPFI:
			providers = append(providers, pipfi.New())
		case provider.PASTECNETORG:
			providers = append(providers, pastecnetorg.New())
		default:
			return ErrUnknownProvider
		}
	}

	u.providers = providers
	return nil
}

// getMaxChunkSizeForProviders returns the smallest MaxSize value among the configured providers.
// It is used to ensure data chunks respect the most restrictive provider limit.
func (u *Umbra) getMaxChunkSizeForProviders() int64 {
	var maxSize int64 = -1

	for _, p := range u.providers {
		pMax := p.MaxSize()
		if maxSize < 0 || pMax < maxSize {
			maxSize = pMax
		}
	}

	return maxSize
}

// getRandomProvider selects and returns a random provider from the configured set.
// It returns ErrUnknownProvider when no providers are available, and falls back to
// the first provider if randomness cannot be obtained.
func (u *Umbra) getRandomProvider() (provider.Provider, error) {
	if len(u.providers) == 0 {
		return nil, ErrUnknownProvider
	}

	providerCount := big.NewInt(int64(len(u.providers)))
	idx, err := rand.Int(rand.Reader, providerCount)
	if err != nil {
		return nil, err
	}

	return u.providers[idx.Int64()], nil
}

func (u *Umbra) getUniqueRadomProvider(providers []provider.Provider) (provider.Provider, error) {
	for {
		p, err := u.getRandomProvider()
		if err != nil {
			return nil, err
		}

		unique := true
		for _, existing := range providers {
			if existing.Name() == p.Name() {
				unique = false
				break
			}
		}

		if unique {
			return p, nil
		}
	}
}

func (u *Umbra) getProviderByName(name string) (provider.Provider, error) {
	for _, p := range u.providers {
		if p.Name() == name {
			return p, nil
		}
	}

	return nil, ErrUnknownProvider
}

func (u *Umbra) getProviderMinExpireDuration() time.Duration {
	var minExpire time.Duration = -1

	for _, p := range u.providers {
		expire := p.Expire()
		if minExpire < 0 || (expire > 0 && expire < minExpire) {
			minExpire = expire
		}
	}

	return minExpire
}
