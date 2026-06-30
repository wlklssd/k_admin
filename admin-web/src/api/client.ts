const API_BASE = import.meta.env.VITE_API_BASE || '';

export interface ApiEnvelope<T> {
  code: number;
  message?: string;
  msg?: string;
  data: T;
}

export async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers);
  const hasBody = init.body !== undefined && init.body !== null;

  if (hasBody && !(init.body instanceof FormData) && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json');
  }

  const token = localStorage.getItem('pezmax_admin_token');
  if (token && token !== 'demo-token') {
    headers.set('Authorization', `Bearer ${token}`);
  }

  const response = await fetch(`${API_BASE}${path}`, {
    ...init,
    headers,
  });
  const payload = await response.json().catch(() => null);

  if (!response.ok) {
    throw new Error(payload?.message || payload?.msg || response.statusText || '请求失败');
  }

  if (payload && typeof payload === 'object' && 'code' in payload) {
    const envelope = payload as ApiEnvelope<T>;
    if (envelope.code !== 0) {
      throw new Error(envelope.message || envelope.msg || '请求失败');
    }
    return envelope.data;
  }

  return payload as T;
}
