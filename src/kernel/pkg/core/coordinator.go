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
	"gaia/kernel/pkg/common"
	"gaia/kernel/pkg/logger"
	"gaia/kernel/pkg/policy"
	"gaia/kernel/pkg/registry"
	"gaia/kernel/pkg/state"
	"gaia/kernel/pkg/types"
	"log/slog"
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
	log      *slog.Logger
	events   *common.EventBus
	backoff   *BackoffCalculator
	replans   int
	taskStore *state.TaskStore
	policy    *policy.PolicyEngine
	validator *policy.SchemaValidator
	audit     *logger.AuditLogger
}

// NewCoordinator initializes a new task coordinator with the required kernel subsystems.
func NewCoordinator(t *types.Task, c *KernelConfig, r registry.CapabilityRegistry, p Planner, trans AgentTransport, ts *state.TaskStore) *Coordinator {
	pe, _ := policy.NewPolicyEngine()
	return &Coordinator{
		task:      t,
		config:    c,
		stateMgr:  state.NewActiveStateManager(t.TaskID),
		registry:  r,
		planner:   p,
		transport: trans,
		log:       logger.Sub("task_id", t.TaskID),
		events:    common.GetEventBus(),
		backoff:   NewBackoffCalculator(),
		taskStore: ts,
		policy:    pe,
		validator: &policy.SchemaValidator{},
		audit:     logger.GetAuditLogger(),
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

	c.task.Status = types.TaskStatusPlanning
	c.task.UpdatedAt = time.Now().UTC()
	
	c.log.Info("Task transitioned to planning")
	c.events.Emit(common.Event{Type: types.EventTaskPlanning, TaskID: c.task.TaskID})
	c.audit.Log("kernel", string(types.EventTaskPlanning), c.task.TaskID, nil)
	
	// Checkpoint 1: Initial Submission
	if err := c.taskStore.SaveTask(c.task); err != nil {
		c.log.Error("Failed to checkpoint task in Phase 1", "error", err)
	}
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
		if err != nil {
			// Check if the error is a schema violation (malformed JSON)
			lastErr = err
			c.log.Warn("Planner returned invalid JSON", "error", err)
			if types.ErrorCode(err.Error()) == types.ErrorCodeSchemaViolation || attempt == 0 {
				correctionPrompt = BuildCorrectionPrompt("INVALID_JSON_HERE", err.Error())
				c.events.Emit(common.Event{Type: types.EventPlanRejected, TaskID: c.task.TaskID})
				continue
			}
			break
		}

		// Success!
		c.task.Status = types.TaskStatusExecuting
		c.task.UpdatedAt = time.Now().UTC()
		c.task.Plan = plan.Steps
		c.log.Info("Plan generated successfully", "step_count", len(plan.Steps))
		c.events.Emit(common.Event{Type: types.EventPlanGenerated, TaskID: c.task.TaskID})
		c.audit.Log("planner", string(types.EventPlanGenerated), c.task.TaskID, map[string]interface{}{"step_count": len(plan.Steps)})

		// Checkpoint 2: Plan Generation
		if err := c.taskStore.SaveTask(c.task); err != nil {
			c.log.Error("Failed to checkpoint plan in Phase 2", "error", err)
		}
		return nil
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

func (c *Coordinator) failTask(err error) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.task.Status = types.TaskStatusFailed
	c.task.UpdatedAt = time.Now().UTC()
	
	c.log.Error("Task failed", "error", err)
	c.events.Emit(common.Event{
		Type:    types.EventTaskFailed,
		TaskID:  c.task.TaskID,
		Payload: map[string]interface{}{"error": err.Error()},
	})
	c.audit.Log("kernel", string(types.EventTaskFailed), c.task.TaskID, map[string]interface{}{"error": err.Error()})
	
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
			c.log.Info("Task completed successfully")
			c.events.Emit(common.Event{Type: types.EventTaskCompleted, TaskID: c.task.TaskID})
		}
		return nil
	}

	c.log.Debug("DAG resolved ready steps", "count", len(readySteps))

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

			// Phase 5: Policy Engine (Firewall)
			// We build a context for the CEL engine
			agentManifest := &types.AgentManifest{AgentID: "mock.agent"} // Fetched from Registry in real impl
			policyContext := map[string]interface{}{
				"task": map[string]interface{}{
					"goal": c.task.Goal,
				},
				"step": map[string]interface{}{
					"capability": step.Capability,
				},
				"agent": map[string]interface{}{
					"agent_id": agentManifest.AgentID,
				},
				"env": c.config.Environment,
			}

			// Example: Global kernel policy strings (would be in config in real impl)
			globalPolicies := []string{
				"env == 'production' ? true : true", // Placeholder
			}

			if err := c.policy.EvaluateAll(globalPolicies, policyContext); err != nil {
				c.log.Warn("Policy denied execution", "step_id", step.StepID, "error", err)
				c.audit.Log("policy_engine", "POLICY_DENIED", step.StepID, map[string]interface{}{"reason": err.Error()})
				c.handleStepFailure(step, &types.Error{Code: types.ErrorCodePolicyDenied, Message: err.Error()}, agentManifest)
				return
			}

			// Phase 5.0.1: Human-in-the-Loop Simulation
			// If a capability is marked as "RESTRICTED_ACTION", pause execution for approval.
			if step.Capability == "RESTRICTED_ACTION" {
				c.mu.Lock()
				step.Status = types.StepStatusAwaitingApproval
				c.mu.Unlock()
				c.log.Warn("Step requires human approval", "step_id", step.StepID)
				c.events.Emit(common.Event{Type: types.EventStepApprovalRequired, TaskID: c.task.TaskID, StepID: step.StepID})
				return
			}

			// Phase 5.1: Schema Validation
			// Find capability schema in manifest
			var capSchema map[string]interface{}
			for _, cap := range agentManifest.Capabilities {
				if cap.Name == step.Capability {
					capSchema = cap.InputSchema
					break
				}
			}

			if err := c.validator.ValidateStepInput(step.Input, capSchema); err != nil {
				c.log.Warn("Schema violation detected", "step_id", step.StepID, "error", err)
				c.handleStepFailure(step, &types.Error{Code: types.ErrorCodeSchemaViolation, Message: err.Error()}, agentManifest)
				return
			}

			// Phase 6: Agent Routing & Dispatch
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

			c.log.Info("Dispatching step", "step_id", step.StepID, "capability", step.Capability)
			c.events.Emit(common.Event{Type: types.EventStepStarted, TaskID: c.task.TaskID, StepID: step.StepID})
			c.audit.Log("kernel", string(types.EventStepStarted), step.StepID, map[string]interface{}{"agent": agentManifest.AgentID})

			resp, err := c.transport.Dispatch(req, agentManifest)
			if err != nil {
				c.log.Error("Step dispatch failed", "step_id", step.StepID, "error", err)
				c.handleStepFailure(step, &types.Error{Code: types.ErrorCodeAgentUnavailable, Message: err.Error()}, agentManifest)
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

				c.log.Info("Step completed", "step_id", step.StepID)
				c.events.Emit(common.Event{Type: types.EventStepCompleted, TaskID: c.task.TaskID, StepID: step.StepID})

				// Checkpoint 3: Step Completion
				if err := c.taskStore.SaveTask(c.task); err != nil {
					c.log.Error("Failed to checkpoint step completion", "step_id", step.StepID, "error", err)
				}
			} else {
				c.log.Warn("Step execution failed by agent", "step_id", step.StepID)
				c.handleStepFailure(step, resp.Error, agentManifest)
			}
		}(sPtr)
	}

	wg.Wait()
	close(errChan)

	// In Phase 4, we no longer return the first error from executeDAG immediately.
	// Instead, we let the loop iterate. If a step failed and triggered a replan,
	// the task status will change in the next iteration.
	return nil
}

