// Copyright 2026 GAIA Contributors
//
// Licensed under the MIT License.
package core

import (
	"gaia/kernel/pkg/types"
	"log/slog"
	"strings"
)

// ModelPricing defines the cost structure for a specific LLM model.
// Prices are defined per 1,000,000 (one million) tokens to avoid floating point precision issues
// with extremely small per-token numbers.
type ModelPricing struct {
	InputPricePer1M  float64
	OutputPricePer1M float64
}

// PricingTable serves as the single source of truth for LLM costs.
// Update this table when providers change their pricing.
var PricingTable = map[string]ModelPricing{
	"gpt-4o": {
		InputPricePer1M:  5.00,
		OutputPricePer1M: 15.00,
	},
	"gpt-4o-mini": {
		InputPricePer1M:  0.15,
		OutputPricePer1M: 0.60,
	},
	"gpt-3.5-turbo": {
		InputPricePer1M:  0.50,
		OutputPricePer1M: 1.50,
	},
	"claude-3-5-sonnet-20240620": {
		InputPricePer1M:  3.00,
		OutputPricePer1M: 15.00,
	},
}

// CalculateCost computes the estimated USD cost for a given interaction.
// It uses case-insensitive prefix matching to handle model versioning (e.g., 'gpt-4o-2024-05-13').
func CalculateCost(model string, usage types.UsageMetrics) float64 {
	var pricing ModelPricing
	found := false

	// Search for the best match in the pricing table
	for modelPrefix, p := range PricingTable {
		if strings.HasPrefix(strings.ToLower(model), strings.ToLower(modelPrefix)) {
			pricing = p
			found = true
			break
		}
	}

	// If the model is not found, or it's a local model (e.g., 'llama3', 'mistral'), cost is 0.
	if !found {
		// Log a warning if it's a known cloud provider prefix but missing from our table
		if strings.Contains(model, "gpt-") || strings.Contains(model, "claude-") {
			slog.Warn("Pricing: Unknown cloud model detected, using 0.0 cost", "model", model)
		}
		return 0.0
	}

	// Constants to avoid magic numbers in the calculation
	const million = 1000000.0
	
	inputCost := (float64(usage.PromptTokens) / million) * pricing.InputPricePer1M
	outputCost := (float64(usage.CompletionTokens) / million) * pricing.OutputPricePer1M

	return inputCost + outputCost
}
