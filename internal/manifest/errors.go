package manifest

import "errors"

// Manifest errors.
var (
	ErrInvalidMagic        = errors.New("manifest: invalid magic")
	ErrUnsupportedVer      = errors.New("manifest: unsupported version")
	ErrInvalidCryptoParams = errors.New("manifest: invalid crypto parameters")
	ErrDecryptFailed       = errors.New("manifest: decrypt failed")
)
