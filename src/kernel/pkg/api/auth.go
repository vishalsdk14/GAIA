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
	"errors"
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

var (
	// ErrInvalidToken is returned when the JWT cannot be parsed or verified.
	ErrInvalidToken = errors.New("unauthorized: invalid or expired token")

	// ErrMissingToken is returned when the Authorization header is missing or malformed.
	ErrMissingToken = errors.New("unauthorized: missing or malformed Authorization header")
)

// GAIAClaims defines the standard claims expected in an agent's JWT.
// The 'sub' claim is mapped to the AgentID.
type GAIAClaims struct {
	AgentID string `json:"sub"`
	jwt.RegisteredClaims
}

// VerifyToken parses and validates a JWT string against the provided secret.
// It returns the verified AgentID (from the 'sub' claim) on success.
func VerifyToken(tokenStr string, secret []byte) (string, error) {
	// 1. Remove 'Bearer ' prefix if present
	tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")
	tokenStr = strings.TrimSpace(tokenStr)

	if tokenStr == "" {
		return "", ErrMissingToken
	}

	// 2. Parse and verify signature
	token, err := jwt.ParseWithClaims(tokenStr, &GAIAClaims{}, func(t *jwt.Token) (interface{}, error) {
		// Ensure the signing method is HMAC (HS256)
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return secret, nil
	})

	if err != nil || !token.Valid {
		return "", ErrInvalidToken
	}

	// 3. Extract identity
	claims, ok := token.Claims.(*GAIAClaims)
	if !ok || claims.AgentID == "" {
		return "", ErrInvalidToken
	}

	return claims.AgentID, nil
}
