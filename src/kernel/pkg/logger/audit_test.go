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
	"bytes"
	"os"
	"testing"
)

func TestAuditIntegrity(t *testing.T) {
	tempFile := "test_audit.log"
	secret := []byte("super-secret-key")
	defer os.Remove(tempFile)

	// 1. Initialize and Log
	al, err := InitAuditLogger(tempFile, secret)
	if err != nil {
		t.Fatalf("Failed to init: %v", err)
	}

	err = al.Log("user-1", "LOGIN", "system", nil)
	if err != nil {
		t.Fatalf("Failed to log: %v", err)
	}

	err = al.Log("user-1", "DELETE", "resource-123", map[string]interface{}{"reason": "cleanup"})
	if err != nil {
		t.Fatalf("Failed to log second entry: %v", err)
	}

	// 2. Verify Valid Chain
	if err := al.VerifyChain(); err != nil {
		t.Errorf("Valid chain verification failed: %v", err)
	}

	// 3. Tamper with file (Specifically change a signed field)
	input, _ := os.ReadFile(tempFile)
	// Replace "user-1" with "hacker" in the raw bytes
	tampered := bytes.Replace(input, []byte("user-1"), []byte("hacker"), 1)
	os.WriteFile(tempFile, tampered, 0644)

	// 4. Verify Invalid Chain
	if err := al.VerifyChain(); err == nil {
		t.Error("Tampered chain was verified as valid!")
	} else {
		t.Logf("Successfully detected tampering: %v", err)
	}
}
