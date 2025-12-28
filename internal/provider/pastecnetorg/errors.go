package pastecnetorg

import (
	"errors"
	"fmt"
)

// Error definitions.
var (
	ErrConnectionFailed   = errors.New("termbin: connection failed")
	ErrPayloadSendFailed  = errors.New("termbin: unable to send payload")
	ErrResponseReadFailed = errors.New("termbin: unable to read response")
	ErrMetaURLMissing     = errors.New("termbin: url missing or invalid in meta")
	ErrDataFetchFailed    = errors.New("termbin: unable to fetch data")
	ErrBodyReadFailed     = errors.New("termbin: unable to read response body")
	ErrDecodeFailed       = errors.New("termbin: unable to decode base64 content")
)

func wrapConnectionErr(err error) error {
	return errors.Join(ErrConnectionFailed, fmt.Errorf("dial %s over %s: %w", termbinEndpoint, tcpNetwork, err))
}

func wrapPayloadErr(err error) error {
	return errors.Join(ErrPayloadSendFailed, fmt.Errorf("write payload: %w", err))
}

func wrapResponseReadErr(err error) error {
	return errors.Join(ErrResponseReadFailed, fmt.Errorf("read response: %w", err))
}

func wrapDataFetchErr(resource string, err error) error {
	return errors.Join(ErrDataFetchFailed, fmt.Errorf("fetch %s: %w", resource, err))
}

func wrapBodyReadErr(err error) error {
	return errors.Join(ErrBodyReadFailed, fmt.Errorf("read body: %w", err))
}

func wrapDecodeErr(err error) error {
	return errors.Join(ErrDecodeFailed, fmt.Errorf("decode base64 payload: %w", err))
}
