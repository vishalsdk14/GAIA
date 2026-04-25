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
	"gaia/kernel/pkg/registry"
	"gaia/kernel/pkg/state"
	"gaia/kernel/pkg/types"
	"sync"
	"time"
)

// Coordinator manages the lifecycle of a single Task.
// It executes the 10-phase control loop in a dedicated goroutine, ensuring that
// state transitions are atomic and that invariants from docs/specs/control-loop.md are held.
type Coordinator struct {
	mu       sync.Mutex
	config   *KernelConfig
	task     *types.Task
	stateMgr *state.ActiveStateManager
	registry registry.CapabilityRegistry
	planner  Planner
}

// NewCoordinator initializes a new task coordinator with the required kernel subsystems.
func NewCoordinator(t *types.Task, c *KernelConfig, r registry.CapabilityRegistry, p Planner) *Coordinator {
	return &Coordinator{
		task:     t,
		config:   c,
		stateMgr: state.NewActiveStateManager(t.TaskID),
		registry: r,
		planner:  p,
	}
}

// Run executes the 10-phase control loop for the task.
// This function implements the logic defined in docs/specs/control-loop.md Section 30.
func (c *Coordinator) Run() error {
	for {
		// Kernel Invariant 1: Progress Guarantee.
		// Check for termination states before each iteration.
		if c.isTerminal() {
			return nil
		}

		// Phase 1: Loop Entry (Submission)
		if err := c.phase1Submission(); err != nil {
			return c.failTask(err)
		}

		// Phase 2: Planning (Reasoning)
		if err := c.phase2Planning(); err != nil {
			return c.failTask(err)
		}

		// Skeleton Placeholder for Phases 3-10
		// In Phase 3 (Runtime), these will be implemented to handle DAG resolution,
		// parallel dispatch, policy checks, and results processing.
		fmt.Printf("Coordinator [task=%s]: Skeleton loop iteration complete. Yielding.\n", c.task.TaskID)
		time.Sleep(1 * time.Second) // Yield to prevent busy loop in skeleton phase
		
		// For foundation phase, we break after one iteration to prevent infinite loop
		// since we haven't implemented the termination logic for a terminal state yet.
		break 
	}
	return nil
}

// phase1Submission implements Phase 1 of the control loop.
// It validates the pre-conditions and transitions the task from 'pending' to 'planning'.
func (c *Coordinator) phase1Submission() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.task.Status != types.TaskStatusPending {
		return nil // Already past submission
	}

	// Validate pre-conditions (Control Loop Spec 1.2)
	if c.task.Goal == "" {
		return fmt.Errorf("core: task goal cannot be empty")
	}

	// Transition Task: pending -> planning
	c.task.Status = types.TaskStatusPlanning
	c.task.UpdatedAt = time.Now().UTC()
	
	// TODO: Emit TASK_PLANNING event
	return nil
}

// phase2Planning implements Phase 2 of the control loop.
// It invokes the probabilistic planner and validates the returned plan DAG.
func (c *Coordinator) phase2Planning() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.task.Status != types.TaskStatusPlanning {
		return nil
	}

	// 1. Assemble planner input (State Snapshot + Capabilities)
	activeState := c.stateMgr.GetSnapshot()
	
	// 2. Invoke the Planner Adapter (Local or Cloud)
	// Note: Capabilities list would be fetched from registry in a real implementation
	plan, err := c.planner.GeneratePlan(c.task.Goal, activeState, nil)
	if err != nil {
		// In this foundation phase, the planners return errors since network is not implemented.
		// We catch it here to satisfy the skeleton flow.
		fmt.Printf("Coordinator [task=%s]: Planner invocation skipped (foundation phase)\n", c.task.TaskID)
		return nil 
	}

	// 3. Post-Planning Transition
	c.task.Status = types.TaskStatusExecuting
	c.task.UpdatedAt = time.Now().UTC()
	c.task.Plan = plan
	
	// TODO: Emit PLAN_GENERATED & TASK_EXECUTING events
	return nil
}

// isTerminal checks if the task has reached a final state (completed, failed, or cancelled).
func (c *Coordinator) isTerminal() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.task.Status == types.TaskStatusCompleted || 
		   c.task.Status == types.TaskStatusFailed || 
		   c.task.Status == types.TaskStatusCancelled
}

// failTask is a helper to transition a task to the 'failed' state and record the error.
func (c *Coordinator) failTask(err error) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.task.Status = types.TaskStatusFailed
	c.task.UpdatedAt = time.Now().UTC()
	// TODO: Record error in task and emit TASK_FAILED event
	
	return err
}
