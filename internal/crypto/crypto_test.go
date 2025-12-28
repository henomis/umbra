package crypto

import (
	"bytes"
	crypto_rand "crypto/rand"
	"encoding/hex"
	"testing"
)

func mockRandReader(t *testing.T, payload []byte) {
	t.Helper()

	originalReader := crypto_rand.Reader
	crypto_rand.Reader = bytes.NewReader(payload)
	t.Cleanup(func() {
		crypto_rand.Reader = originalReader
	})
}

func sequentialBytes(n int) []byte {
	data := make([]byte, n)
	for i := range data {
		data[i] = byte(i + 1)
	}
	return data
}

func newCrypto(t *testing.T) *Crypto {
	t.Helper()

	c, err := New([]byte("unit-test-password"))
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	return c
}

func TestNew(t *testing.T) {
	password := []byte("test-password")
	c, err := New(password)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	if c == nil {
		t.Fatal("New returned nil crypto")
	}

	if c.parameters == nil {
		t.Fatal("New returned nil parameters")
	}

	if c.parameters.KDF != KDFArgon2id {
		t.Errorf("KDF = %d, want %d", c.parameters.KDF, KDFArgon2id)
	}

	if c.parameters.Cipher != CipherXChaCha20Poly1305 {
		t.Errorf("Cipher = %d, want %d", c.parameters.Cipher, CipherXChaCha20Poly1305)
	}

	// Verify salt and nonce are not all zeros
	var zeroSalt [16]byte
	var zeroNonce [24]byte
	if bytes.Equal(c.parameters.Salt[:], zeroSalt[:]) {
		t.Error("Salt should not be all zeros")
	}

	if bytes.Equal(c.parameters.Nonce[:], zeroNonce[:]) {
		t.Error("Nonce should not be all zeros")
	}
}

func TestNewGeneratesDeterministicParameters(t *testing.T) {
	randomData := sequentialBytes(40)
	mockRandReader(t, randomData)

	password := []byte("deterministic-password")
	c, err := New(password)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	var expectedSalt [16]byte
	copy(expectedSalt[:], randomData[:16])
	var expectedNonce [24]byte
	copy(expectedNonce[:], randomData[16:])

	if !bytes.Equal(c.parameters.Salt[:], expectedSalt[:]) {
		t.Errorf("Salt mismatch: got %x want %x", c.parameters.Salt, expectedSalt)
	}

	if !bytes.Equal(c.parameters.Nonce[:], expectedNonce[:]) {
		t.Errorf("Nonce mismatch: got %x want %x", c.parameters.Nonce, expectedNonce)
	}

	if params := c.Parameters(); params != c.parameters {
		t.Errorf("Parameters() returned %v, want %v", params, c.parameters)
	}
}

func TestParameters(t *testing.T) {
	c := newCrypto(t)

	params := c.Parameters()
	if params == nil {
		t.Fatal("Parameters() returned nil")
	}

	if params != c.parameters {
		t.Error("Parameters() should return the same pointer as internal parameters")
	}
}

func TestSetParameters(t *testing.T) {
	c := newCrypto(t)

	newParams := &Parameters{
		KDF:    KDFArgon2id,
		Cipher: CipherXChaCha20Poly1305,
		Salt:   [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		Nonce:  [24]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24},
	}

	err := c.SetParameters(newParams)
	if err != nil {
		t.Fatalf("SetParameters returned error: %v", err)
	}

	if c.parameters != newParams {
		t.Error("SetParameters did not update parameters")
	}
}

func TestSetParametersUnsupportedKDF(t *testing.T) {
	c := newCrypto(t)

	// Save original parameters
	originalParams := c.parameters

	// Create parameters with unsupported KDF
	c.parameters = &Parameters{
		KDF:    99, // Unsupported KDF
		Cipher: CipherXChaCha20Poly1305,
	}

	newParams := &Parameters{
		KDF:    KDFArgon2id,
		Cipher: CipherXChaCha20Poly1305,
	}

	err := c.SetParameters(newParams)
	if err != ErrUnsupportedKDF {
		t.Errorf("SetParameters error = %v, want %v", err, ErrUnsupportedKDF)
	}

	// Restore original parameters
	c.parameters = originalParams
}

func TestSetParametersUnsupportedCipher(t *testing.T) {
	c := newCrypto(t)

	// Save original parameters
	originalParams := c.parameters

	// Create parameters with unsupported cipher
	c.parameters = &Parameters{
		KDF:    KDFArgon2id,
		Cipher: 99, // Unsupported cipher
	}

	newParams := &Parameters{
		KDF:    KDFArgon2id,
		Cipher: CipherXChaCha20Poly1305,
	}

	err := c.SetParameters(newParams)
	if err != ErrUnsupportedCipher {
		t.Errorf("SetParameters error = %v, want %v", err, ErrUnsupportedCipher)
	}

	// Restore original parameters
	c.parameters = originalParams
}

func TestEncodeDecodeRoundTrip(t *testing.T) {
	c := newCrypto(t)

	plaintext := []byte("this is a test payload")
	aad := []byte("associated data")

	ciphertext, err := c.Encode(plaintext, aad)
	if err != nil {
		t.Fatalf("Encode returned error: %v", err)
	}

	decoded, err := c.Decode(ciphertext, aad)
	if err != nil {
		t.Fatalf("Decode returned error: %v", err)
	}

	if !bytes.Equal(decoded, plaintext) {
		t.Errorf("Decode mismatch: got %s want %s", decoded, plaintext)
	}
}

