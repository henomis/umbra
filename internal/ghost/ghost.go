package ghost

import "slices"

// Mode represents the mode of ghost generation.
type Mode = string

// Supported ghost modes.
const (
	Image  Mode = "image"
	QRCode Mode = "qrcode"
)

var ghostModes = []Mode{Image, QRCode}

// Modes returns the list of supported ghost modes.
func Modes() []Mode {
	return ghostModes
}

// IsValidGhostMode checks if the provided mode is a valid ghost mode.
func IsValidGhostMode(mode string) bool {
	return slices.Contains(ghostModes, mode)
}
