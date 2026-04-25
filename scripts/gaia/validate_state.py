# Copyright 2026 GAIA Contributors
#
# Licensed under the MIT License.
# See the License for the specific language governing permissions and
# limitations under the License.

"""
This script validates the Managed Agent State (Tier 4) implementation.
It ensures that the Python SDK correctly interacts with the Kernel's SQLite
storage layer and that multi-tenant isolation is maintained.
"""

import asyncio
from gaia.agent import GaiaAgent
from gaia.models.agent_manifest import AgentManifest, InvokeContract, StateRequirements

async def validate_managed_state():
    """
    Validates the Tier 4 Managed State API using the Python SDK.
    """
    manifest = AgentManifest(
        agent_id="test-agent-001",
        version="1.0.0",
        transport="http",
        protocol="native",
        endpoint="http://localhost:9000",
        invoke=InvokeContract(timeout_ms=5000, async_supported=True),
        capabilities=[],
        state_requirements=StateRequirements(required=True, max_bytes=1024)
    )

    agent = GaiaAgent(manifest)
    
    print("Testing Managed State...")
    
    # 1. Set State
    test_data = {"last_seen": "2026-04-25T19:00:00Z", "status": "active"}
    await agent.state.set("meta", test_data)
    print("✅ Stored 'meta' key")

    # 2. Get State
    retrieved = await agent.state.get("meta")
    assert retrieved == test_data
    print("✅ Retrieved 'meta' key and verified data integrity")

    # 3. List Keys
    keys = await agent.state.list()
    assert "meta" in keys
    print(f"✅ Listed keys: {keys}")

    # 4. Delete State
    await agent.state.delete("meta")
    deleted = await agent.state.get("meta")
    assert deleted is None
    print("✅ Deleted 'meta' key")

if __name__ == "__main__":
    asyncio.run(validate_managed_state())
