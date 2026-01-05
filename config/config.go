package config

import "github.com/henomis/umbra/internal/ghost"

// Config holds the configuration for the application.
type Config struct {
	ManifestPath string
	Password     string
	Quiet        bool
	Providers    []string
	// Options      map[string]string // for future use
	GhostMode string

	Upload   *Upload
	Download *Download
}

// Upload holds the upload-specific configuration.
type Upload struct {
	InputFilePath string
	ChunkSize     int64
	Chunks        int
	Copies        int
}

// Download holds the download-specific configuration.
type Download struct {
	OutputFilePath string
}

// Validate checks the configuration for validity.
func (c *Config) Validate() error {
	// Common validations
	if c.ManifestPath == "" {
		return ErrInvalidInputFilePath
	}

	if c.Password == "" {
		return ErrInvalidPassword
	}

	// Mode-specific validations
	if c.Upload != nil && c.Download != nil {
		return ErrInvalidMode
	}

	if c.Upload != nil {
		// Upload-specific validations
		if c.Upload.InputFilePath == "" {
			return ErrInvalidInputFilePath
		}
		if c.Upload.ChunkSize == 0 && c.Upload.Chunks == 0 {
			return ErrInvalidChunkConfig
		} else if c.Upload.ChunkSize != 0 && c.Upload.Chunks != 0 {
			return ErrInvalidChunkConfig
		}

		if c.Upload.Copies <= 0 {
			return ErrInvalidCopies
		}

		if c.GhostMode != "" && !ghost.IsValidGhostMode(c.GhostMode) {
			return ErrInvalidGhostMode
		}
	}

	if c.Download != nil {
		// Download-specific validations
		if c.Download.OutputFilePath == "" {
			return ErrInvalidOutputFilePath
		}

		if c.GhostMode != "" && !ghost.IsValidGhostMode(c.GhostMode) {
			return ErrInvalidGhostMode
		}
	}

	return nil
}
