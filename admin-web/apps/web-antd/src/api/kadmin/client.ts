import { useAccessStore } from '@vben/stores';

const API_BASE = import.meta.env.VITE_GLOB_API_URL || '/api';

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

  const token = useAccessStore().accessToken;
  if (token) {
    headers.set('Authorization', `Bearer ${token}`);
  }

  const apiPath = path.startsWith('/api/') ? path.slice(4) : path;
  const response = await fetch(`${API_BASE}${apiPath}`, {
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
