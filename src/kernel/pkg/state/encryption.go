// Copyright 2026 GAIA Contributors
//
// Licensed under the MIT License.
// You may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://opensource.org/licenses/MIT
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package state

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
)

const (
	// KeySizeAES256 is the required length of the master encryption key in bytes (32 bytes = 256 bits).
	KeySizeAES256 = 32

	// NonceSizeGCM is the standard size for the AES-GCM initialization vector.
	NonceSizeGCM = 12
)

var (
	// ErrInvalidKeySize is returned when the provided key does not match KeySizeAES256.
	ErrInvalidKeySize = errors.New("invalid encryption key size: must be 32 bytes")

	// ErrDecryptionFailed is returned when the data cannot be decrypted or authenticated.
	ErrDecryptionFailed = errors.New("decryption failed: ciphertext may be corrupted or key is incorrect")
)

// Encryptor handles the encryption and decryption of sensitive data using AES-GCM.
type Encryptor struct {
	key []byte
}

// Key returns the underlying master key bytes.
func (e *Encryptor) Key() []byte {
	return e.key
}

// NewEncryptor initializes a new Encryptor with the provided master key.
// The key must be exactly 32 bytes long.
func NewEncryptor(key []byte) (*Encryptor, error) {
	if len(key) != KeySizeAES256 {
		return nil, ErrInvalidKeySize
	}
	return &Encryptor{key: key}, nil
}

// NewEncryptorFromHex initializes a new Encryptor from a hex-encoded string.
func NewEncryptorFromHex(hexKey string) (*Encryptor, error) {
	key, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, err
	}
	return NewEncryptor(key)
}

// Encrypt takes a plaintext byte slice and returns the AES-GCM ciphertext.
// The returned slice includes the nonce at the beginning.
func (e *Encryptor) Encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Seal appends the ciphertext to the nonce
	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// Decrypt takes a ciphertext byte slice (with nonce prefix) and returns the plaintext.
func (e *Encryptor) Decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, ErrDecryptionFailed
	}

	nonce, actualCiphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, actualCiphertext, nil)
	if err != nil {
		return nil, ErrDecryptionFailed
	}

	return plaintext, nil
}