func TestEncodeDecodeEmptyPlaintext(t *testing.T) {
	c := newCrypto(t)

	plaintext := []byte("")
	aad := []byte("aad")

	ciphertext, err := c.Encode(plaintext, aad)
	if err != nil {
		t.Fatalf("Encode returned error: %v", err)
	}

	decoded, err := c.Decode(ciphertext, aad)
	if err != nil {
		t.Fatalf("Decode returned error: %v", err)
	}

	if !bytes.Equal(decoded, plaintext) {
		t.Errorf("Decode mismatch: got %s want %s", decoded, plaintext)
	}
}

func TestEncodeDecodeNilAAD(t *testing.T) {
	c := newCrypto(t)

	plaintext := []byte("test with nil aad")

	ciphertext, err := c.Encode(plaintext, nil)
	if err != nil {
		t.Fatalf("Encode returned error: %v", err)
	}

	decoded, err := c.Decode(ciphertext, nil)
	if err != nil {
		t.Fatalf("Decode returned error: %v", err)
	}

	if !bytes.Equal(decoded, plaintext) {
		t.Errorf("Decode mismatch: got %s want %s", decoded, plaintext)
	}
}

func TestEncodeDecodeWrongAAD(t *testing.T) {
	c := newCrypto(t)

	plaintext := []byte("test payload")
	aad := []byte("correct aad")

	ciphertext, err := c.Encode(plaintext, aad)
	if err != nil {
		t.Fatalf("Encode returned error: %v", err)
	}

	wrongAAD := []byte("wrong aad")
	_, err = c.Decode(ciphertext, wrongAAD)
	if err == nil {
		t.Fatal("Decode should fail with wrong AAD")
	}
}

func TestDecodeTamperedCiphertext(t *testing.T) {
	c := newCrypto(t)

	plaintext := []byte("integrity protected")
	aad := []byte("aad")

	ciphertext, err := c.Encode(plaintext, aad)
	if err != nil {
		t.Fatalf("Encode returned error: %v", err)
	}

	tampered := append([]byte(nil), ciphertext...)
	tampered[0] ^= 0xff

	if _, err := c.Decode(tampered, aad); err == nil {
		t.Fatal("Decode should fail when ciphertext is tampered")
	}
}

func TestDecodeTamperedTag(t *testing.T) {
	c := newCrypto(t)

	plaintext := []byte("integrity protected")
	aad := []byte("aad")

	ciphertext, err := c.Encode(plaintext, aad)
	if err != nil {
		t.Fatalf("Encode returned error: %v", err)
	}

	// Tamper with the authentication tag (last 16 bytes)
	if len(ciphertext) > 16 {
		tampered := append([]byte(nil), ciphertext...)
		tampered[len(tampered)-1] ^= 0xff

		if _, err := c.Decode(tampered, aad); err == nil {
			t.Fatal("Decode should fail when authentication tag is tampered")
		}
	}
}

func TestEncodeDifferentPasswords(t *testing.T) {
	password1 := []byte("password1")
	password2 := []byte("password2")

	c1, err := New(password1)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	c2, err := New(password2)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	// Use same parameters for both
	c2.parameters = c1.parameters

	plaintext := []byte("test message")
	aad := []byte("aad")

	ciphertext1, err := c1.Encode(plaintext, aad)
	if err != nil {
		t.Fatalf("Encode returned error: %v", err)
	}

	ciphertext2, err := c2.Encode(plaintext, aad)
	if err != nil {
		t.Fatalf("Encode returned error: %v", err)
	}

	// Different passwords should produce different ciphertexts
	if bytes.Equal(ciphertext1, ciphertext2) {
		t.Error("Different passwords should produce different ciphertexts")
	}

	// Decoding with wrong password should fail
	if _, err := c1.Decode(ciphertext2, aad); err == nil {
		t.Error("Decode should fail with wrong password")
	}
}

func TestDeriveKeyDeterministic(t *testing.T) {
	password := []byte("test-password")
	salt := []byte("0123456789abcdef")

	key := deriveKey(password, salt)
	expected, err := hex.DecodeString("556da3a97c3f3953bc6ebdd6a07c575c4f4dcd125ad90a23af94e3c28f7fc2de")
	if err != nil {
		t.Fatalf("failed to decode expected key: %v", err)
	}

	if !bytes.Equal(key, expected) {
		t.Errorf("deriveKey() = %x, want %x", key, expected)
	}

	if len(key) != 32 {
		t.Errorf("deriveKey() returned %d-byte key, want 32", len(key))
	}
}

func TestDeriveKeySameInputsSameOutput(t *testing.T) {
	password := []byte("same-password")
	salt := []byte("same-salt-value")

	key1 := deriveKey(password, salt)
	key2 := deriveKey(password, salt)

	if !bytes.Equal(key1, key2) {
		t.Error("deriveKey should produce same output for same inputs")
	}
}

func TestDeriveKeyDifferentSalts(t *testing.T) {
	password := []byte("password")
	salt1 := []byte("salt1-----------")
	salt2 := []byte("salt2-----------")

	key1 := deriveKey(password, salt1)
	key2 := deriveKey(password, salt2)

	if bytes.Equal(key1, key2) {
		t.Error("deriveKey should produce different keys for different salts")
	}
}

func TestEncodeLargePayload(t *testing.T) {
	c := newCrypto(t)

	// Test with a large payload
	plaintext := make([]byte, 1024*1024) // 1 MB
	for i := range plaintext {
		plaintext[i] = byte(i % 256)
	}
	aad := []byte("large payload test")

	ciphertext, err := c.Encode(plaintext, aad)
	if err != nil {
		t.Fatalf("Encode returned error: %v", err)
	}

	decoded, err := c.Decode(ciphertext, aad)
	if err != nil {
		t.Fatalf("Decode returned error: %v", err)
	}

	if !bytes.Equal(decoded, plaintext) {
		t.Error("Decode failed for large payload")
	}
}
