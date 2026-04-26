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
	"github.com/joho/godotenv"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
)

func main() {
	// Load environment variables from .env file (Phase 11: Developer Experience)
	if err := godotenv.Load(); err != nil {
		// We don't fatal here because .env is optional in production (env vars take precedence)
		log.Println("Note: No .env file found, using system environment variables.")
	}

	// Initialize default configuration
	config := core.DefaultConfig()

	// Initialize structured logging
	logger.Init(config.LogLevel)

	// 5. Initialize Secret Registry (Phase 8) - Moved up for Audit & Encryption
	secrets := core.NewSecretRegistry()

	// Initialize tamper-proof Audit Logger (Phase 10: HMAC-SHA256)
	auditSecret, err := secrets.GetSecret("AUDIT_SECRET")
	if err != nil {
		log.Fatalf("CRITICAL: AUDIT_SECRET is required but not set in the Secret Registry. Failed to initialize tamper-proof audit log: %v", err)
	}
	if len(auditSecret) < 32 {
		log.Fatalf("CRITICAL: AUDIT_SECRET is too weak. HMAC-SHA256 requires at least 32 bytes of entropy for Enterprise Governance.")
	}

	if al, err := logger.InitAuditLogger(config.AuditLogPath, []byte(auditSecret)); err != nil {
		log.Fatalf("CRITICAL: Failed to initialize audit log: %v\n", err)
	} else {
		// Phase 10: Verify chain integrity on startup
		if err := al.VerifyChain(); err != nil {
			log.Fatalf("CRITICAL: Audit log integrity check failed: %v. The system may have been tampered with.", err)
		}
		logger.L.Info("Audit log integrity verified (HMAC-SHA256)")
	}

	logger.L.Info("GAIA Orchestration Kernel initializing...", 
		"version", "0.1.0-alpha",
		"log_level", config.LogLevel,
		"db_path", config.DBPath,
	)

	// Phase 11: Performance Profiling
	if config.EnablePerformanceProfiling {
		go func() {
			logger.L.Info("Performance profiling enabled", "port", "6060")
			if err := http.ListenAndServe("localhost:6060", nil); err != nil {
				logger.L.Error("pprof server failed", "error", err)
			}
		}()
	}

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

	// 6. Initialize & Start API Gateway
	server := api.NewServer(orch, reg, store)

	// Phase 8: Encryption at Rest
	// Use the SecretRegistry to fetch the master encryption key.
	// Support both raw bytes (32-char) and hex-encoded (64-char) keys.
	if encKey, err := secrets.GetSecret("ENCRYPTION_KEY"); err == nil {
		var keyBytes []byte
		if len(encKey) == 64 {
			// Assume hex encoding for 64-character strings
			if enc, err := state.NewEncryptorFromHex(encKey); err == nil {
				if err := store.EnableEncryption(enc.Key()); err == nil {
					logger.L.Info("Encryption at Rest enabled (Hex key)")
				}
			}
		} else {
			keyBytes = []byte(encKey)
			if err := store.EnableEncryption(keyBytes); err != nil {
				log.Fatalf("Critical: Failed to enable encryption: %v", err)
			}
			logger.L.Info("Encryption at Rest enabled (Raw key)")
		}
	}

	// Phase 8: Configure Security Modes & JWT
	server.AuthMode = getEnv("GAIA_AUTH_MODE", "legacy")
	
	// JWT Configuration (Standard/Strict Mode)
	if jwtSecret, err := secrets.GetSecret("JWT_SECRET"); err == nil {
		server.JWTSecret = []byte(jwtSecret)
	} else if server.AuthMode == "standard" {
		log.Fatalf("Critical: GAIA_AUTH_MODE=standard requires GAIA_JWT_SECRET")
	}

	server.CACertPath = getEnv("GAIA_CA_CERT", "./certs/ca.crt")
	server.ServerCertPath = getEnv("GAIA_SERVER_CERT", "./certs/server.crt")
	server.ServerKeyPath = getEnv("GAIA_SERVER_KEY", "./certs/server.key")
	
	port := getEnv("GAIA_PORT", ":8080")
	logger.L.Info("Kernel Gateway starting", 
		"addr", port, 
		"auth_mode", server.AuthMode,
	)
	
	if err := server.Start(port); err != nil {
		log.Fatalf("Critical: API Server failed: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
