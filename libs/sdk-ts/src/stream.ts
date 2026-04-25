// Copyright 2026 GAIA Contributors
//
// Licensed under the MIT License.
// See the License for the specific language governing permissions and
// limitations under the License.

/**
 * This module implements the WebSocket stream listener for the GAIA TypeScript SDK.
 * It provides a reactive interface for observing the GAIA 10-phase control loop
 * in real-time as events are emitted by the Kernel.
 */
import { WebSocket } from 'ws';
import { Event } from './types';
import { EventEmitter } from 'events';

/**
 * GaiaStream connects to the GAIA Kernel's real-time event stream.
 * It allows developers to react to step completions, planning updates, and failures
 * without polling the REST API, ensuring low-latency observability.
 */
export class GaiaStream extends EventEmitter {
  private ws: WebSocket | null = null;
  private url: string;

  constructor(baseURL: string = 'ws://localhost:8080') {
    super();
    // Convert http/https to ws/wss if needed
    this.url = baseURL.replace(/^http/, 'ws') + '/api/v1/stream';
  }

  /**
   * connect opens the persistent WebSocket connection.
   */
  connect(): void {
    this.ws = new WebSocket(this.url);

    this.ws.on('open', () => {
      this.emit('connected');
    });

    this.ws.on('message', (data: string) => {
      try {
        const event: Event = JSON.parse(data);
        this.emit('event', event);
        this.emit(event.type, event);
      } catch (err) {
        this.emit('error', new Error('Failed to parse event: ' + err));
      }
    });

    this.ws.on('close', () => {
      this.emit('disconnected');
    });

    this.ws.on('error', (err) => {
      this.emit('error', err);
    });
  }

  /**
   * disconnect closes the WebSocket connection gracefully.
   */
  disconnect(): void {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }
}
