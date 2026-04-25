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

package api

import (
	"gaia/kernel/pkg/common"
	"gaia/kernel/pkg/logger"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // For dev purposes
}

func (s *Server) handleStream(w http.ResponseWriter, r *http.Request) {
	taskID := r.URL.Query().Get("task_id")
	
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.L.Error("WebSocket upgrade failed", "error", err)
		return
	}
	defer conn.Close()

	// Subscribe to the global Event Bus
	bus := common.GetEventBus()
	eventChan := bus.Subscribe()

	logger.L.Info("Client connected to event stream", "task_id", taskID)

	for event := range eventChan {
		// Filter by taskID if provided
		if taskID != "" && event.TaskID != taskID {
			continue
		}

		// Push event to WebSocket
		if err := conn.WriteJSON(event); err != nil {
			logger.L.Warn("WebSocket write failed, closing stream", "error", err)
			break
		}
	}
}
