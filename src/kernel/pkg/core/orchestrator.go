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
	"gaia/kernel/pkg/logger"
	"gaia/kernel/pkg/registry"
	"gaia/kernel/pkg/state"
	"gaia/kernel/pkg/types"
	"sync"
	"time"
)

// Orchestrator is the top-level manager that handles multi-tenant task coordination.
// it implements the "Goal Manager" role from the GAIA architecture.
type Orchestrator struct {
	mu           sync.RWMutex
	config       *KernelConfig
	registry     registry.CapabilityRegistry
	taskStore    *state.TaskStore
	activeTasks  map[string]*Coordinator
	planner      Planner
	transport    AgentTransport
	quota        *QuotaManager
}

// NewOrchestrator initializes the kernel's central task management hub.
func NewOrchestrator(c *KernelConfig, r registry.CapabilityRegistry, ts *state.TaskStore, p Planner, trans AgentTransport) *Orchestrator {
	return &Orchestrator{
		config:      c,
		registry:    r,
		taskStore:   ts,
		activeTasks: make(map[string]*Coordinator),
		planner:     p,
		transport:   trans,
		quota:       NewQuotaManager(100, 50),
	}
}

// Config returns the current kernel configuration.
func (o *Orchestrator) Config() *KernelConfig {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.config
}

// SubmitTask initializes a new task, persists it, and starts the coordination loop.
func (o *Orchestrator) SubmitTask(goal string) (*types.Task, error) {
	// Phase 11: Check Task Quota
	if err := o.quota.AcquireTask(); err != nil {
		return nil, err
	}

	o.mu.Lock()
	defer o.mu.Unlock()

	if len(o.activeTasks) >= o.config.MaxConcurrentTasks {
		return nil, fmt.Errorf("kernel: %w: max_concurrent_tasks reached", fmt.Errorf(string(types.ErrorCodeInternalError)))
	}

	taskID := fmt.Sprintf("task-%d", time.Now().UnixNano())
	task := &types.Task{
		TaskID:    taskID,
		Goal:      goal,
		Status:    types.TaskStatusPending,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	// 1. Persist to Tier 2 (Stateful Re-entry)
	if err := o.taskStore.SaveTask(task); err != nil {
		return nil, fmt.Errorf("orchestrator: failed to persist task: %w", err)
	}

	// 2. Initialize Coordinator
	coord := NewCoordinator(task, o.config, o.registry, o.planner, o.transport, o.taskStore, o.quota)
	o.activeTasks[taskID] = coord

	// 3. Start Control Loop in background
	go func() {
		defer func() {
			o.mu.Lock()
			delete(o.activeTasks, taskID)
			o.mu.Unlock()
		}()

		if err := coord.Run(); err != nil {
			logger.L.Error("Task coordination failed", "task_id", taskID, "error", err)
		}
	}()

	return task, nil
}

// GetTask retrieves a task from memory or storage.
func (o *Orchestrator) GetTask(taskID string) (*types.Task, error) {
	o.mu.RLock()
	if coord, ok := o.activeTasks[taskID]; ok {
		o.mu.RUnlock()
		return coord.task, nil
	}
	o.mu.RUnlock()

	return o.taskStore.GetTask(taskID)
}

// ApproveStep proxies the approval request to the active coordinator for the given task.
func (o *Orchestrator) ApproveStep(taskID, stepID string) error {
	o.mu.RLock()
	coord, ok := o.activeTasks[taskID]
	o.mu.RUnlock()

	if !ok {
		return fmt.Errorf("orchestrator: task %s is not currently active", taskID)
	}

	return coord.ApproveStep(stepID)
}

// UpdatePlan proxies the plan modification request to the active coordinator.
func (o *Orchestrator) UpdatePlan(taskID string, newPlan []types.Step) error {
	o.mu.RLock()
	coord, ok := o.activeTasks[taskID]
	o.mu.RUnlock()

	if !ok {
		return fmt.Errorf("orchestrator: task %s is not currently active", taskID)
	}

	return coord.UpdatePlan(newPlan)
}
