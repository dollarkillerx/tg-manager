const API_BASE = import.meta.env.VITE_API_BASE || '';

let idCounter = 0;

export async function rpc<T>(method: string, params?: Record<string, unknown>): Promise<T> {
  const id = String(++idCounter);
  const response = await fetch(`${API_BASE}/api/rpc`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      jsonrpc: '2.0',
      method,
      params: params ?? {},
      id,
    }),
  });

  const data = await response.json();

  if (data.error) {
    throw new Error(data.error.message || 'RPC Error');
  }

  return data.result as T;
}
