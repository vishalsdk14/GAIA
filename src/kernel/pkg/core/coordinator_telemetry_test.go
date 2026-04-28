package core

import (
	"gaia/kernel/pkg/types"
	"testing"
)

// TestCoordinator_SyncMetadata verifies that the syncMetadata helper correctly
// mirrors the task's primary telemetry fields into the Metadata map.
func TestCoordinator_SyncMetadata(t *testing.T) {
	task := &types.Task{
		TokensPrompt:     100,
		TokensCompletion: 200,
		EstimatedCostUSD: 0.05,
	}
	
	c := &Coordinator{
		task: task,
	}
	
	c.syncMetadata()
	
	usage, ok := task.Metadata["usage"].(map[string]interface{})
	if !ok {
		t.Fatal("Metadata['usage'] not found or not a map")
	}
	
	if usage["prompt_tokens"] != 100 {
		t.Errorf("Expected prompt_tokens 100, got %v", usage["prompt_tokens"])
	}
	if usage["total_tokens"] != 300 {
		t.Errorf("Expected total_tokens 300, got %v", usage["total_tokens"])
	}
	
	cost, ok := task.Metadata["cost"].(map[string]interface{})
	if !ok {
		t.Fatal("Metadata['cost'] not found or not a map")
	}
	
	if cost["usd"] != 0.05 {
		t.Errorf("Expected cost usd 0.05, got %v", cost["usd"])
	}
}
