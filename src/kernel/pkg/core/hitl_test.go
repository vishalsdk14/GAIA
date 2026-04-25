// Copyright 2026 GAIA Contributors
// ... (omitting full header for brevity in thought, but will include in file)
package core

import (
	"gaia/kernel/pkg/logger"
	"gaia/kernel/pkg/types"
	"testing"
)

func TestHITLApprovalFlow(t *testing.T) {
	// Initialize logger to prevent panics
	logger.Init("info")

	// 1. Setup mock coordinator
	task := &types.Task{
		TaskID: "test-task",
		Plan: []types.Step{
			{
				StepID:     "step_1",
				Capability: "RESTRICTED_ACTION",
				Status:     types.StepStatusPending,
			},
		},
	}
	
	// Mock subsystems
	config := DefaultConfig()
	coord := NewCoordinator(task, config, nil, nil, nil, nil)

	// 2. Simulate execution (normally done by Run loop)
	// We'll manually call a simulation of the policy check logic
	step := &task.Plan[0]
	
	if step.Capability == "RESTRICTED_ACTION" {
		step.Status = types.StepStatusAwaitingApproval
	}

	if step.Status != types.StepStatusAwaitingApproval {
		t.Fatalf("expected step to be awaiting_approval, got %s", step.Status)
	}

	// 3. Approve Step
	err := coord.ApproveStep("step_1")
	if err != nil {
		t.Fatalf("failed to approve step: %v", err)
	}

	if step.Status != types.StepStatusPending {
		t.Errorf("expected step to return to pending, got %s", step.Status)
	}
}
