// Copyright 2026 GAIA Contributors
//
// Licensed under the MIT License.
// See the License for the specific language governing permissions and
// limitations under the License.

/**
 * This module implements the GaiaAgent helper class for the TypeScript SDK.
 * It provides a simplified interface for building GAIA-compatible agents,
 * managing the handshake and Tier 4 state isolation.
 */
import axios, { AxiosInstance } from 'axios';
import { AgentManifest, Request, Response } from './types';
import { GaiaClient } from './client';

/**
 * GaiaAgent simplifies the process of building and connecting an agent to the GAIA Kernel.
 * It abstracts the complexity of registration and provides a secure, partitioned
 * managed state API for the agent to persist its context.
 */
export class GaiaAgent {
  private manifest: AgentManifest;
  private kernelURL: string;
  private stateClient: AxiosInstance;
  private gaia: GaiaClient;

  constructor(manifest: AgentManifest, kernelURL: string = 'http://localhost:8080') {
    this.manifest = manifest;
    this.kernelURL = kernelURL;
    this.gaia = new GaiaClient(kernelURL);
    this.stateClient = axios.create({
      baseURL: kernelURL,
      headers: {
        'Content-Type': 'application/json',
        'X-Agent-ID': manifest.agent_id,
      },
    });
  }

  /**
   * state provides access to the Managed Agent State (Tier 4).
   * This state is persisted in the Kernel's SQLite database and is strictly
   * isolated to this agent's identity via the X-Agent-ID header.
   */
  get state() {
    return {
      get: async <T>(key: string): Promise<T | null> => {
        try {
          const resp = await this.stateClient.get(`/internal/v1/state/${key}`);
          return resp.data as T;
        } catch (err: any) {
          if (err.response?.status === 404) return null;
          throw err;
        }
      },
      set: async (key: string, data: any): Promise<void> => {
        await this.stateClient.put(`/internal/v1/state/${key}`, data);
      },
      delete: async (key: string): Promise<void> => {
        await this.stateClient.delete(`/internal/v1/state/${key}`);
      },
      list: async (): Promise<string[]> => {
        const resp = await this.stateClient.get('/internal/v1/state');
        return resp.data.keys;
      }
    };
  }

  /**
   * start registers the agent with the kernel and begins listening for requests.
   * In Phase 7, this handles the foundational handshake; Phase 8 will implement
   * the persistent bi-directional transport for the Native protocol.
   */
  async start(): Promise<void> {
    console.log(`GAIA Agent [${this.manifest.agent_id}] registering...`);
    await this.gaia.register(this.manifest);
    console.log(`GAIA Agent [${this.manifest.agent_id}] active.`);
    // TODO: Implement WebSocket listener for inbound requests from Kernel
  }

  /**
   * stop gracefully deregisters the agent from the Kernel.
   */
  async stop(): Promise<void> {
    console.log(`GAIA Agent [${this.manifest.agent_id}] deregistering...`);
    await this.gaia.deregister(this.manifest.agent_id);
  }
}
