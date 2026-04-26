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

// Package state implements the multi-tiered state management architecture for the GAIA kernel.
// This file specifically implements the persistent SQLite storage for Tier 4 (Managed Agent State),
// ensuring strict multi-tenant isolation and quota enforcement.
package state

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"gaia/kernel/pkg/types"
	"os"
	"path/filepath"

	// modernc.org/sqlite is a pure-Go implementation of SQLite, eliminating CGO
	// dependency issues and ensuring easy cross-compilation of the GAIA kernel.
	_ "modernc.org/sqlite"
)

// SQLiteStore is the Phase 3+ implementation for persistent kernel state.
// It handles Tier 4 (Managed Agent State) and Tier 2 (Task History).
type SQLiteStore struct {
	DB        *sql.DB
	encryptor *Encryptor
}

const (
	// EncryptionPrefix is prepended to encrypted blobs in the database to distinguish
	// them from legacy plaintext JSON documents.
	EncryptionPrefix = "GAIA:ENC:"
)

// NewSQLiteStore initializes the database connection and ensures core schemas exist.
func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	// Ensure directory exists so SQLite doesn't panic on file creation
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("state: failed to create db directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("state: failed to open sqlite db: %w", err)
	}

	// 1. Create the strongly isolated multi-tenant table for Agent State
	query := `
	CREATE TABLE IF NOT EXISTS agent_state (
		agent_id TEXT NOT NULL,
		state_key TEXT NOT NULL,
		state_data JSON NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (agent_id, state_key)
	);
	
	CREATE TABLE IF NOT EXISTS agent_state_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		agent_id TEXT NOT NULL,
		state_key TEXT NOT NULL,
		state_data JSON NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	if _, err := db.Exec(query); err != nil {
		return nil, fmt.Errorf("state: failed to create tables: %w", err)
	}

	return &SQLiteStore{DB: db}, nil
}

// EnableEncryption initializes the internal encryptor with a 32-byte master key.
func (s *SQLiteStore) EnableEncryption(key []byte) error {
	enc, err := NewEncryptor(key)
	if err != nil {
		return err
	}
	s.encryptor = enc
	return nil
}

// Put saves a key-value document into the SQLite store.
func (s *SQLiteStore) Put(agentID string, key string, data interface{}, manifest *types.AgentManifest) error {
	if manifest.StateRequirements == nil || !manifest.StateRequirements.Required {
		return fmt.Errorf("state: %w: agent did not request state_requirements", fmt.Errorf(string(types.ErrorCodePolicyDenied)))
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("state: failed to marshal data: %w", err)
	}

	if manifest.StateRequirements.MaxBytes > 0 {
		var currentSize int
		err := s.DB.QueryRow(`SELECT COALESCE(SUM(LENGTH(state_data)), 0) FROM agent_state WHERE agent_id = ?`, agentID).Scan(&currentSize)
		if err != nil {
			return fmt.Errorf("state: failed to check quota: %w", err)
		}
		
		if currentSize+len(bytes) > manifest.StateRequirements.MaxBytes {
			return fmt.Errorf("state: %w: quota exceeded", fmt.Errorf(string(types.ErrorCodePolicyDenied)))
		}
	}

	query := `
	INSERT INTO agent_state (agent_id, state_key, state_data, updated_at) 
	VALUES (?, ?, ?, CURRENT_TIMESTAMP)
	ON CONFLICT(agent_id, state_key) DO UPDATE SET 
		state_data=excluded.state_data, 
		updated_at=CURRENT_TIMESTAMP;`
	
	finalData := string(bytes)
	if s.encryptor != nil {
		encrypted, err := s.encryptor.Encrypt(bytes)
		if err != nil {
			return fmt.Errorf("state: encryption failed: %w", err)
		}
		// Prepend prefix to identify encrypted data
		finalData = EncryptionPrefix + string(encrypted)
	}

	_, err = s.DB.Exec(query, agentID, key, finalData)
	if err != nil {
		return err
	}

	// Phase 10: Record history for restoration support
	_, err = s.DB.Exec(`INSERT INTO agent_state_history (agent_id, state_key, state_data) VALUES (?, ?, ?)`, agentID, key, finalData)
	return err
}

// Get retrieves a JSON document strictly from the agent's partitioned namespace.
func (s *SQLiteStore) Get(agentID string, key string) (interface{}, error) {
	var dataStr string
	err := s.DB.QueryRow(`SELECT state_data FROM agent_state WHERE agent_id = ? AND state_key = ?`, agentID, key).Scan(&dataStr)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("state: failed to retrieve data: %w", err)
	}

	var rawBytes []byte
	// Check if data is encrypted by looking for the GAIA:ENC: prefix
	if len(dataStr) > len(EncryptionPrefix) && dataStr[:len(EncryptionPrefix)] == EncryptionPrefix {
		if s.encryptor == nil {
			return nil, errors.New("state: data is encrypted but no master key is provided")
		}
		
		encryptedData := []byte(dataStr[len(EncryptionPrefix):])
		decrypted, err := s.encryptor.Decrypt(encryptedData)
		if err != nil {
			return nil, fmt.Errorf("state: %w", err)
		}
		rawBytes = decrypted
	} else {
		rawBytes = []byte(dataStr)
	}

	var data interface{}
	if err := json.Unmarshal(rawBytes, &data); err != nil {
		return nil, fmt.Errorf("state: failed to unmarshal retrieved data: %w", err)
	}
	return data, nil
}

// Delete removes a specific key from an agent's namespace.
// Phase 11: Records a tombstone in history for reliable restoration.
func (s *SQLiteStore) Delete(agentID string, key string) error {
	_, err := s.DB.Exec(`DELETE FROM agent_state WHERE agent_id = ? AND state_key = ?`, agentID, key)
	if err != nil {
		return fmt.Errorf("state: failed to delete key: %w", err)
	}

	// Record Tombstone (state_data = NULL)
	_, err = s.DB.Exec(`INSERT INTO agent_state_history (agent_id, state_key, state_data) VALUES (?, ?, NULL)`, agentID, key)
	return err
}

// DeleteNamespace acts as the "Kill Switch" to instantly purge all data for an agent.
func (s *SQLiteStore) DeleteNamespace(agentID string) error {
	// For simplicity in Phase 11, we don't tombstone every key in a namespace drop,
	// but in a production system, this would be a single high-level event.
	_, err := s.DB.Exec(`DELETE FROM agent_state WHERE agent_id = ?`, agentID)
	if err != nil {
		return fmt.Errorf("state: failed to drop namespace: %w", err)
	}
	return nil
}

// ListKeys returns all keys stored by an agent.
func (s *SQLiteStore) ListKeys(agentID string, limit, offset int) ([]string, error) {
	rows, err := s.DB.Query(`SELECT state_key FROM agent_state WHERE agent_id = ? ORDER BY state_key LIMIT ? OFFSET ?`, agentID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("state: failed to list keys: %w", err)
	}
	defer rows.Close()

	var keys []string
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	return keys, nil
}

// GetUsage returns the current byte size of an agent's storage.
func (s *SQLiteStore) GetUsage(agentID string) (int, error) {
	var currentSize int
	err := s.DB.QueryRow(`SELECT COALESCE(SUM(LENGTH(state_data)), 0) FROM agent_state WHERE agent_id = ?`, agentID).Scan(&currentSize)
	if err != nil {
		return 0, fmt.Errorf("state: failed to check usage: %w", err)
	}
	return currentSize, nil
}

/**
 * RestoreAgentState rolls back an agent's namespace to the latest state available before the targetTime.
 * This is an administrative operation that ensures business continuity.
 */
func (s *SQLiteStore) RestoreAgentState(agentID string, targetTime string) error {
	// 1. Delete current state for the agent
	if err := s.DeleteNamespace(agentID); err != nil {
		return err
	}

	// 2. Reconstruct state from history: find the latest entry for each key <= targetTime
	// Note: We skip entries where state_data is NULL (Tombstones)
	query := `
	INSERT INTO agent_state (agent_id, state_key, state_data, updated_at)
	SELECT h.agent_id, h.state_key, h.state_data, h.created_at
	FROM agent_state_history h
	WHERE h.agent_id = ? AND h.created_at <= ?
	AND h.state_data IS NOT NULL
	AND h.id IN (
		SELECT MAX(id) FROM agent_state_history 
		WHERE agent_id = ? AND created_at <= ?
		GROUP BY state_key
	);`

	_, err := s.DB.Exec(query, agentID, targetTime, agentID, targetTime)
	return err
}
