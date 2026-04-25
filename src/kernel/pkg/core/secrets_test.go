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
	"os"
	"testing"
)

func TestSecretRegistry(t *testing.T) {
	// 1. Setup Env
	os.Setenv("GAIA_TEST_SECRET", "super-secret-value")
	defer os.Unsetenv("GAIA_TEST_SECRET")

	registry := NewSecretRegistry()

	// 2. Test Success
	val, err := registry.GetSecret("TEST_SECRET")
	if err != nil {
		t.Fatalf("failed to get secret: %v", err)
	}
	if val != "super-secret-value" {
		t.Errorf("expected super-secret-value, got %s", val)
	}

	// 3. Test Missing
	_, err = registry.GetSecret("NON_EXISTENT")
	if err == nil {
		t.Error("expected error for non-existent secret, got nil")
	}
}

type MockProvider struct {
	val string
}

func (m *MockProvider) GetSecret(name string) (string, error) {
	if name == "MOCK_KEY" {
		return m.val, nil
	}
	return "", ErrSecretNotFound
}

func TestSecretRegistryFallback(t *testing.T) {
	registry := NewSecretRegistry()
	registry.AddProvider(&MockProvider{val: "mock-value"})

	// 1. Test Fallback to Mock
	val, err := registry.GetSecret("MOCK_KEY")
	if err != nil {
		t.Fatalf("failed to get secret from mock: %v", err)
	}
	if val != "mock-value" {
		t.Errorf("expected mock-value, got %s", val)
	}
}
