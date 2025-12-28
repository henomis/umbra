package dummy

import "errors"

// Error definitions.
var (
	ErrPathRequired = errors.New("dummy: path required")
	ErrInvalidData  = errors.New("dummy: invalid data")
)
