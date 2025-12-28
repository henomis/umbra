package crypto

import (
	"crypto/rand"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/chacha20poly1305"
)

// Crypto constants.
const (
	KDFArgon2id             = 1
	CipherXChaCha20Poly1305 = 1
)

// Crypto represents the crypto structure.
type Crypto struct {
	parameters *Parameters
	password   []byte
}

// Parameters holds the crypto parameters.
type Parameters struct {
	KDF    uint8
	Cipher uint8
	Salt   [16]byte
	Nonce  [24]byte
}

// New creates a new Crypto instance with generated parameters.
func New(password []byte) (*Crypto, error) {
	crypto := &Crypto{
		parameters: &Parameters{
			KDF:    KDFArgon2id,
			Cipher: CipherXChaCha20Poly1305,
			Salt:   [16]byte{},
			Nonce:  [24]byte{},
		},
	}

	if _, err := rand.Read(crypto.parameters.Salt[:]); err != nil {
		return nil, err
	}

	if _, err := rand.Read(crypto.parameters.Nonce[:]); err != nil {
		return nil, err
	}

	crypto.password = password

	return crypto, nil
}

// SetParameters sets the crypto parameters.
func (c *Crypto) SetParameters(parameters *Parameters) error {
	if c.parameters.KDF != KDFArgon2id {
		return ErrUnsupportedKDF
	}
	if c.parameters.Cipher != CipherXChaCha20Poly1305 {
		return ErrUnsupportedCipher
	}
	c.parameters = parameters

	return nil
}

// Parameters returns the crypto parameters.
func (c *Crypto) Parameters() *Parameters {
	return c.parameters
}

// Encode encrypts the given content using the stored parameters and password.
func (c *Crypto) Encode(content, additionalData []byte) ([]byte, error) {
	key := deriveKey(c.password, c.parameters.Salt[:])

	// Encrypt payload
	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, err
	}

	ciphertext := aead.Seal(nil, c.parameters.Nonce[:], content, additionalData) // AAD = header + crypto params
	return ciphertext, nil
}

// Decode decrypts the given ciphertext using the stored parameters and password.
func (c *Crypto) Decode(ciphertext, additionalData []byte) ([]byte, error) {
	key := deriveKey(c.password, c.parameters.Salt[:])

	// Decrypt payload
	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, err
	}

	plaintext, err := aead.Open(nil, c.parameters.Nonce[:], ciphertext, additionalData) // AAD = header + crypto params
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// deriveKey derives a key from the given password and salt using Argon2id.
func deriveKey(password, salt []byte) []byte {
	return argon2.IDKey(
		password,
		salt,
		4,       // iterations (OWASP minimum)
		64*1024, // memory KB
		4,       // parallelism
		32,      // key length
	)
}