// handleStepFailure implements the 4-tier escalation path (Retry -> Fallback -> Replan -> Abort).
func (c *Coordinator) handleStepFailure(step *types.Step, err *types.Error, agent *types.AgentManifest) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.log.Warn("Handling step failure", "step_id", step.StepID, "error_code", err.Code, "retry_count", step.RetryCount)

	// Tier 1: Retry (if retryable and attempts < max)
	// We use the configured kernel defaults if no specific policy exists on the step.
	policy := &types.RetryPolicy{
		MaxAttempts: c.config.RetryMaxAttempts,
		Backoff:     "exponential",
		BaseDelayMS: c.config.RetryBaseDelayMS,
		MaxDelayMS:  c.config.RetryMaxDelayMS,
	}

	if IsRetryable(step, err, agent) && step.RetryCount < policy.MaxAttempts {
		step.RetryCount++
		delay := c.backoff.GetDelay(policy, step.RetryCount)
		
		c.log.Info("Escalation Tier 1: Retrying step", "step_id", step.StepID, "delay_ms", delay.Milliseconds())
		
		// Reset to pending so GetReadySteps picks it up again in next iteration
		// In a real impl, we'd use a timer to delay the transition
		step.Status = types.StepStatusPending
		return
	}

	// Tier 2: Fallback (Skipped for brevity in Task 1, usually involves Registry lookup for same capability)

	// Tier 3: Replan
	if c.replans < c.config.MaxReplans {
		c.replans++
		c.task.Status = types.TaskStatusPlanning
		c.log.Info("Escalation Tier 3: Triggering Re-plan", "replan_count", c.replans)
		c.events.Emit(common.Event{Type: types.EventReplanTriggered, TaskID: c.task.TaskID})
		return
	}

	// Tier 4: Abort
	step.Status = types.StepStatusFailed
	step.Error = err
	c.task.Status = types.TaskStatusFailed
	c.log.Error("Escalation Tier 4: Aborting task", "step_id", step.StepID)
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

