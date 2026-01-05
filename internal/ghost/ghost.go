package ghost

import "slices"

type Mode = string

const (
	Image  Mode = "image"
	QRCode Mode = "qrcode"
)

var ghostModes = []Mode{Image, QRCode}

// GhostModes returns the list of supported ghost modes.
func GhostModes() []Mode {
	return ghostModes
}

// IsValidGhostMode checks if the provided mode is a valid ghost mode.
func IsValidGhostMode(mode string) bool {
	return slices.Contains(ghostModes, mode)
}
