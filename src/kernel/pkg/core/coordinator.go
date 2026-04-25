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
	transport AgentTransport
}

// NewCoordinator initializes a new task coordinator with the required kernel subsystems.
func NewCoordinator(t *types.Task, c *KernelConfig, r registry.CapabilityRegistry, p Planner, trans AgentTransport) *Coordinator {
	return &Coordinator{
		task:     t,
		config:   c,
		stateMgr: state.NewActiveStateManager(t.TaskID),
		registry: r,
		planner:  p,
		transport: trans,
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

		// Phase 3-10: DAG Execution Engine
		if c.task.Status == types.TaskStatusExecuting {
			if err := c.executeDAG(); err != nil {
				return c.failTask(err)
			}
		}

		// Check for termination after DAG iteration
		if c.isTerminal() {
			return nil
		}

		// Prevent busy looping if waiting for async steps
		time.Sleep(100 * time.Millisecond)
	}
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
// It includes Phase 2.3 failure recovery (Correction Prompts).
func (c *Coordinator) phase2Planning() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.task.Status != types.TaskStatusPlanning {
		return nil
	}

	activeState := c.stateMgr.GetSnapshot()
	
	// Phase 2.3: Failure Recovery Loop
	var lastErr error
	var correctionPrompt string
	
	for attempt := 0; attempt < 2; attempt++ {
		goal := c.task.Goal
		if correctionPrompt != "" {
			goal = correctionPrompt
		}

		plan, err := c.planner.GeneratePlan(goal, activeState, nil)
		if err == nil {
			// Success!
			c.task.Status = types.TaskStatusExecuting
			c.task.UpdatedAt = time.Now().UTC()
			c.task.Plan = plan.Steps
			return nil
		}

		// Check if the error is a schema violation (malformed JSON)
		lastErr = err
		if types.ErrorCode(err.Error()) == types.ErrorCodeSchemaViolation || attempt == 0 {
			// In foundation phase, we simulate the correction prompt build
			correctionPrompt = BuildCorrectionPrompt("INVALID_JSON_HERE", err.Error())
			fmt.Printf("Coordinator [task=%s]: Retrying with correction prompt (attempt %d)\n", c.task.TaskID, attempt+1)
			continue
		}
		break
	}

	return fmt.Errorf("core: planner failed after retries: %w", lastErr)
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

// executeDAG manages the parallel dispatch of ready steps (Phases 3-8).
func (c *Coordinator) executeDAG() error {
	// Phase 3: DAG Resolution
	readySteps := GetReadySteps(c.task.Plan)

	if len(readySteps) == 0 {
		// Are there any pending steps left?
		hasPending := false
		for _, s := range c.task.Plan {
			if s.Status == types.StepStatusPending || s.Status == types.StepStatusRunning || s.Status == types.StepStatusPendingAsync {
				hasPending = true
				break
			}
		}
		if !hasPending {
			// Phase 10: Loop Termination (Success)
			c.mu.Lock()
			c.task.Status = types.TaskStatusCompleted
			now := time.Now().UTC()
			c.task.FinishedAt = &now
			c.mu.Unlock()
		}
		return nil
	}

	// Per-Agent Throttling (Phase 3.2)
	// We track how many steps are currently running for each agent in this task.
	agentCounts := make(map[string]int)
	for _, s := range c.task.Plan {
		if s.Status == types.StepStatusRunning || s.Status == types.StepStatusPendingAsync {
			// In a real impl, we'd know which agent was assigned.
			// Here we assume "mock.agent" for simplicity.
			agentCounts["mock.agent"]++
		}
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(readySteps))

	for _, sPtr := range readySteps {
		// Check throttle
		targetAgent := "mock.agent" 
		if agentCounts[targetAgent] >= c.config.MaxConcurrentPerAgent {
			fmt.Printf("Coordinator [task=%s]: Throttling step %s for agent %s\n", c.task.TaskID, sPtr.StepID, targetAgent)
			continue
		}
		agentCounts[targetAgent]++

		wg.Add(1)
		go func(step *types.Step) {
			defer wg.Done()
			
			// Phase 4: Interpolation
			c.mu.Lock()
			hotStateMap := c.stateMgr.GetSnapshot() // Collapses DeltaLog and returns map
			c.mu.Unlock()

			resolvedInput, err := ResolveInterpolation(step.Input, hotStateMap)
			if err != nil {
				c.failStep(step, "INTERPOLATION_FAILED", err.Error())
				errChan <- err
				return
			}
			step.Input = resolvedInput

			// Phase 6: Agent Routing & Dispatch
			// We skip Policy (Phase 5) for brevity, assuming the policy engine runs globally.
			agentManifest := &types.AgentManifest{AgentID: "mock.agent"} // Fetched from Registry in real impl
			
			req := &types.Request{
				Type:       "REQUEST",
				RequestID:  "req-" + step.StepID,
				TaskID:     c.task.TaskID,
				StepID:     step.StepID,
				Capability: step.Capability,
				Input:      step.Input,
				Mode:       types.RequestModeSync,
			}

			c.mu.Lock()
			step.Status = types.StepStatusRunning
			c.mu.Unlock()

			resp, err := c.transport.Dispatch(req, agentManifest)
			if err != nil {
				c.failStep(step, "DISPATCH_FAILED", err.Error())
				errChan <- err
				return
			}

			// Phase 7: Result Processing
			if resp.Success {
				c.mu.Lock()
				step.Status = types.StepStatusDone
				step.Output = resp.Output
				
				// Add to Delta Log safely
				c.stateMgr.AppendResult(step.StepID, step.Output)
				c.mu.Unlock()
			} else {
				c.failStep(step, "EXECUTION_FAILED", "Agent returned error")
				errChan <- fmt.Errorf("agent returned failure")
			}
		}(sPtr)
	}

	wg.Wait()
	close(errChan)

	// If any step failed unrecoverably, we would trigger Escalation (Phase 8) here.
	// For now, we return the first error encountered, if any.
	for err := range errChan {
		if err != nil {
			return err
		}
	}
	return nil
}

// failStep is a helper to mark a step as failed.
func (c *Coordinator) failStep(step *types.Step, code string, msg string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	step.Status = types.StepStatusFailed
	step.Error = &types.Error{
		Code:    types.ErrorCode(code),
		Message: msg,
	}
}

// HandleAsyncCompletion implements Phase 9: Async Callback Endpoint.
func (c *Coordinator) HandleAsyncCompletion(jobID string, completion *types.AsyncCompletion) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var targetStep *types.Step
	for i := range c.task.Plan {
		if c.task.Plan[i].JobID == jobID {
			targetStep = &c.task.Plan[i]
			break
		}
	}

	if targetStep == nil {
		return fmt.Errorf("coordinator: unknown job_id %s", jobID)
	}

	if completion.Success {
		targetStep.Status = types.StepStatusDone
		targetStep.Output = completion.Output
		c.stateMgr.AppendResult(targetStep.StepID, targetStep.Output)
	} else {
		targetStep.Status = types.StepStatusFailed
		targetStep.Error = completion.Error
	}

	return nil
}
