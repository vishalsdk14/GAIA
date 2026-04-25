# Copyright 2026 GAIA Contributors
#
# Licensed under the MIT License.
# See the License for the specific language governing permissions and
# limitations under the License.

"""
This module implements the GaiaAgent helper class for the Python SDK.
It provides a high-level, decorator-based interface for building GAIA-compatible
agents, hiding the complexity of manifest generation and state management.
"""

import httpx
import asyncio
from typing import Callable, Any, Dict, Optional, List
from .models.agent_manifest import AgentManifest
from .client import GaiaClient

class GaiaAgent:
    """
    GaiaAgent simplifies building GAIA-compatible agents in Python.
    It provides decorators for registering capabilities and helpers for state management,
    ensuring that agents can focus on business logic while the SDK handles the
    underlying transport and protocol compliance.
    """
    def __init__(self, manifest: AgentManifest, kernel_url: str = "http://localhost:8080"):
        self.manifest = manifest
        self.kernel_url = kernel_url
        self.gaia = GaiaClient(kernel_url)
        self.capabilities: Dict[str, Callable] = {}
        self.state_client = httpx.AsyncClient(
            base_url=kernel_url,
            headers={"X-Agent-ID": manifest.agent_id}
        )

    def capability(self, name: str):
        """
        Decorator to register a function as a GAIA capability.
        The function's signature and docstring should ideally match the 
        capability definition in the agent manifest.
        """
        def decorator(func: Callable):
            self.capabilities[name] = func
            return func
        return decorator

    @property
    def state(self):
        """
        Managed Agent State (Tier 4) accessor.
        Provides a simplified interface for persistent key-value storage
        that is strictly isolated to this agent's namespace in the Kernel.
        """
        class StateProxy:
            def __init__(self, client: httpx.AsyncClient):
                self.client = client

            async def get(self, key: str) -> Optional[Any]:
                resp = await self.client.get(f"/internal/v1/state/{key}")
                if resp.status_code == 404:
                    return None
                resp.raise_for_status()
                return resp.json()

            async def set(self, key: str, value: Any):
                resp = await self.client.put(f"/internal/v1/state/{key}", json=value)
                resp.raise_for_status()

            async def delete(self, key: str):
                resp = await self.client.delete(f"/internal/v1/state/{key}")
                resp.raise_for_status()

            async def list(self) -> List[str]:
                resp = await self.client.get("/internal/v1/state")
                resp.raise_for_status()
                return resp.json().get("keys", [])

        return StateProxy(self.state_client)

    async def start(self):
        """
        Starts the agent and begins listening for requests from the GAIA Kernel.
        In Phase 7, this handles the foundational handshake; Phase 8 will add 
        persistent WebSocket listener logic for the Native protocol.
        """
        print(f"GAIA Agent [{self.manifest.agent_id}] registering...")
        await self.gaia.register(self.manifest.dict(exclude_none=True))
        print(f"GAIA Agent [{self.manifest.agent_id}] active with {len(self.capabilities)} capabilities.")
        # TODO: Implement WebSocket listener for bi-directional Native protocol
        try:
            while True:
                await asyncio.sleep(1)
        except asyncio.CancelledError:
            await self.stop()

    async def stop(self):
        """
        Gracefully deregisters the agent from the Kernel.
        """
        print(f"GAIA Agent [{self.manifest.agent_id}] deregistering...")
        await self.gaia.deregister(self.manifest.agent_id)
        await self.state_client.aclose()
        await self.gaia.close()
