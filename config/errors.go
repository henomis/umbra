package config

import "fmt"

// Config errors.
var (
	ErrInvalidInputFilePath  = fmt.Errorf("input file path must not be empty")
	ErrInvalidOutputFilePath = fmt.Errorf("output file path must not be empty")
	ErrInvalidMode           = fmt.Errorf("either download or upload mode must be specified")
	ErrInvalidChunkConfig    = fmt.Errorf("either ChunkSize or Chunks must be specified")
	ErrInvalidCopies         = fmt.Errorf("copies must be a positive integer")
	ErrInvalidPassword       = fmt.Errorf("password must not be empty")
	ErrInvalidManifestPath   = fmt.Errorf("manifest path must not be empty")
)
