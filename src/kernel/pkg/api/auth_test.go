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

package api

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TestVerifyToken verifies that the JWT parsing and identity extraction work as expected.
func TestVerifyToken(t *testing.T) {
	secret := []byte("test-secret-key")
	agentID := "test-agent-123"

	// 1. Generate a valid token
	claims := &GAIAClaims{
		AgentID: agentID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(secret)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	// 2. Test valid token
	verifiedID, err := VerifyToken(tokenStr, secret)
	if err != nil {
		t.Errorf("failed to verify valid token: %v", err)
	}
	if verifiedID != agentID {
		t.Errorf("verified identity mismatch: expected %s, got %s", agentID, verifiedID)
	}

	// 3. Test token with 'Bearer ' prefix
	verifiedID, err = VerifyToken("Bearer "+tokenStr, secret)
	if err != nil {
		t.Errorf("failed to verify token with Bearer prefix: %v", err)
	}
	if verifiedID != agentID {
		t.Errorf("verified identity mismatch with prefix: expected %s, got %s", agentID, verifiedID)
	}

	// 4. Test invalid secret
	_, err = VerifyToken(tokenStr, []byte("wrong-secret"))
	if err == nil {
		t.Error("verified token with wrong secret")
	}

	// 5. Test expired token
	expiredClaims := &GAIAClaims{
		AgentID: agentID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
		},
	}
	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims)
	expiredTokenStr, _ := expiredToken.SignedString(secret)
	_, err = VerifyToken(expiredTokenStr, secret)
	if err == nil {
		t.Error("verified expired token")
	}
}
