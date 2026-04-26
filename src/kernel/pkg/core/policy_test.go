package core

import (
	"gaia/kernel/pkg/logger"
	"gaia/kernel/pkg/types"
	"testing"
)

func TestCheckUsagePolicies(t *testing.T) {
	config := DefaultConfig()
	config.GlobalPolicies = []string{"usage.tokens < 100"}

	task := &types.Task{
		TaskID: "test-task",
		Goal:   "Test Policy",
		Metadata: map[string]interface{}{
			"usage": map[string]interface{}{
				"tokens": 150,
			},
		},
	}

	logger.Init("INFO")
	coord := NewCoordinator(task, config, nil, nil, nil, nil)
	err := coord.checkUsagePolicies()
	if err == nil {
		t.Fatal("expected error due to policy violation, got nil")
	}

	expectedErr := "governance limit exceeded: usage.tokens < 100"
	if err.Error() != expectedErr {
		t.Fatalf("expected error %q, got %q", expectedErr, err.Error())
	}

	// Test passing policy
	task.Metadata["usage"] = map[string]interface{}{"tokens": 50}
	err = coord.checkUsagePolicies()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
