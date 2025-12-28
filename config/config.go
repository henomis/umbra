package config

// Config holds the configuration for the application.
type Config struct {
	ManifestPath   string
	InputFilePath  string
	OutputFilePath string
	ChunkSize      int64
	Chunks         int
	Copies         int
	Password       string
	Providers      []string
	Options        map[string]string
	Quiet          bool
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

	// Upload-specific validations
	// if c.InputFilePath != "" {
	// 	if c.ChunkSize == 0 && c.Chunks == 0 {
	// 		return ErrInvalidChunkConfig
	// 	} else if c.ChunkSize > 0 && c.Chunks > 0 {
	// 		return ErrInvalidChunkConfig
	// 	}

	// 	if c.Copies <= 0 {
	// 		return ErrInvalidCopies
	// 	}
	// }

	// Download-specific validations
	// if c.OutputFilePath != "" && c.InputFilePath == "" {
	// 	// This is a download scenario, no additional validations needed
	// }

	// At least one of InputFilePath or OutputFilePath should be set
	// if c.InputFilePath == "" && c.OutputFilePath == "" {
	// 	return ErrInvalidInputFilePath
	// } else if c.InputFilePath != "" && c.OutputFilePath != "" {
	// 	return ErrInvalidMode
	// }

	return nil
}
