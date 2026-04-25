# Copyright 2026 GAIA Contributors
#
# Licensed under the MIT License.
# See the License for the specific language governing permissions and
# limitations under the License.

"""
This module implements the core asynchronous client for the GAIA Python SDK.
It provides a high-level abstraction over the GAIA Kernel REST API, enabling
efficient task submission and status monitoring.
"""

import httpx
import asyncio
from typing import Optional, List
from .models.task import Task

class GaiaClient:
    """
    GaiaClient provides an asynchronous interface for interacting with the GAIA Kernel.
    It encapsulates the connection pooling and JSON-to-Model conversion logic,
    ensuring type-safety for Python developers.
    """
    def __init__(self, base_url: str = "http://localhost:8080", cert: Optional[tuple] = None, verify: Optional[str] = None, auth_token: Optional[str] = None):
        self.base_url = base_url
        headers = {}
        if auth_token:
            headers["Authorization"] = f"Bearer {auth_token}"
        
        self.client = httpx.AsyncClient(
            base_url=base_url, 
            cert=cert, 
            verify=verify, 
            headers=headers
        )

    async def submit(self, goal: str) -> Task:
        """
        Submits a high-level goal to the GAIA Orchestrator.
        This triggers Phase 1 (Submission) and Phase 2 (Planning) in the Kernel.
        """
        response = await self.client.post("/api/v1/tasks", json={"goal": goal})
        response.raise_for_status()
        return Task(**response.json())

    async def get_task(self, task_id: str) -> Task:
        """
        Retrieves the current status of a task, including its execution plan.
        This allows clients to track progress as the Kernel moves through the 10-phase loop.
        """
        response = await self.client.get(f"/api/v1/tasks/{task_id}")
        response.raise_for_status()
        return Task(**response.json())

    async def wait_for_completion(self, task_id: str, interval: float = 2.0) -> Task:
        """
        Polls the kernel until the task reaches a terminal state (completed or failed).
        This is a convenience helper for synchronous-style workflows in an async environment.
        """
        task = await self.get_task(task_id)
        while task.status in ["pending", "running"]:
            await asyncio.sleep(interval)
            task = await self.get_task(task_id)
        return task

    async def close(self):
        await self.client.aclose()

    async def list_agents(self) -> List[dict]:
        """
        Retrieves all currently active agents in the GAIA ecosystem.
        """
        response = await self.client.get("/api/v1/registry/agents")
        response.raise_for_status()
        return response.json()

    async def list_capabilities(self) -> List[dict]:
        """
        Retrieves all available tools and skills across all agents.
        """
        response = await self.client.get("/api/v1/registry/capabilities")
        response.raise_for_status()
        return response.json()

    async def register(self, manifest: dict):
        """
        Sends an AgentManifest to the Kernel.
        """
        response = await self.client.post("/api/v1/registry/register", json=manifest)
        response.raise_for_status()

    async def deregister(self, agent_id: str):
        """
        Removes an agent from the GAIA ecosystem.
        """
        response = await self.client.delete(f"/api/v1/registry/agents/{agent_id}")
        response.raise_for_status()
