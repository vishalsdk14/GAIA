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
	"fmt"
	"gaia/kernel/pkg/core"
	"gaia/kernel/pkg/logger"
)

func main() {
	// Initialize default configuration
	config := core.DefaultConfig()

	// Initialize structured logging
	logger.Init(config.LogLevel)

	logger.L.Info("GAIA Orchestration Kernel initializing...", 
		"version", "0.1.0-alpha",
		"log_level", config.LogLevel,
		"db_path", config.DBPath,
	)

	fmt.Println("GAIA Kernel Core is ready.")
	
	// TODO: Initialize Registry, HTTP Server, and Control Loop Hub
	logger.L.Info("Kernel ready for task submission.")
}
