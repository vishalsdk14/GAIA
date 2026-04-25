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

package policy

import (
	"fmt"
	"gaia/kernel/pkg/types"

	"github.com/google/cel-go/cel"
)

// PolicyEngine implements Phase 5 (Policy Evaluation) using Google's CEL.
type PolicyEngine struct {
	env *cel.Env
}

// NewPolicyEngine initializes a CEL environment with the standard GAIA variable declarations.
func NewPolicyEngine() (*PolicyEngine, error) {
	e, err := cel.NewEnv(
		cel.Variable("task", cel.MapType(cel.StringType, cel.AnyType)),
		cel.Variable("step", cel.MapType(cel.StringType, cel.AnyType)),
		cel.Variable("agent", cel.MapType(cel.StringType, cel.AnyType)),
		cel.Variable("capability", cel.MapType(cel.StringType, cel.AnyType)),
		cel.Variable("usage", cel.MapType(cel.StringType, cel.AnyType)),
		cel.Variable("cost", cel.MapType(cel.StringType, cel.AnyType)),
		cel.Variable("env", cel.StringType),
	)
	if err != nil {
		return nil, fmt.Errorf("policy: failed to initialize CEL env: %w", err)
	}
	return &PolicyEngine{env: e}, nil
}

// Evaluate runs a CEL expression against the provided context.
// Returns true if the policy is satisfied, false if denied.
func (pe *PolicyEngine) Evaluate(policy string, context map[string]interface{}) (bool, error) {
	ast, issues := pe.env.Compile(policy)
	if issues != nil && issues.Err() != nil {
		return false, fmt.Errorf("policy: compilation error: %w", issues.Err())
	}

	program, err := pe.env.Program(ast)
	if err != nil {
		return false, fmt.Errorf("policy: program initialization error: %w", err)
	}

	out, _, err := program.Eval(context)
	if err != nil {
		return false, fmt.Errorf("policy: evaluation error: %w", err)
	}

	result, ok := out.Value().(bool)
	if !ok {
		return false, fmt.Errorf("policy: expression must return a boolean")
	}

	return result, nil
}

// EvaluateAll runs a list of policies and returns the first failure, if any.
func (pe *PolicyEngine) EvaluateAll(policies []string, context map[string]interface{}) error {
	for _, p := range policies {
		success, err := pe.Evaluate(p, context)
		if err != nil {
			return err
		}
		if !success {
			return fmt.Errorf("policy: %w: denied by rule: %s", fmt.Errorf(string(types.ErrorCodePolicyDenied)), p)
		}
	}
	return nil
}
