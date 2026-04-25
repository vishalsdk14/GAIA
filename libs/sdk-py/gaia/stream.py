# Copyright 2026 GAIA Contributors
#
# Licensed under the MIT License.
# See the License for the specific language governing permissions and
# limitations under the License.

"""
This module implements the WebSocket stream listener for the GAIA Python SDK.
It provides an asynchronous, event-driven interface for observing Kernel events.
"""

import json
import asyncio
import websockets
from typing import Callable, Dict, List, Any

class GaiaStream:
    """
    GaiaStream connects to the GAIA Kernel's real-time event stream.
    It allows developers to register callbacks for specific event types.
    """
    def __init__(self, base_url: str = "http://localhost:8080"):
        self.url = base_url.replace("http", "ws") + "/api/v1/stream"
        self.callbacks: Dict[str, List[Callable]] = {}
        self.ws = None

    def on(self, event_type: str, callback: Callable):
        """
        Registers a callback for a specific event type.
        """
        if event_type not in self.callbacks:
            self.callbacks[event_type] = []
        self.callbacks[event_type].append(callback)

    async def listen(self):
        """
        Connects to the stream and begins processing messages.
        """
        async with websockets.connect(self.url) as ws:
            self.ws = ws
            async for message in ws:
                try:
                    event = json.loads(message)
                    event_type = event.get("type")
                    
                    # Trigger specific callbacks
                    if event_type in self.callbacks:
                        for cb in self.callbacks[event_type]:
                            if asyncio.iscoroutinefunction(cb):
                                await cb(event)
                            else:
                                cb(event)
                    
                    # Trigger catch-all callbacks
                    if "*" in self.callbacks:
                        for cb in self.callbacks["*"]:
                            if asyncio.iscoroutinefunction(cb):
                                await cb(event)
                            else:
                                cb(event)
                                
                except Exception as e:
                    print(f"Error processing stream event: {e}")

    async def stop(self):
        if self.ws:
            await self.ws.close()
