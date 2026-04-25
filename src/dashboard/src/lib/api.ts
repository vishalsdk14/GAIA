// src/lib/api.ts
const KERNEL_URL = process.env.NEXT_PUBLIC_KERNEL_URL || 'http://localhost:8080';

export async function fetcher(url: string) {
  const res = await fetch(`${KERNEL_URL}${url}`);
  if (!res.ok) {
    throw new Error('Failed to fetch data');
  }
  return res.json();
}

export async function approveStep(taskID: string, stepID: string) {
  const res = await fetch(`${KERNEL_URL}/api/v1/tasks/${taskID}/steps/${stepID}/approve`, {
    method: 'POST',
  });
  if (!res.ok) {
    throw new Error('Failed to approve step');
  }
  return res.json();
}

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
