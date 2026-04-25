---
sidebar_position: 1
---

# Getting Started with GAIA

GAIA is a high-resiliency agent orchestration kernel. This guide will help you set up the GAIA kernel and connect your first agent using our SDKs.

## 1. Prerequisites

- **Go 1.22+** (for the Kernel)
- **Node.js 18+** (for the TypeScript SDK)
- **Python 3.10+** (for the Python SDK)

## 2. Start the GAIA Kernel

First, clone the repository and start the orchestration kernel:

```bash
cd src/kernel
go run main.go
```

The kernel will start an API gateway at `http://localhost:8080`.

## 3. Connect an Agent (Python)

Building a GAIA-compatible agent is easy with the Python SDK. Here is a simple agent with a single capability:

```python
from gaia.agent import GaiaAgent
from gaia.models import AgentManifest

manifest = AgentManifest(
    agent_id="weather-agent",
    version="1.0.0",
    transport="http",
    protocol="native",
    endpoint="http://localhost:9000",
    capabilities=[{
        "name": "get_weather",
        "description": "Returns the current weather for a city"
    }]
)

agent = GaiaAgent(manifest)

@agent.capability("get_weather")
async def handle_weather(city: str):
    return f"The weather in {city} is sunny, 22°C."

if __name__ == "__main__":
    import asyncio
    asyncio.run(agent.start())
```

## 4. Submit a Goal

Once your agent is connected, you can submit goals to the kernel via the TypeScript SDK:

```typescript
import { GaiaClient } from '@gaia-kernel/sdk';

const client = new GaiaClient();

async function main() {
  const task = await client.submit("What is the weather in London?");
  console.log(`Task submitted: ${task.task_id}`);

  const result = await client.waitForCompletion(task.task_id);
  console.log(`Final Result: ${result.status}`);
}

main();
```

## Next Steps

- Explore the [API Reference](/docs/category/api-reference)
- Learn about the [10-Phase Control Loop](/docs/concepts/control-loop)
- Configure [Managed State](/docs/guides/managed-state)
