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
	"fmt"
	"sync"
)

// QuotaManager tracks resource consumption per task to prevent kernel exhaustion.
type QuotaManager struct {
	mu           sync.Mutex
	maxTasks     int
	activeTasks  int
	maxSteps     int
	taskStepCount map[string]int
}

func NewQuotaManager(maxTasks, maxStepsPerTask int) *QuotaManager {
	return &QuotaManager{
		maxTasks:      maxTasks,
		maxSteps:      maxStepsPerTask,
		taskStepCount: make(map[string]int),
	}
}

func (q *QuotaManager) AcquireTask() error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.activeTasks >= q.maxTasks {
		return fmt.Errorf("quota: kernel task limit reached (%d)", q.maxTasks)
	}
	q.activeTasks++
	return nil
}

func (q *QuotaManager) ReleaseTask(taskID string) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.activeTasks > 0 {
		q.activeTasks--
	}
	delete(q.taskStepCount, taskID)
}

func (q *QuotaManager) IncrementStep(taskID string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.taskStepCount[taskID]++
	if q.taskStepCount[taskID] > q.maxSteps {
		return fmt.Errorf("quota: task step limit exceeded (%d)", q.maxSteps)
	}
	return nil
}
