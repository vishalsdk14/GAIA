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
	"bytes"
	"crypto/rand"
	"gaia/kernel/pkg/types"
	"os"
	"testing"
)

// TestEncryptor verifies that the AES-GCM encryption and decryption are reversible.
func TestEncryptor(t *testing.T) {
	key := make([]byte, KeySizeAES256)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("failed to generate random key: %v", err)
	}

	enc, err := NewEncryptor(key)
	if err != nil {
		t.Fatalf("failed to create encryptor: %v", err)
	}

	plaintext := []byte("hello gaia security")
	ciphertext, err := enc.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	if bytes.Equal(plaintext, ciphertext) {
		t.Error("ciphertext matches plaintext; no encryption occurred")
	}

	decrypted, err := enc.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("decryption failed: %v", err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Errorf("decrypted data mismatch: expected %s, got %s", plaintext, decrypted)
	}
}

// TestSQLiteEncryption verifies that the SQLite store correctly encrypts data when enabled.
func TestSQLiteEncryption(t *testing.T) {
	dbPath := "./test_encrypted.db"
	defer os.Remove(dbPath)

	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	key := make([]byte, KeySizeAES256)
	copy(key, "32-byte-long-master-key-for-test")
	if err := store.EnableEncryption(key); err != nil {
		t.Fatalf("failed to enable encryption: %v", err)
	}

	agentID := "test-agent"
	manifest := &types.AgentManifest{
		StateRequirements: &types.StateRequirements{
			Required: true,
			MaxBytes: 1024,
		},
	}

	testData := map[string]string{"secret": "value"}
	if err := store.Put(agentID, "key1", testData, manifest); err != nil {
		t.Fatalf("failed to put data: %v", err)
	}

	// 1. Verify it's encrypted in the database
	var dataStr string
	err = store.DB.QueryRow("SELECT state_data FROM agent_state WHERE agent_id = ? AND state_key = ?", agentID, "key1").Scan(&dataStr)
	if err != nil {
		t.Fatalf("failed to query raw data: %v", err)
	}

	if len(dataStr) <= len(EncryptionPrefix) || dataStr[:len(EncryptionPrefix)] != EncryptionPrefix {
		t.Errorf("data in DB is not prefixed with %s: %s", EncryptionPrefix, dataStr)
	}

	// 2. Verify it's decrypted correctly
	retrieved, err := store.Get(agentID, "key1")
	if err != nil {
		t.Fatalf("failed to get data: %v", err)
	}

	m, ok := retrieved.(map[string]interface{})
	if !ok || m["secret"] != "value" {
		t.Errorf("retrieved data mismatch or incorrect type: %v", retrieved)
	}
}
