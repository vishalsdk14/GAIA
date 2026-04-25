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

package logger

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// AuditEntry represents a single tamper-proof log entry.
type AuditEntry struct {
	LogID     string                 `json:"log_id"`
	Timestamp time.Time              `json:"timestamp"`
	Actor     string                 `json:"actor"`
	Action    string                 `json:"action"`
	Resource  string                 `json:"resource"`
	Context   map[string]interface{} `json:"context,omitempty"`
	Hash      string                 `json:"hash"`
	PrevHash  string                 `json:"prev_hash"`
}

// AuditLogger implements a tamper-proof, append-only audit log with SHA-256 chaining.
type AuditLogger struct {
	mu       sync.Mutex
	filePath string
	lastHash string
}

var globalAudit *AuditLogger
var auditOnce sync.Once

// InitAuditLogger initializes the global audit log file.
func InitAuditLogger(path string) (*AuditLogger, error) {
	auditOnce.Do(func() {
		globalAudit = &AuditLogger{
			filePath: path,
			lastHash: "0000000000000000000000000000000000000000000000000000000000000000",
		}
	})
	return globalAudit, nil
}

// Log records a new action into the tamper-proof audit trail.
func (a *AuditLogger) Log(actor, action, resource string, context map[string]interface{}) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	entry := AuditEntry{
		LogID:     fmt.Sprintf("audit-%d", time.Now().UnixNano()),
		Timestamp: time.Now().UTC(),
		Actor:     actor,
		Action:    action,
		Resource:  resource,
		Context:   context,
		PrevHash:  a.lastHash,
	}

	// Calculate Hash: SHA-256(Timestamp + Actor + Action + Resource + PrevHash)
	payload := fmt.Sprintf("%s|%s|%s|%s|%s", entry.Timestamp.Format(time.RFC3339), entry.Actor, entry.Action, entry.Resource, entry.PrevHash)
	h := sha256.New()
	h.Write([]byte(payload))
	entry.Hash = hex.EncodeToString(h.Sum(nil))

	// Update chain
	a.lastHash = entry.Hash

	// Append to file (NDJSON)
	f, err := os.OpenFile(a.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	bytes, _ := json.Marshal(entry)
	if _, err := f.Write(append(bytes, '\n')); err != nil {
		return err
	}

	return nil
}

// GetAuditLogger returns the singleton instance.
func GetAuditLogger() *AuditLogger {
	return globalAudit
}
