package core

import (
	"gaia/kernel/pkg/types"
	"testing"
)

// TestCalculateCost verifies the deterministic prefix matching logic and 
// ensuring that the most specific model match is used for pricing.
func TestCalculateCost(t *testing.T) {
	tests := []struct {
		name     string
		model    string
		usage    types.UsageMetrics
		wantCost float64
	}{
		{
			name:  "GPT-4o exact match",
			model: "gpt-4o",
			usage: types.UsageMetrics{PromptTokens: 1000000, CompletionTokens: 0},
			wantCost: 5.00,
		},
		{
			name:  "GPT-4o-mini should match mini not standard 4o",
			model: "gpt-4o-mini-2024-07-18",
			usage: types.UsageMetrics{PromptTokens: 1000000, CompletionTokens: 0},
			wantCost: 0.15,
		},
		{
			name:  "Unknown model returns zero",
			model: "llama3-70b",
			usage: types.UsageMetrics{PromptTokens: 1000, CompletionTokens: 1000},
			wantCost: 0.0,
		},
		{
			name:  "Claude match",
			model: "claude-3-5-sonnet-20240620",
			usage: types.UsageMetrics{PromptTokens: 0, CompletionTokens: 1000000},
			wantCost: 15.00,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateCost(tt.model, tt.usage)
			if got != tt.wantCost {
				t.Errorf("CalculateCost(%s) = %v, want %v", tt.model, got, tt.wantCost)
			}
		})
	}
}
