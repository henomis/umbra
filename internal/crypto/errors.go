package crypto

import "errors"

// Crypto errors.
var (
	ErrUnsupportedKDF    = errors.New("manifest: unsupported KDF")
	ErrUnsupportedCipher = errors.New("manifest: unsupported cipher")
)
