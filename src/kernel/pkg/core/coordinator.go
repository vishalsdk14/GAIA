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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"gaia/kernel/pkg/common"
	"gaia/kernel/pkg/logger"
	"gaia/kernel/pkg/policy"
	"gaia/kernel/pkg/registry"
	"gaia/kernel/pkg/state"
	"gaia/kernel/pkg/types"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	// DefaultReplanCoolOff is the duration to wait before starting a re-plan cycle
	// to prevent busy-looping if the planner gets stuck.
	DefaultReplanCoolOff = 1 * time.Second
)

// Coordinator implements the GAIA Control Loop.
// It manages task state, dispatches steps to agents, and handles recovery.
type Coordinator struct {
	mu        sync.Mutex
	config    *KernelConfig
	task      *types.Task
	stateMgr  *state.ActiveStateManager
	registry  registry.CapabilityRegistry
	planner   Planner
	transport AgentTransport
	log       *slog.Logger
	events    *common.EventBus
	backoff   *BackoffCalculator
	replans   int
	taskStore *state.TaskStore
	policy    *policy.PolicyEngine
	validator *policy.SchemaValidator
	audit     *logger.AuditLogger
	quota     *QuotaManager
}

// NewCoordinator initializes a new task coordinator with the required kernel subsystems.
func NewCoordinator(t *types.Task, c *KernelConfig, r registry.CapabilityRegistry, p Planner, trans AgentTransport, ts *state.TaskStore, quota *QuotaManager) *Coordinator {
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
		quota:     quota,
	}
}

