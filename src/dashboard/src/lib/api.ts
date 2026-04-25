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

// src/lib/api.ts
const KERNEL_URL = process.env.NEXT_PUBLIC_KERNEL_URL || 'http://localhost:8080';

/**
 * fetcher is a generic wrapper around fetch for SWR data fetching.
 */
export async function fetcher(url: string) {
  const res = await fetch(`${KERNEL_URL}${url}`);
  if (!res.ok) {
    throw new Error('Failed to fetch data');
  }
  return res.json();
}

/**
 * approveStep sends a POST request to the Kernel to manually unblock a paused step.
 */
export async function approveStep(taskID: string, stepID: string) {
  const res = await fetch(`${KERNEL_URL}/api/v1/tasks/${taskID}/steps/${stepID}/approve`, {
    method: 'POST',
  });
  if (!res.ok) {
    throw new Error('Failed to approve step');
  }
  return res.json();
}

/**
 * submitTask initiates a new agentic task by sending the user's goal to the Orchestrator.
 */
export async function submitTask(goal: string) {
  const res = await fetch(`${KERNEL_URL}/api/v1/tasks`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ goal }),
  });
  if (!res.ok) {
    throw new Error('Failed to submit task');
  }
  return res.json();
}
