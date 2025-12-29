// manifest.go
package manifest

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/henomis/umbra/internal/crypto"
)

var manifestMagic = [4]byte{0x86, 0x90, 0x99, 0x8b}

const (
	// Version1 indicates the first version of the manifest format.
	Version1 uint32 = 1
)

// Header represents the manifest file header.
type Header struct {
	Magic   [4]byte
	Version uint32
}

// Manifest represents the manifest structure.
type Manifest struct {
	header Header
	crypto *crypto.Crypto
}

// New creates a new Manifest instance.
func New(crypto *crypto.Crypto) *Manifest {
	return &Manifest{
		header: Header{
			Magic:   manifestMagic,
			Version: Version1,
		},
		crypto: crypto,
	}
}

// Version returns the manifest version.
func (m *Manifest) Version() uint32 {
	return m.header.Version
}

// CryptoParameters returns the crypto parameters used in the manifest.
func (m *Manifest) CryptoParameters() *crypto.Parameters {
	return m.crypto.Parameters()
}

// Encode writes the manifest to the provided writer.
func (m *Manifest) Encode(w io.Writer, content []byte) error {
	headerBuf := new(bytes.Buffer)
	if err := binary.Write(headerBuf, binary.LittleEndian, m.header); err != nil {
		return err
	}

	if err := binary.Write(headerBuf, binary.LittleEndian, m.crypto.Parameters()); err != nil {
		return err
	}

	ciphertext, err := m.crypto.Encode(content, headerBuf.Bytes())
	if err != nil {
		return err
	}

	// Write manifest
	if _, err := w.Write(headerBuf.Bytes()); err != nil {
		return err
	}
	if _, err := w.Write(ciphertext); err != nil {
		return err
	}

	return nil
}

// Decode reads and decrypts the manifest from the provided reader.
func (m *Manifest) Decode(r io.Reader) ([]byte, error) {
	// Read header
	var header Header
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	if header.Magic != manifestMagic {
		return nil, ErrInvalidMagic
	}

	if header.Version != Version1 {
		return nil, ErrUnsupportedVer
	}

	// Read crypto params
	var cryptoParameters crypto.Parameters
	if err := binary.Read(r, binary.LittleEndian, &cryptoParameters); err != nil {
		return nil, err
	}

	if err := m.crypto.SetParameters(&cryptoParameters); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidCryptoParams, err)
	}

	// Read remaining as ciphertext
	ciphertext, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	headerBuf := new(bytes.Buffer)
	if err := binary.Write(headerBuf, binary.LittleEndian, header); err != nil {
		return nil, err
	}

	if err := binary.Write(headerBuf, binary.LittleEndian, cryptoParameters); err != nil {
		return nil, err
	}

	content, err := m.crypto.Decode(ciphertext, headerBuf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDecryptFailed, err)
	}

	return content, nil
}