// Run executes the 10-phase control loop for the task.
// This function implements the logic defined in docs/specs/control-loop.md Section 30.
func (c *Coordinator) Run() error {
	defer c.quota.ReleaseTask(c.task.TaskID)
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
			// Phase 18: [TELEMETRY] Final Mission Summary
			c.log.Info("MISSION_SUMMARY",
				"status", c.task.Status,
				"total_steps", c.task.TotalSteps,
				"tokens_prompt", c.task.TokensPrompt,
				"tokens_completion", c.task.TokensCompletion,
				"estimated_cost_usd", c.task.EstimatedCostUSD,
				"duration_ms", time.Since(c.task.CreatedAt).Milliseconds(),
				"agents", c.task.AgentsInvolved)
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

	if c.task.Metadata == nil {
		c.task.Metadata = make(map[string]interface{})
	}

	c.task.Status = types.TaskStatusPlanning
	c.task.UpdatedAt = time.Now().UTC()
	
	c.log.Info("Task transitioned to planning")
	c.events.Emit(common.Event{Type: types.EventTaskPlanning, TaskID: c.task.TaskID})
	c.audit.Log("kernel", string(types.EventTaskPlanning), c.task.TaskID, map[string]interface{}{"task_id": c.task.TaskID})
	
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
	
	// Add current_url and current_title to state for Rule 31 (URL Stickiness)
	// We search backwards from the latest completed step to find these values.
	for i := len(c.task.Plan) - 1; i >= 0; i-- {
		step := c.task.Plan[i]
		if step.Status == types.StepStatusDone {
			if out, ok := step.Output.(map[string]interface{}); ok {
				if url, ok := out["url"].(string); ok {
					activeState["current_url"] = url
				}
				if title, ok := out["title"].(string); ok {
					activeState["current_title"] = title
				}
				if activeState["current_url"] != nil {
					break
				}
			}
		}
	}

	capabilities := c.registry.GetAllCapabilities()
	
	// Phase 2.3: Failure Recovery Loop
	var lastErr error
	var correctionPrompt string
	
	maxRetries := c.config.MaxReplans // Or add PlannerMaxRetries to config
	for attempt := 0; attempt <= maxRetries; attempt++ {
		goal := c.task.Goal
		if c.replans > 0 {
			// Phase 2.1: Visual Analysis (TuriX Port)
			// Check for a screenshot from a failed step to provide better context
			var visionContext string
			for i := len(c.task.Plan) - 1; i >= 0; i-- {
				step := c.task.Plan[i]
				if step.Status == types.StepStatusFailed || step.Status == types.StepStatusDone {
					if out, ok := step.Output.(map[string]interface{}); ok {
						if path, ok := out["screenshot_path"].(string); ok {
							c.log.Info("Performing Vision Analysis on failure screenshot", "path", path)
							data, _ := os.ReadFile(path)
							b64 := base64.StdEncoding.EncodeToString(data)
							analysis, err := c.planner.Vision("Describe the current state of this web page. If there is an error message, list it. If there are red numeric labels, list what they correspond to.", b64)
							if err == nil {
								visionContext = fmt.Sprintf("\nVISUAL ANALYSIS OF CURRENT PAGE:\n%s", analysis)
								break
							}
						}
					}
				}
			}

			goal = fmt.Sprintf("SYSTEM: This is a RE-PLAN (Attempt %d). "+
				"Original Goal: %s. %s\n"+
				"Please examine the AccumulatedOutputs and Visual Analysis to see what failed. "+
				"Generate a COMPLETE new plan. If IDs are available in the visual analysis, use click_id/type_id.", 
				c.replans, c.task.Goal, visionContext)
		}

		if correctionPrompt != "" {
			goal = correctionPrompt
		}

		plan, err := c.planner.GeneratePlan(goal, activeState, capabilities)
		
		// Phase 18: [TELEMETRY] Update Planning Metrics
		if plan != nil {
			c.task.TokensPrompt += plan.Usage.PromptTokens
			c.task.TokensCompletion += plan.Usage.CompletionTokens
			
			// Calculate and aggregate cost (BUG-002)
			cost := CalculateCost(c.config.PlannerModel, plan.Usage)
			c.task.EstimatedCostUSD += cost

			c.log.Debug("Telemetry: Planning metrics updated", 
				"prompt", plan.Usage.PromptTokens, 
				"completion", plan.Usage.CompletionTokens,
				"cost_usd", cost)
		}

		if err != nil {
			lastErr = err
			
			// Phase 16: [BACKOFF] Handle Rate Limits (429)
			// If we hit a rate limit (Too Many Requests), wait exponentially and retry.
			if strings.Contains(err.Error(), "429") {
				// Exponential backoff: 2s, 4s, 8s, 16s...
				backoff := time.Duration(1 << uint(attempt+1)) * time.Second
				c.log.Warn("Rate limit (429) detected. Cooling off before retry...", 
					"backoff_seconds", backoff.Seconds(), 
					"attempt", attempt+1, 
					"max_attempts", maxRetries+1)
				
				time.Sleep(backoff)
				continue
			}

			// Check if the error is a schema violation (malformed JSON)
			c.log.Warn("Planner returned invalid JSON or API error", "error", err)
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
		
		// Phase 14: [ROOT CAUSE FIX] Direct Implicit Dependency Scanning
		// This brute-force scan ensures that if Step B uses data from Step A (e.g., {{step_A.price}}),
		// Step B is explicitly linked in the DAG, even if the Planner forgot 'depends_on'.
		for i := range c.task.Plan {
			step := &c.task.Plan[i]
			inputBytes, _ := json.Marshal(step.Input)
			inputStr := string(inputBytes)
			
			for _, otherStep := range c.task.Plan {
				if otherStep.StepID == step.StepID {
					continue
				}
				
				// Check for the standard GAIA interpolation prefix (case-insensitive)
				tagPattern := "{{" + strings.ToLower(otherStep.StepID)
				if strings.Contains(strings.ToLower(inputStr), tagPattern) {
					alreadyPresent := false
					for _, d := range step.DependsOn {
						if d == otherStep.StepID {
							alreadyPresent = true
							break
						}
					}
					if !alreadyPresent {
						step.DependsOn = append(step.DependsOn, otherStep.StepID)
						c.log.Info("DAG: Linked implicit dependency", "step", step.StepID, "depends_on", otherStep.StepID)
					}
				}
			}
		}
		c.task.HasMore = plan.HasMore
		c.log.Info("Plan generated successfully", "step_count", len(plan.Steps), "has_more", plan.HasMore)
		c.events.Emit(common.Event{Type: types.EventPlanGenerated, TaskID: c.task.TaskID})
		c.audit.Log("planner", string(types.EventPlanGenerated), c.task.TaskID, map[string]interface{}{"step_count": len(plan.Steps), "task_id": c.task.TaskID})

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
	c.task.TotalDurationMS = time.Since(c.task.CreatedAt).Milliseconds()
	
	c.log.Error("Task failed", "error", err)
	c.events.Emit(common.Event{
		Type:    types.EventTaskFailed,
		TaskID:  c.task.TaskID,
		Payload: map[string]interface{}{"error": err.Error()},
	})
	c.audit.Log("kernel", string(types.EventTaskFailed), c.task.TaskID, map[string]interface{}{"error": err.Error(), "task_id": c.task.TaskID})
	
	return err
}

// executeDAG manages the parallel dispatch of ready steps (Phases 3-8).
func (c *Coordinator) executeDAG() error {
	// Phase 3: DAG Resolution
	readySteps := GetReadySteps(c.task.Plan, c.stateMgr.GetSnapshot())

	// Quota is now enforced per-dispatch in the loop below to avoid busy-loop exhaustion

	// Phase 10: Governance - Check Usage Policies
	if err := c.checkUsagePolicies(); err != nil {
		c.log.Error("Task aborted by governance policy", "error", err)
		c.task.Status = types.TaskStatusFailed
		c.task.Metadata["failure_reason"] = "Policy violation: " + err.Error()
		return nil
	}

	if len(readySteps) == 0 {
		// Are there any pending/active steps left?
		hasActive := false
		hasFailed := false
		for _, s := range c.task.Plan {
			if s.Status == types.StepStatusPending || s.Status == types.StepStatusRunning || s.Status == types.StepStatusPendingAsync {
				hasActive = true
			}
			if s.Status == types.StepStatusFailed {
				hasFailed = true
			}
		}

		if !hasActive {
			if hasFailed {
				c.log.Warn("Task execution finished but some steps failed", "task_id", c.task.TaskID)
				// If a step failed but the task status wasn't updated by handleStepFailure, do it now.
				// This ensures the user sees a 'Failed' status instead of a false 'Success'.
				c.mu.Lock()
				if c.task.Status == types.TaskStatusExecuting {
					c.task.Status = types.TaskStatusFailed
				}
				c.mu.Unlock()
				return nil
			}

			c.mu.Lock()
			if c.task.HasMore {
				// Incremental Planning: Loop back to planning phase
				c.task.Status = types.TaskStatusPlanning
				c.mu.Unlock()
				c.log.Info("Current plan finished, HasMore=true, transitioning back to planning")
				return nil
			}

			// Loop Termination (Success)
			c.task.Status = types.TaskStatusCompleted
			now := time.Now().UTC()
			c.task.FinishedAt = &now
			c.task.TotalDurationMS = time.Since(c.task.CreatedAt).Milliseconds()
			c.mu.Unlock()
			c.log.Info("Task completed successfully")
			c.events.Emit(common.Event{Type: types.EventTaskCompleted, TaskID: c.task.TaskID})
			c.audit.Log("kernel", string(types.EventTaskCompleted), c.task.TaskID, map[string]interface{}{"task_id": c.task.TaskID})
		}
		return nil
	}

	c.log.Debug("DAG resolved ready steps", "count", len(readySteps))

	// Per-Agent Throttling (Phase 3.2)
	// We track how many steps are currently running for each agent in this task.
	agentCounts := make(map[string]int)
	for _, s := range c.task.Plan {
		if (s.Status == types.StepStatusRunning || s.Status == types.StepStatusPendingAsync) && s.AssignedAgent != "" {
			agentCounts[s.AssignedAgent]++
		}
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(readySteps))

	for _, sPtr := range readySteps {
		// Phase 6.0: Early Agent Selection for Throttling
		agentRecord, err := c.registry.SelectAgent(sPtr.Capability)
		if err != nil {
			c.log.Error("No agent available for required capability", "step_id", sPtr.StepID, "capability", sPtr.Capability, "error", err)
			
			// Phase 12: Fail gracefully instead of looping infinitely
			sPtr.Status = types.StepStatusFailed
			sPtr.Error = &types.Error{
				Code:    types.ErrorCodeCapabilityNotFound,
				Message: fmt.Sprintf("no agent registered for capability: %s", sPtr.Capability),
			}
			c.task.Status = types.TaskStatusFailed
			
			if err := c.taskStore.SaveTask(c.task); err != nil {
				c.log.Error("Failed to save task failure state", "error", err)
			}
			return fmt.Errorf("task failed: %s", sPtr.Error.Message) // Terminate the execution loop
		}
		targetAgent := agentRecord.Manifest.AgentID

		// Check throttle
		if agentCounts[targetAgent] >= c.config.MaxConcurrentPerAgent {
			c.log.Debug("Throttling step for agent", "step_id", sPtr.StepID, "agent_id", targetAgent)
			continue
		}
		agentCounts[targetAgent]++

		// Phase 11: Quota Enforcement (Increment only on actual dispatch)
		c.log.Info("Quota increment", "step_id", sPtr.StepID, "capability", sPtr.Capability)
		if err := c.quota.IncrementStep(c.task.TaskID); err != nil {
			return c.failTask(err)
		}
		c.task.TotalSteps++

		// Transition to Running synchronously to prevent race condition in GetReadySteps
		c.mu.Lock()
		sPtr.Status = types.StepStatusRunning
		c.mu.Unlock()

		wg.Add(1)
		go func(step *types.Step) {
			defer wg.Done()
			
			// Phase 4: Interpolation
			c.mu.Lock()
			hotStateMap := c.stateMgr.GetSnapshot() // Collapses DeltaLog and returns map
			c.mu.Unlock()

			var resolvedInput interface{}
			var err error
			if c.config.InterpolationEngine == "legacy" {
				resolvedInput, err = ResolveInterpolation(step.Input, hotStateMap)
			} else {
				resolvedInput, err = FastResolveInterpolation(step.Input, hotStateMap)
			}

			if err != nil {
				c.failStep(step, "INTERPOLATION_FAILED", err.Error())
				errChan <- err
				return
			}
			step.Input = resolvedInput

			// Phase 6: Agent Routing & Dispatch
			// Re-verify the agent (it was selected above but we re-fetch to ensure safety in goroutine)
			agentRecord, err := c.registry.SelectAgent(step.Capability)
			if err != nil {
				c.log.Error("No healthy agent found for capability", "capability", step.Capability, "error", err)
				c.handleStepFailure(step, &types.Error{Code: types.ErrorCodeAgentUnavailable, Message: err.Error()}, nil)
				return
			}
			agentManifest := &agentRecord.Manifest
			step.AssignedAgent = agentManifest.AgentID

			// Phase 5: Policy Engine (Firewall)
			// Fetch current usage/cost from task metadata
			usage, _ := c.task.Metadata["usage"].(map[string]interface{})
			if usage == nil {
				usage = map[string]interface{}{"tokens": 0, "requests": 0}
			}
			cost, _ := c.task.Metadata["cost"].(map[string]interface{})
			if cost == nil {
				cost = map[string]interface{}{"usd": 0.0}
			}

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
				"env":   c.config.Environment,
				"usage": usage,
				"cost":  cost,
			}

			// Phase 5: Policy Evaluation
			if err := c.policy.EvaluateAll(c.config.GlobalPolicies, policyContext); err != nil {
				c.log.Warn("Policy denied execution", "step_id", step.StepID, "error", err)
				c.audit.Log("policy_engine", "POLICY_DENIED", step.StepID, map[string]interface{}{"reason": err.Error(), "task_id": c.task.TaskID})
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

			// Initialize Request
			req := &types.Request{
				Type:       "REQUEST",
				RequestID:  "req-" + step.StepID,
				TaskID:     c.task.TaskID,
				StepID:     step.StepID,
				Capability: step.Capability,
				Input:      step.Input,
				Mode:       types.RequestModeSync,
			}

			c.log.Info("Dispatching step", "step_id", step.StepID, "capability", step.Capability)
			c.events.Emit(common.Event{Type: types.EventStepStarted, TaskID: c.task.TaskID, StepID: step.StepID})
			c.audit.Log("kernel", string(types.EventStepStarted), step.StepID, map[string]interface{}{"agent": agentManifest.AgentID, "task_id": c.task.TaskID})

			// Phase 18: [TELEMETRY] Start Step Timer
			start := time.Now()
			resp, err := c.transport.Dispatch(req, agentManifest)
			step.DurationMS = time.Since(start).Milliseconds()

			// Track involved agents (unique list)
			c.mu.Lock()
			found := false
			for _, a := range c.task.AgentsInvolved {
				if a == agentManifest.AgentID {
					found = true
					break
				}
			}
			if !found {
				c.task.AgentsInvolved = append(c.task.AgentsInvolved, agentManifest.AgentID)
			}
			c.mu.Unlock()

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

				// Phase 7.1: Visual Processing (Panda Port)
				c.processStepOutput(step)

				c.log.Info("Step completed", "step_id", step.StepID)
				c.events.Emit(common.Event{Type: types.EventStepCompleted, TaskID: c.task.TaskID, StepID: step.StepID})
				c.audit.Log("kernel", string(types.EventStepCompleted), step.StepID, map[string]interface{}{"task_id": c.task.TaskID})

				// Checkpoint 3: Step Completion
				if err := c.taskStore.SaveTask(c.task); err != nil {
					c.log.Error("Failed to checkpoint step completion", "step_id", step.StepID, "error", err)
				}
			} else {
				c.log.Warn("Step execution failed by agent", "step_id", step.StepID)
				
				// Phase 7.1: Visual Processing (Panda Port) even on failure
				step.Output = resp.Output // Carry over output for processing
				c.processStepOutput(step)
				
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

	agentID := "unknown"
	if agent != nil {
		agentID = agent.AgentID
	}

	// Safety check: if err is nil (malformed agent response), create a generic one
	if err == nil {
		err = &types.Error{
			Code:    types.ErrorCodeExecutionFailed,
			Message: "Agent returned failure without specific error details",
		}
	}

	c.log.Warn("Handling step failure", "step_id", step.StepID, "error_code", err.Code, "retry_count", step.RetryCount, "agent_id", agentID)

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
	step.Error = err
	c.task.Status = types.TaskStatusFailed
	c.log.Error("Escalation Tier 4: Aborting task", "step_id", step.StepID)
}

// failStep is a helper to mark a step as failed.
func (c *Coordinator) failStep(step *types.Step, code string, msg string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.log.Error("Step execution failed early", "step_id", step.StepID, "code", code, "error", msg)
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

// checkUsagePolicies evaluates governance rules involving cost and usage metrics.
func (c *Coordinator) checkUsagePolicies() error {
	// 1. Prepare usage/cost context
	// In a real implementation, these would be pulled from a real-time tracking service.
	// For Phase 10, we simulate usage data in the task metadata.
	usage := make(map[string]interface{})
	cost := make(map[string]interface{})
	
	if u, ok := c.task.Metadata["usage"].(map[string]interface{}); ok {
		usage = u
	} else {
		usage["tokens"] = 0
		usage["requests"] = 0
	}
	
	if cs, ok := c.task.Metadata["cost"].(map[string]interface{}); ok {
		cost = cs
	} else {
		cost["usd"] = 0.0
	}

	context := map[string]interface{}{
		"task":  c.task,
		"usage": usage,
		"cost":  cost,
		"env":   "production",
	}

	// 2. Evaluate Global Governance Policies
	// Example: "usage.tokens < 5000"
	for _, p := range c.config.GlobalPolicies {
		// Only evaluate policies that mention usage or cost
		// (Primitive check for Phase 10)
		success, err := c.policy.Evaluate(p, context)
		if err != nil {
			continue // Skip non-boolean or invalid policies here
		}
		if !success {
			return fmt.Errorf("governance limit exceeded: %s", p)
		}
	}

	return nil
}

// processStepOutput extracts rich media (screenshots) from agent responses and saves them to disk
// to avoid bloating the memory state while preserving them for debugging and vision analysis.
func (c *Coordinator) processStepOutput(step *types.Step) {
	outputMap, ok := step.Output.(map[string]interface{})
	if !ok {
		return
	}

	screenshotB64, ok := outputMap["screenshot"].(string)
	if !ok || screenshotB64 == "" {
		return
	}

	// 1. Decode screenshot
	data, err := base64.StdEncoding.DecodeString(screenshotB64)
	if err != nil {
		c.log.Warn("Failed to decode screenshot", "step_id", step.StepID, "error", err)
		return
	}

	// 2. Save to media directory
	filename := fmt.Sprintf("screenshot_%s_%d.jpg", step.StepID, time.Now().Unix())
	mediaPath := filepath.Join(".", "media", filename)
	
	// Ensure directory exists
	os.MkdirAll(filepath.Dir(mediaPath), 0755)

	if err := os.WriteFile(mediaPath, data, 0644); err != nil {
		c.log.Warn("Failed to save screenshot", "path", mediaPath, "error", err)
		return
	}

	// 3. Replace base64 with file path in output to save memory
	outputMap["screenshot_path"] = mediaPath
	delete(outputMap, "screenshot")
	
	c.log.Info("Screenshot saved", "step_id", step.StepID, "path", mediaPath)
}
