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

package core

import (
	"fmt"
	"gaia/kernel/pkg/types"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WSTransport implements the WebSocket-based bi-directional transport layer.
type WSTransport struct {
	mu    sync.Mutex
	conns map[string]*websocket.Conn
}

// NewWSTransport initializes the WebSocket connection pool.
func NewWSTransport() *WSTransport {
	return &WSTransport{
		conns: make(map[string]*websocket.Conn),
	}
}

// Dispatch pushes a request frame over a persistent WebSocket connection.
func (t *WSTransport) Dispatch(req *types.Request, agent *types.AgentManifest) (*types.Response, error) {
	if agent.Endpoint == "" {
		return nil, fmt.Errorf("transport: ws agent has no endpoint")
	}

	conn, err := t.getConn(agent.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("transport: ws connection failed: %w", err)
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	// 1. Send Request Frame
	if err := conn.WriteJSON(req); err != nil {
		return nil, fmt.Errorf("transport: ws write failed: %w", err)
	}

	// 2. Read Response Frame (Sync mode)
	// Note: In a production kernel, we would use a correlation map to handle
	// async frames, but here we implement the basic sync request-response over WS.
	var resp types.Response
	if err := conn.ReadJSON(&resp); err != nil {
		return nil, fmt.Errorf("transport: ws read failed: %w", err)
	}

	return &resp, nil
}

func (t *WSTransport) getConn(endpoint string) (*websocket.Conn, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if conn, exists := t.conns[endpoint]; exists {
		return conn, nil
	}

	dialer := websocket.DefaultDialer
	conn, _, err := dialer.Dial(endpoint, nil)
	if err != nil {
		return nil, err
	}

	t.conns[endpoint] = conn

	// Spawn Keep-Alive (Ping/Pong) goroutine
	go t.keepAlive(endpoint, conn)

	return conn, nil
}

func (t *WSTransport) keepAlive(endpoint string, conn *websocket.Conn) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		t.mu.Lock()
		if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
			delete(t.conns, endpoint)
			conn.Close()
			t.mu.Unlock()
			return
		}
		t.mu.Unlock()
	}
}
