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

// Package policy implements the security firewall for the GAIA kernel.
// It leverages Google's Common Expression Language (CEL) to evaluate
// dynamic, non-Turing complete security rules in microseconds.
package policy

import (
	"fmt"

	"github.com/google/cel-go/cel"
)

// Engine encapsulates the CEL execution environment.
// It allows the kernel to compile and evaluate dynamic string-based policies
// against the runtime context of Tasks and Steps.
type Engine struct {
	env *cel.Env
}

// NewEngine initializes the CEL environment and declares the canonical
// variables that policy rules are allowed to reference.
// We use cel.DynType to allow flexible JSON-like traversal of our schemas.
func NewEngine() (*Engine, error) {
	env, err := cel.NewEnv(
		cel.Variable("task", cel.DynType),
		cel.Variable("step", cel.DynType),
		cel.Variable("agent", cel.DynType),
		cel.Variable("capability", cel.DynType),
		cel.Variable("env", cel.StringType),
	)
	if err != nil {
		return nil, fmt.Errorf("policy: failed to init CEL env: %w", err)
	}

	return &Engine{env: env}, nil
}

// Evaluate compiles and executes a CEL rule string against a provided data context.
// The context must provide the variables declared in NewEngine.
// It returns a boolean indicating whether the policy passed, or an error if the
// rule is malformed or evaluation fails.
//
// Note: In Phase 4 (Production), the Abstract Syntax Trees (ASTs) generated
// by Compile() should be cached in an LRU cache to prevent recompilation
// overhead on hot-path evaluations.
func (e *Engine) Evaluate(rule string, context map[string]interface{}) (bool, error) {
	// 1. Compile the string rule into an AST
	ast, issues := e.env.Compile(rule)
	if issues != nil && issues.Err() != nil {
		return false, fmt.Errorf("policy: failed to compile rule: %w", issues.Err())
	}

	// 2. Generate the executable program
	prg, err := e.env.Program(ast)
	if err != nil {
		return false, fmt.Errorf("policy: failed to build program: %w", err)
	}

	// 3. Evaluate the program against the contextual state
	out, _, err := prg.Eval(context)
	if err != nil {
		return false, fmt.Errorf("policy: failed to evaluate rule: %w", err)
	}

	// 4. Ensure the output is strictly a boolean
	result, ok := out.Value().(bool)
	if !ok {
		return false, fmt.Errorf("policy: rule evaluation did not return a boolean type")
	}

	return result, nil
}
