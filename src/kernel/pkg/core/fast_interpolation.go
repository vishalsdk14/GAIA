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
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
)

// bufferPool reduces allocations during string interpolation.
var bufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

// FastResolveInterpolation provides a high-performance alternative to the legacy
// JSON-based interpolation. It uses a single-pass scanner and buffer pooling
// to minimize allocations and CPU overhead.
func FastResolveInterpolation(input interface{}, hotState map[string]interface{}) (interface{}, error) {
	// If input is nil, return nil immediately.
	if input == nil {
		return nil, nil
	}

	// For Phase 11, we still start with JSON marshaling to handle arbitrary structures,
	// but we optimize the replacement pass to be zero-allocation on the buffer.
	bytesInput, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("fast_interpolation: marshal failed: %w", err)
	}

	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)

	// Pre-allocate buffer capacity to avoid re-allocations
	buf.Grow(len(bytesInput) + 128)

	i := 0
	for i < len(bytesInput) {
		// Look for start tag {{
		start := bytes.Index(bytesInput[i:], []byte("{{"))
		if start == -1 {
			buf.Write(bytesInput[i:])
			break
		}
		
		// Write everything before the tag
		buf.Write(bytesInput[i : i+start])
		i += start + 2 // Skip {{

		// Look for end tag }}
		end := bytes.Index(bytesInput[i:], []byte("}}"))
		if end == -1 {
			// Malformed tag, write remaining and exit
			buf.WriteString("{{")
			buf.Write(bytesInput[i:])
			break
		}

		// Extract key and trim whitespace
		key := string(bytes.TrimSpace(bytesInput[i : i+end]))
		i += end + 2 // Skip }}

		// Resolve key (Phase 11: Support dot-notation for nested objects)
		if val, exists := GetNestedValue(hotState, key); exists {
			switch v := val.(type) {
			case string:
				// Escape the string for JSON but remove the surrounding quotes 
				// since the tag is already inside quotes in the JSON input.
				valBytes, _ := json.Marshal(v)
				if len(valBytes) >= 2 {
					buf.Write(valBytes[1 : len(valBytes)-1])
				}
			default:
				valBytes, _ := json.Marshal(v)
				buf.Write(valBytes)
			}
		} else {
			return nil, fmt.Errorf("fast_interpolation: variable {{%s}} could not be resolved in the current state", key)
		}
	}

	var resolved interface{}
	if err := json.Unmarshal(buf.Bytes(), &resolved); err != nil {
		return nil, fmt.Errorf("fast_interpolation: unmarshal failed: %w", err)
	}

	return resolved, nil
}

// GetNestedValue extracts a value from a nested map or slice using dot notation (e.g., "state.user.id" or "step.map[0].id").
func GetNestedValue(m map[string]interface{}, path string) (interface{}, bool) {
	parts := strings.Split(path, ".")
	var current interface{} = m

	for _, part := range parts {
		// Handle array indexing like "map[0]"
		if strings.Contains(part, "[") && strings.HasSuffix(part, "]") {
			bracketIdx := strings.Index(part, "[")
			arrayName := part[:bracketIdx]
			indexStr := part[bracketIdx+1 : len(part)-1]
			
			var index int
			if _, err := fmt.Sscanf(indexStr, "%d", &index); err != nil {
				return nil, false
			}

			// First, get the array from the current map
			if curMap, ok := current.(map[string]interface{}); ok {
				if val, exists := curMap[arrayName]; exists {
					current = val
				} else {
					return nil, false
				}
			} else {
				return nil, false
			}

			// Then, get the element from the array
			if curSlice, ok := current.([]interface{}); ok {
				if index >= 0 && index < len(curSlice) {
					current = curSlice[index]
				} else {
					return nil, false
				}
			} else if curSlice, ok := current.([]map[string]interface{}); ok {
				if index >= 0 && index < len(curSlice) {
					current = curSlice[index]
				} else {
					return nil, false
				}
			} else {
				return nil, false
			}
			continue
		}

		if curMap, ok := current.(map[string]interface{}); ok {
			if val, exists := curMap[part]; exists {
				current = val
			} else if part == "output" {
				continue 
			} else {
				return nil, false
			}
		} else if curSlice, ok := current.([]interface{}); ok {
			// Handle dot-notation indexing (e.g., "results.0")
			if idx, err := strconv.Atoi(part); err == nil {
				if idx >= 0 && idx < len(curSlice) {
					current = curSlice[idx]
				} else {
					return nil, false
				}
			} else {
				return nil, false
			}
		} else {
			return nil, false
		}
	}

	return current, true
}
