// Copyright 2026 GAIA Contributors
//
// Licensed under the MIT License.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package main implements a utility to export GAIA core types as JSON Schemas.
// This ensures that all SDKs (TypeScript, Python) stay in lock-step with the
// Kernel's data model.
package main

import (
	"encoding/json"
	"fmt"
	"gaia/kernel/pkg/types"
	"os"
	"path/filepath"

	"github.com/invopop/jsonschema"
)

func main() {
	schemasDir := "../../docs/specs/schemas"
	
	// List of types to export
	typeMap := map[string]interface{}{
		"agent-manifest.json":   types.AgentManifest{},
		"task.json":             types.Task{},
		"step.json":             types.Step{},
		"request.json":          types.Request{},
		"response.json":         types.Response{},
		"event.json":            types.Event{},
		"async-completion.json": types.AsyncCompletion{},
	}

	for filename, t := range typeMap {
		schema := jsonschema.Reflect(t)
		data, err := json.MarshalIndent(schema, "", "  ")
		if err != nil {
			fmt.Printf("Error reflecting %s: %v\n", filename, err)
			continue
		}

		path := filepath.Join(schemasDir, filename)
		if err := os.WriteFile(path, data, 0644); err != nil {
			fmt.Printf("Error writing %s: %v\n", filename, err)
			continue
		}
		fmt.Printf("✅ Exported %s\n", filename)
	}
}
