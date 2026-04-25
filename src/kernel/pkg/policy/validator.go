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

package policy

import (
	"fmt"
	"gaia/kernel/pkg/types"

	"github.com/xeipuuv/gojsonschema"
)

// SchemaValidator enforces structural contracts between the Kernel and Agents.
type SchemaValidator struct{}

// ValidateStepInput ensures the input data matches the capability's declared JSON schema.
func (v *SchemaValidator) ValidateStepInput(input interface{}, schema map[string]interface{}) error {
	if schema == nil {
		return nil // No schema defined, skip validation
	}

	schemaLoader := gojsonschema.NewGoLoader(schema)
	documentLoader := gojsonschema.NewGoLoader(input)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return fmt.Errorf("policy: schema validation engine error: %w", err)
	}

	if !result.Valid() {
		var errMsg string
		for _, desc := range result.Errors() {
			errMsg += fmt.Sprintf("- %s\n", desc)
		}
		return fmt.Errorf("policy: %w: %s", fmt.Errorf(string(types.ErrorCodeSchemaViolation)), errMsg)
	}

	return nil
}
