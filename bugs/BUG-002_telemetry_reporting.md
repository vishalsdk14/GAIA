# BUG: Missing LLM Observability and Mission Telemetry

## Summary
The GAIA Kernel lacks sufficient logging and telemetry for LLM interactions. While the `RequestMetrics` data structure exists, it is not populated during the planning or execution phases. Furthermore, cloud-based LLM calls are completely silent in the logs, and there is no post-mission summary report for the user.

## Environment
- **Workspace**: `GAIA`
- **Component**: `pkg/core/planner.go`, `pkg/core/coordinator.go`

## Steps to Reproduce
1. Configure GAIA to use a Cloud LLM (OpenAI/Anthropic).
2. Submit a goal via `./gaia submit`.
3. Monitor the Kernel logs during the planning phase.
4. Check the task status after completion.

## Actual Results
1. No logs are emitted when the Kernel contacts the Cloud LLM.
2. The `tokens_used` and `duration_ms` fields in the task status remain empty or default to zero.
3. Once a mission finishes, the Kernel simply stops without providing a summary of resources consumed (total steps, total tokens, total cost).

## Expected Results
1. **Audit Logs**: Every LLM contact (local or cloud) should be logged with the request prompt (truncated) and the response status.
2. **Telemetry Extraction**: The Planner should extract token usage from the LLM response (`usage` field in OpenAI/Anthropic/Ollama) and propagate it to the mission state.
3. **Mission Report**: Upon mission completion (Success or Failure), the Kernel should emit a `MISSION_SUMMARY` event containing:
   - Total Steps Executed
   - Total Tokens Consumed (Prompt + Completion)
   - Total Execution Duration
   - Final Outcome

## Root Cause Analysis
- `CloudLLMPlanner.GeneratePlan` in `planner.go` does not contain any `slog` calls.
- The JSON unmarshaling in `planner.go` only captures the `content` of the message, ignoring the `usage` statistics provided by the provider APIs.
- The `Coordinator` does not have a "Phase 11: Final Reporting" logic to aggregate metrics from all steps.

## Proposed Solution
1. **Instrumentation**: Add `slog.Info` and `slog.Debug` calls to both `LocalLLMPlanner` and `CloudLLMPlanner`.
2. **Metric Harvesting**: Update the LLM response structs in `planner.go` to include `Usage` fields and save these to the `Task` metadata.
3. **Report Generator**: Implement a summary function in `coordinator.go` that runs after the DAG execution finishes.
