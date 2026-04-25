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

package main

import (
	"gaia/kernel/pkg/api"
	"gaia/kernel/pkg/core"
	"gaia/kernel/pkg/logger"
	"gaia/kernel/pkg/registry"
	"gaia/kernel/pkg/state"
	"log"
)

func main() {
	// Initialize default configuration
	config := core.DefaultConfig()

	// Initialize structured logging
	logger.Init(config.LogLevel)

	// Initialize tamper-proof Audit Logger
	if _, err := logger.InitAuditLogger(config.AuditLogPath); err != nil {
		log.Printf("Warning: Failed to initialize audit log: %v\n", err)
	}

	logger.L.Info("GAIA Orchestration Kernel initializing...", 
		"version", "0.1.0-alpha",
		"log_level", config.LogLevel,
		"db_path", config.DBPath,
	)

	// 1. Initialize State Storage (Tier 2 & 4)
	store, err := state.NewSQLiteStore(config.DBPath)
	if err != nil {
		log.Fatalf("Critical: Failed to initialize SQLite store: %v", err)
	}
	taskStore, _ := state.NewTaskStore(store.DB)

	// 2. Initialize Registry
	reg := registry.NewInMemoryRegistry()

	// 3. Initialize Planner & Transport
	planner, err := core.NewPlanner(config)
	if err != nil {
		log.Fatalf("Critical: Failed to initialize planner: %v", err)
	}
	transport := core.NewProtocolDispatcher() 

	// 4. Initialize Orchestrator (Goal Manager)
	orch := core.NewOrchestrator(config, reg, taskStore, planner, transport)

	// 5. Initialize & Start API Gateway
	server := api.NewServer(orch, reg)
	
	port := ":8080"
	logger.L.Info("Kernel Gateway starting", "addr", port)
	if err := server.Start(port); err != nil {
		log.Fatalf("Critical: API Server failed: %v", err)
	}
}
