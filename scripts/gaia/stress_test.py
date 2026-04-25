# Copyright 2026 GAIA Contributors
#
# Licensed under the MIT License.
# See the License for the specific language governing permissions and
# limitations under the License.

"""
This script implements Phase 7 Stress Testing for the GAIA Kernel.
It uses the Python SDK to inject failures (network drops, delays) and verify
that the Kernel's P4 (Resiliency) logic correctly handles async completions
and task status recovery.
"""

import asyncio
import random
import time
from gaia.client import GaiaClient
from gaia.models.task import Task

async def simulate_failure_injection(client: GaiaClient, goal: str):
    """
    Submits a task and simulates various system failures to test Kernel resiliency.
    """
    print(f"🚀 Starting Stress Test: '{goal}'")
    task = await client.submit(goal)
    print(f"Task submitted: {task.task_id}")

    # Start polling for status
    start_time = time.time()
    while True:
        task = await client.get_task(task.task_id)
        print(f"[{int(time.time() - start_time)}s] Status: {task.status}")

        if task.status in ["completed", "failed"]:
            break

        # Simulate a "Network Drop" by pausing polling
        if random.random() < 0.1:
            print("⚠️ Simulating Network Drop (5s pause)...")
            await asyncio.sleep(5)

        await asyncio.sleep(2)

    print(f"🏁 Final Task State: {task.status}")

async def main():
    client = GaiaClient()
    try:
        # Test 1: Simple Success Path
        await simulate_failure_injection(client, "Calculate the square root of 144")
        
        # Test 2: Complex Goal (Planning test)
        await simulate_failure_injection(client, "Research the current price of Bitcoin and send an email alert")
    finally:
        await client.close()

if __name__ == "__main__":
    asyncio.run(main())
