// Copyright 2026 GAIA Contributors
//
// Licensed under the MIT License.
// See the License for the specific language governing permissions and
// limitations under the License.

/**
 * This module implements the primary TypeScript client for the GAIA Kernel.
 * It provides a type-safe abstraction over the REST gateway, enabling
 * goal submission and deterministic task tracking.
 */
import axios, { AxiosInstance } from 'axios';
import * as https from 'https';
import { Task } from './types';

/**
 * TlsConfig provides the necessary certificates for mTLS (Mutual TLS) authentication.
 */
export interface TlsConfig {
  cert: string | Buffer;
  key: string | Buffer;
  ca?: string | Buffer;
  rejectUnauthorized?: boolean;
}

/**
 * ClientConfig provides the configuration options for the GaiaClient.
 */
export interface ClientConfig {
  baseURL?: string;
  tls?: TlsConfig;
  authToken?: string;
}

/**
 * GaiaClient provides a high-level interface for interacting with the GAIA Kernel.
 * It handles task submission, status polling, and agent discovery.
 */
export class GaiaClient {
  private client: AxiosInstance;

  constructor(config: ClientConfig = {}) {
    const baseURL = config.baseURL || 'http://localhost:8080';
    let httpsAgent;

    if (config.tls) {
      httpsAgent = new https.Agent({
        cert: config.tls.cert,
        key: config.tls.key,
        ca: config.tls.ca,
        rejectUnauthorized: config.tls.rejectUnauthorized ?? true,
      });
    }

    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
    };

    if (config.authToken) {
      headers['Authorization'] = `Bearer ${config.authToken}`;
    }

    this.client = axios.create({
      baseURL,
      httpsAgent,
      headers,
    });
  }

  /**
   * submit sends a high-level goal to the GAIA Orchestrator.
   * This initiates the 10-phase control loop, starting with Phase 1 (Submission).
   * @param goal The natural language goal for the agentic system.
   * @returns The newly created Task object.
   */
  async submit(goal: string): Promise<Task> {
    const response = await this.client.post<Task>('/api/v1/tasks', { goal });
    return response.data;
  }

  /**
   * getTask retrieves the current state and plan for a specific task.
   * This is used to monitor the Kernel's progress as it resolves the goal into steps.
   * @param taskID The unique identifier of the task.
   */
  async getTask(taskID: string): Promise<Task> {
    const response = await this.client.get<Task>(`/api/v1/tasks/${taskID}`);
    return response.data;
  }

  /**
   * waitForCompletion polls the kernel until the task reaches a terminal state.
   * This provides a promise-based way to wait for complex, multi-step goals to finish.
   * @param taskID The unique identifier of the task.
   * @param intervalMS Polling interval in milliseconds.
   */
  async waitForCompletion(taskID: string, intervalMS: number = 2000): Promise<Task> {
    let task = await this.getTask(taskID);
    while (task.status === 'pending' || task.status === 'running') {
      await new Promise(resolve => setTimeout(resolve, intervalMS));
      task = await this.getTask(taskID);
    }
    return task;
  }

  /**
   * listAgents retrieves all currently active agents in the GAIA ecosystem.
   */
  async listAgents(): Promise<any[]> {
    const response = await this.client.get('/api/v1/registry/agents');
    return response.data;
  }

  /**
   * listCapabilities retrieves all available tools and skills across all agents.
   */
  async listCapabilities(): Promise<any[]> {
    const response = await this.client.get('/api/v1/registry/capabilities');
    return response.data;
  }

  /**
   * register sends an AgentManifest to the Kernel.
   * @param manifest The agent's authoritative registration record.
   */
  async register(manifest: any): Promise<void> {
    await this.client.post('/api/v1/registry/register', manifest);
  }

  /**
   * deregister removes an agent from the GAIA ecosystem.
   * @param agentID The unique ID of the agent to remove.
   */
  async deregister(agentID: string): Promise<void> {
    await this.client.delete(`/api/v1/registry/agents/${agentID}`);
  }
}
