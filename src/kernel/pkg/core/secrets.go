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

package core

import (
	"errors"
	"fmt"
	"os"
)

var (
	// ErrSecretNotFound is returned when a requested secret is not found in any provider.
	ErrSecretNotFound = errors.New("secret not found")
)

// SecretProvider defines the interface for fetching sensitive configuration data.
// This allows the Kernel to be decoupled from specific secret stores (Env, Vault, etc.).
type SecretProvider interface {
	// GetSecret retrieves a secret by name. Returns ErrSecretNotFound if the secret is missing.
	GetSecret(name string) (string, error)
}

// EnvSecretProvider is the default implementation that retrieves secrets from
// environment variables. This is suitable for local development and simple deployments.
type EnvSecretProvider struct {
	prefix string
}

// NewEnvSecretProvider creates a new provider that looks for environment variables.
// An optional prefix (e.g. "GAIA_") can be provided.
func NewEnvSecretProvider(prefix string) *EnvSecretProvider {
	return &EnvSecretProvider{prefix: prefix}
}

// GetSecret implements the SecretProvider interface.
func (p *EnvSecretProvider) GetSecret(name string) (string, error) {
	key := p.prefix + name
	val := os.Getenv(key)
	if val == "" {
		return "", fmt.Errorf("%w: %s", ErrSecretNotFound, key)
	}
	return val, nil
}

// SecretRegistry manages multiple secret providers and provides fallback logic.
type SecretRegistry struct {
	providers []SecretProvider
}

// NewSecretRegistry initializes a registry with the default EnvSecretProvider.
func NewSecretRegistry() *SecretRegistry {
	return &SecretRegistry{
		providers: []SecretProvider{
			NewEnvSecretProvider("GAIA_"),
		},
	}
}

// AddProvider adds a new provider to the end of the search chain.
func (r *SecretRegistry) AddProvider(p SecretProvider) {
	r.providers = append(r.providers, p)
}

// GetSecret searches through all registered providers in order.
func (r *SecretRegistry) GetSecret(name string) (string, error) {
	for _, p := range r.providers {
		secret, err := p.GetSecret(name)
		if err == nil {
			return secret, nil
		}
	}
	return "", fmt.Errorf("%w: %s", ErrSecretNotFound, name)
}
