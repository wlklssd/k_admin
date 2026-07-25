import { createPinia, setActivePinia } from 'pinia';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { useAccessStore } from '@vben/stores';

import { request } from './client';

function jsonResponse(body: unknown, status = 200) {
  return new Response(JSON.stringify(body), {
    headers: { 'Content-Type': 'application/json' },
    status,
  });
}

describe('kadmin request client', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('refreshes access token and retries once after 401', async () => {
    const accessStore = useAccessStore();
    accessStore.setAccessToken('old-access-token');
    accessStore.setRefreshToken('old-refresh-token');

    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(jsonResponse({ code: 1, message: 'invalid token' }, 401))
      .mockResolvedValueOnce(
        jsonResponse({
          code: 0,
          data: {
            accessToken: 'new-access-token',
            refreshToken: 'new-refresh-token',
          },
        }),
      )
      .mockResolvedValueOnce(jsonResponse({ code: 0, data: { ok: true } }));
    vi.stubGlobal('fetch', fetchMock);

    await expect(request<{ ok: boolean }>('/api/logs')).resolves.toEqual({
      ok: true,
    });

    expect(fetchMock).toHaveBeenCalledTimes(3);
    expect(fetchMock.mock.calls[1]?.[0]).toBe('/api/auth/refresh');
    expect(accessStore.accessToken).toBe('new-access-token');
    expect(accessStore.refreshToken).toBe('new-refresh-token');
    expect(
      (fetchMock.mock.calls[2]?.[1]?.headers as Headers).get('Authorization'),
    ).toBe('Bearer new-access-token');
  });
});
