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
	"database/sql"
	"encoding/json"
	"fmt"
	"gaia/kernel/pkg/types"
	"os"
	"path/filepath"

	// modernc.org/sqlite is a pure-Go implementation of SQLite, eliminating CGO
	// dependency issues and ensuring easy cross-compilation of the GAIA kernel.
	_ "modernc.org/sqlite"
)

// SQLiteAgentStore is the Phase 3 implementation for Tier 4 (Managed Agent State).
// It strictly enforces multi-tenancy at the query level by partitioning all
// operations with the agent_id.
type SQLiteAgentStore struct {
	db *sql.DB
}

// NewSQLiteAgentStore initializes the database connection and ensures the schema exists.
func NewSQLiteAgentStore(dbPath string) (*SQLiteAgentStore, error) {
	// Ensure directory exists so SQLite doesn't panic on file creation
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("state: failed to create db directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("state: failed to open sqlite db: %w", err)
	}

	// 1. Create the strongly isolated multi-tenant table
	query := `
	CREATE TABLE IF NOT EXISTS agent_state (
		agent_id TEXT NOT NULL,
		state_key TEXT NOT NULL,
		state_data JSON NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (agent_id, state_key)
	);`
	if _, err := db.Exec(query); err != nil {
		return nil, fmt.Errorf("state: failed to create agent_state table: %w", err)
	}

	return &SQLiteAgentStore{db: db}, nil
}

// Put saves a key-value document into the SQLite store.
// It enforces strict multi-tenancy and validates storage quotas (max_bytes)
// before allowing the UPSERT to proceed.
func (s *SQLiteAgentStore) Put(agentID string, key string, data interface{}, manifest *types.AgentManifest) error {
	// 1. Verify the agent explicitly requested state storage during Handshake
	if manifest.StateRequirements == nil || !manifest.StateRequirements.Required {
		return fmt.Errorf("state: %w: agent did not request state_requirements", fmt.Errorf(string(types.ErrorCodePolicyDenied)))
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("state: failed to marshal data: %w", err)
	}

	// 2. Quota Enforcement
	// If the agent requested a quota, we must ensure they haven't exceeded it.
	if manifest.StateRequirements.MaxBytes > 0 {
		var currentSize int
		err := s.db.QueryRow(`SELECT COALESCE(SUM(LENGTH(state_data)), 0) FROM agent_state WHERE agent_id = ?`, agentID).Scan(&currentSize)
		if err != nil {
			return fmt.Errorf("state: failed to check quota: %w", err)
		}
		
		// Note: This is an approximation. If the key already exists, the UPSERT will overwrite it,
		// meaning the total size won't increase by len(bytes), but doing a strict additive check
		// prevents catastrophic allocation attempts.
		if currentSize+len(bytes) > manifest.StateRequirements.MaxBytes {
			return fmt.Errorf("state: %w: quota exceeded", fmt.Errorf(string(types.ErrorCodePolicyDenied)))
		}
	}

	// 3. Perform the UPSERT
	query := `
	INSERT INTO agent_state (agent_id, state_key, state_data, updated_at) 
	VALUES (?, ?, ?, CURRENT_TIMESTAMP)
	ON CONFLICT(agent_id, state_key) DO UPDATE SET 
		state_data=excluded.state_data, 
		updated_at=CURRENT_TIMESTAMP;`
	
	_, err = s.db.Exec(query, agentID, key, string(bytes))
	return err
}

// Get retrieves a JSON document strictly from the agent's partitioned namespace.
// It returns nil, nil if the document does not exist, as per the interface contract.
func (s *SQLiteAgentStore) Get(agentID string, key string) (interface{}, error) {
	var dataStr string
	err := s.db.QueryRow(`SELECT state_data FROM agent_state WHERE agent_id = ? AND state_key = ?`, agentID, key).Scan(&dataStr)
	
	if err == sql.ErrNoRows {
		return nil, nil // Return nil if key doesn't exist
	}
	if err != nil {
		return nil, fmt.Errorf("state: failed to retrieve data: %w", err)
	}

	var data interface{}
	if err := json.Unmarshal([]byte(dataStr), &data); err != nil {
		return nil, fmt.Errorf("state: failed to unmarshal retrieved data: %w", err)
	}
	return data, nil
}

// Delete removes a specific key from an agent's namespace.
func (s *SQLiteAgentStore) Delete(agentID string, key string) error {
	_, err := s.db.Exec(`DELETE FROM agent_state WHERE agent_id = ? AND state_key = ?`, agentID, key)
	if err != nil {
		return fmt.Errorf("state: failed to delete key: %w", err)
	}
	return nil
}

// DeleteNamespace acts as the "Kill Switch" to instantly purge all data for an agent.
func (s *SQLiteAgentStore) DeleteNamespace(agentID string) error {
	_, err := s.db.Exec(`DELETE FROM agent_state WHERE agent_id = ?`, agentID)
	if err != nil {
		return fmt.Errorf("state: failed to drop namespace: %w", err)
	}
	return nil
}
