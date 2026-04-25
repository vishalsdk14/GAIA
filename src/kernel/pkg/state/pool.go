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

// Package state implements the multi-tiered state management architecture for the GAIA kernel.
// It provides high-performance, concurrency-safe access to Tier 1 (Active State) and
// strictly sandboxed namespaces for Tier 4 (Managed Agent State).
package state

import (
	"gaia/kernel/pkg/types"
	"sync"
)

// eventPool provides a zero-allocation hot path for Event objects.
var eventPool = sync.Pool{
	New: func() interface{} {
		return &types.Event{}
	},
}

// GetEvent acquires an Event object from the zero-allocation pool.
// Instead of allocating a new Event object on the heap, it retrieves a recycled
// instance, significantly reducing Garbage Collection (GC) pressure during high
// throughput execution phases.
func GetEvent() *types.Event {
	return eventPool.Get().(*types.Event)
}

// PutEvent clears an Event object and returns it to the pool for future reuse.
// This function MUST be called after an event is fully processed and dispatched.
// It aggressively clears all fields (pointers, slices, maps, and strings) to prevent
// memory leaks, ensuring that recycled objects do not hold onto stale memory references.
func PutEvent(e *types.Event) {
	// Clear all pointers and maps to prevent memory leaks during GC
	e.Type = ""
	e.Name = ""
	e.Payload = nil
	e.TaskID = ""
	e.StepID = ""
	e.SequenceNumber = 0
	e.PreviousEventID = ""
	eventPool.Put(e)
}
