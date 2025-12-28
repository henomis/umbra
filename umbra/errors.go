package umbra

import "fmt"

// Error definitions.
var (
	ErrInvalidMode                   = fmt.Errorf("either download or upload mode must be specified")
	ErrUnknownProvider               = fmt.Errorf("unknown provider specified")
	ErrChunkSizeExceedsProviderLimit = fmt.Errorf("configured chunk size exceeds the maximum allowed by the specified providers")
	ErrCopiesExceedProviders         = fmt.Errorf("number of copies cannot exceed number of available providers")
	ErrOutputFileHashMismatch        = fmt.Errorf("output file hash does not match expected value")
)
