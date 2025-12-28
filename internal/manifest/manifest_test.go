package manifest

import (
	"bytes"
	crypto_rand "crypto/rand"
	"encoding/binary"
	"errors"
	"io"
	"testing"

	cryptopkg "github.com/henomis/umbra/internal/crypto"
)

func TestManifestEncodeDecodeRoundTrip(t *testing.T) {
	m := newDeterministicManifest(t)
	payload := []byte("manifest secret payload")

	buf := new(bytes.Buffer)
	if err := m.Encode(buf, payload); err != nil {
		t.Fatalf("Encode returned error: %v", err)
	}

	decoded, err := m.Decode(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("Decode returned error: %v", err)
	}

	if !bytes.Equal(decoded, payload) {
		t.Fatalf("Decode mismatch: got %q want %q", decoded, payload)
	}
}

func TestManifestEncodeWritesHeader(t *testing.T) {
	m := newDeterministicManifest(t)

	buf := new(bytes.Buffer)
	if err := m.Encode(buf, []byte("data")); err != nil {
		t.Fatalf("Encode returned error: %v", err)
	}

	reader := bytes.NewReader(buf.Bytes())

	var header Header
	if err := binary.Read(reader, binary.LittleEndian, &header); err != nil {
		t.Fatalf("binary.Read header failed: %v", err)
	}

	if header.Magic != manifestMagic {
		t.Fatalf("Magic mismatch: got %v want %v", header.Magic, manifestMagic)
	}

	if header.Version != Version1 {
		t.Fatalf("Version mismatch: got %d want %d", header.Version, Version1)
	}

	var params cryptopkg.Parameters
	if err := binary.Read(reader, binary.LittleEndian, &params); err != nil {
		t.Fatalf("binary.Read params failed: %v", err)
	}

	if params != *m.crypto.Parameters() {
		t.Fatalf("Parameters mismatch: got %+v want %+v", params, m.crypto.Parameters())
	}
}

func TestManifestDecodeInvalidMagic(t *testing.T) {
	m := newDeterministicManifest(t)

	buf := new(bytes.Buffer)
	header := Header{Magic: [4]byte{0, 0, 0, 0}, Version: Version1}
	if err := binary.Write(buf, binary.LittleEndian, header); err != nil {
		t.Fatalf("binary.Write header failed: %v", err)
	}

	params := m.crypto.Parameters()
	if err := binary.Write(buf, binary.LittleEndian, params); err != nil {
		t.Fatalf("binary.Write params failed: %v", err)
	}

	buf.WriteString("ciphertext")

	if _, err := m.Decode(bytes.NewReader(buf.Bytes())); !errors.Is(err, ErrInvalidMagic) {
		t.Fatalf("Decode error = %v, want ErrInvalidMagic", err)
	}
}

func TestManifestDecodeUnsupportedVersion(t *testing.T) {
	m := newDeterministicManifest(t)

	buf := new(bytes.Buffer)
	header := Header{Magic: manifestMagic, Version: 42}
	if err := binary.Write(buf, binary.LittleEndian, header); err != nil {
		t.Fatalf("binary.Write header failed: %v", err)
	}

	params := m.crypto.Parameters()
	if err := binary.Write(buf, binary.LittleEndian, params); err != nil {
		t.Fatalf("binary.Write params failed: %v", err)
	}

	buf.WriteString("ciphertext")

	if _, err := m.Decode(bytes.NewReader(buf.Bytes())); !errors.Is(err, ErrUnsupportedVer) {
		t.Fatalf("Decode error = %v, want ErrUnsupportedVer", err)
	}
}

func TestManifestDecodeInvalidCryptoParams(t *testing.T) {
	m := newDeterministicManifest(t)

	buf := new(bytes.Buffer)
	header := Header{Magic: manifestMagic, Version: Version1}
	if err := binary.Write(buf, binary.LittleEndian, header); err != nil {
		t.Fatalf("binary.Write header failed: %v", err)
	}

	params := m.crypto.Parameters()
	params.KDF = 0
	if err := binary.Write(buf, binary.LittleEndian, params); err != nil {
		t.Fatalf("binary.Write params failed: %v", err)
	}

	buf.WriteString("ciphertext")

	if _, err := m.Decode(bytes.NewReader(buf.Bytes())); !errors.Is(err, ErrInvalidCryptoParams) {
		t.Fatalf("Decode error = %v, want ErrInvalidCryptoParams", err)
	}
}

func TestManifestDecodeDecryptError(t *testing.T) {
	m := newDeterministicManifest(t)

	buf := new(bytes.Buffer)
	if err := m.Encode(buf, []byte("payload")); err != nil {
		t.Fatalf("Encode returned error: %v", err)
	}

	tampered := append([]byte(nil), buf.Bytes()...)
	tampered[len(tampered)-1] ^= 0xFF

	if _, err := m.Decode(bytes.NewReader(tampered)); !errors.Is(err, ErrDecryptFailed) {
		t.Fatalf("Decode error = %v, want ErrDecryptFailed", err)
	}
}

func TestManifestEncodeWriterError(t *testing.T) {
	m := newDeterministicManifest(t)

	w := &limitedWriter{limit: 8}
	if err := m.Encode(w, []byte("payload")); err == nil {
		t.Fatal("Encode should fail when writer stops early")
	}
}

func TestManifestNewInitializesHeader(t *testing.T) {
	m := newDeterministicManifest(t)

	if m.header.Magic != manifestMagic {
		t.Fatalf("Manifest header magic mismatch: got %v want %v", m.header.Magic, manifestMagic)
	}

	if m.header.Version != Version1 {
		t.Fatalf("Manifest header version mismatch: got %d want %d", m.header.Version, Version1)
	}
}

func newDeterministicManifest(t *testing.T) *Manifest {
	t.Helper()

	password := []byte("test-password")
	mockRandReader(t, sequentialBytes(40))

	c, err := cryptopkg.New(password)
	if err != nil {
		t.Fatalf("crypto.New returned error: %v", err)
	}

	return New(c)
}

func mockRandReader(t *testing.T, payload []byte) {
	t.Helper()

	original := crypto_rand.Reader
	crypto_rand.Reader = bytes.NewReader(payload)
	t.Cleanup(func() {
		crypto_rand.Reader = original
	})
}

func sequentialBytes(n int) []byte {
	data := make([]byte, n)
	for i := range data {
		data[i] = byte(i + 1)
	}
	return data
}

type limitedWriter struct {
	written int
	limit   int
}

func (w *limitedWriter) Write(p []byte) (int, error) {
	if w.written >= w.limit {
		return 0, io.ErrShortWrite
	}

	space := w.limit - w.written
	if space > len(p) {
		space = len(p)
	}

	w.written += space

	if space < len(p) {
		return space, io.ErrShortWrite
	}

	return space, nil
}
