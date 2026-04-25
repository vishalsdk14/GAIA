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

package common

import (
	"gaia/kernel/pkg/types"
	"sync"
	"time"
)

// Event represents a system-wide telemetry event.
type Event struct {
	Timestamp time.Time              `json:"timestamp"`
	Type      types.EventName        `json:"type"`
	TaskID    string                 `json:"task_id,omitempty"`
	StepID    string                 `json:"step_id,omitempty"`
	Payload   map[string]interface{} `json:"payload,omitempty"`
}

// EventBus is a simple, thread-safe Pub/Sub bus for kernel telemetry.
type EventBus struct {
	mu          sync.RWMutex
	subscribers []chan Event
}

var globalBus *EventBus
var once sync.Once

// GetEventBus returns the singleton EventBus instance.
func GetEventBus() *EventBus {
	once.Do(func() {
		globalBus = &EventBus{
			subscribers: make([]chan Event, 0),
		}
	})
	return globalBus
}

// Subscribe adds a new subscriber channel to the bus.
func (b *EventBus) Subscribe() chan Event {
	b.mu.Lock()
	defer b.mu.Unlock()
	ch := make(chan Event, 100)
	b.subscribers = append(b.subscribers, ch)
	return ch
}

// Emit broadcasts an event to all subscribers.
func (b *EventBus) Emit(e Event) {
	if e.Timestamp.IsZero() {
		e.Timestamp = time.Now().UTC()
	}

	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, ch := range b.subscribers {
		// Non-blocking send to prevent slow subscribers from hanging the kernel
		select {
		case ch <- e:
		default:
		}
	}
}