// ApproveStep manually unblocks a step that is in the StepStatusAwaitingApproval state.
// This allows the task loop to proceed with the next Topological dispatch iteration.
func (c *Coordinator) ApproveStep(stepID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i := range c.task.Plan {
		if c.task.Plan[i].StepID == stepID {
			if c.task.Plan[i].Status != types.StepStatusAwaitingApproval {
				return fmt.Errorf("coordinator: step %s is not awaiting approval (current status: %s)", stepID, c.task.Plan[i].Status)
			}
			
			// Move back to pending so GetReadySteps can pick it up
			c.task.Plan[i].Status = types.StepStatusPending
			c.log.Info("Step approved by human", "step_id", stepID)
			
			// Checkpoint the updated step status
			if c.taskStore != nil {
				if err := c.taskStore.SaveTask(c.task); err != nil {
					c.log.Error("Failed to checkpoint approval", "step_id", stepID, "error", err)
				}
			}

			// Emit event for real-time dashboard updates
			c.events.Emit(common.Event{
				Type:   types.EventStepStarted, // We emit Started or a new EventStepApproved
				TaskID: c.task.TaskID,
				StepID: stepID,
			})
			return nil
		}
	}
	return fmt.Errorf("coordinator: step %s not found in plan", stepID)
}

// UpdatePlan allows manual modification of the task plan mid-execution.
// This implements the "Manual Overrides" requirement for Phase 9.
func (c *Coordinator) UpdatePlan(newPlan []types.Step) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Validation: Ensure we don't modify steps that are already 'done' or 'running'
	// in a way that breaks consistency (simple implementation for now).
	statusMap := make(map[string]types.StepStatus)
	for _, step := range c.task.Plan {
		statusMap[step.StepID] = step.Status
	}

	for _, newStep := range newPlan {
		if currentStatus, exists := statusMap[newStep.StepID]; exists {
			if (currentStatus == types.StepStatusDone || currentStatus == types.StepStatusRunning) && 
				newStep.Status != currentStatus {
				return fmt.Errorf("coordinator: cannot modify status of step %s (currently %s)", newStep.StepID, currentStatus)
			}
		}
	}

	c.task.Plan = newPlan
	c.log.Info("Task plan updated manually")
	
	if c.taskStore != nil {
		return c.taskStore.SaveTask(c.task)
	}
	return nil
}
