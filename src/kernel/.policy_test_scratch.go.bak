package main

import (
	"fmt"
	"gaia/kernel/pkg/core"
	"gaia/kernel/pkg/types"
	"gaia/kernel/pkg/state"
)

func main() {
	config := core.DefaultConfig()
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

	// Coordinator needs a lot of things, but we just want to test checkUsagePolicies.
	// We can't easily call it because it's private.
	// But we can look at how it's implemented and replicate it or use reflection (too complex).
	
	// Actually, let's just use the PolicyEngine directly as that's what it does.
	// But we want to ensure the Coordinator's integration is correct.
	
	// Let's try to run a minimal Coordinator.
	coord := core.NewCoordinator(task, config, nil, nil, nil, nil)
	
	// Since checkUsagePolicies is private, we can't call it from another package.
	// We'll have to rely on the logic in coordinator.go being correct or move the test into the core package.
}
