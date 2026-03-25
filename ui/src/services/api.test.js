import { afterEach, describe, expect, it, vi } from 'vitest'
import { api } from './api'

describe('api', () => {
  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('sets auth header and serializes request body', async () => {
    const fetchMock = vi.spyOn(globalThis, 'fetch').mockResolvedValue({
      ok: true,
      json: vi.fn().mockResolvedValue({ ok: true }),
    })

    const res = await api('POST', '/auth/login', { email: 'a@b.c' }, 'jwt-token')

    expect(res).toEqual({ ok: true })
    expect(fetchMock).toHaveBeenCalledWith('/auth/login', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: 'Bearer jwt-token',
      },
      body: JSON.stringify({ email: 'a@b.c' }),
    })
  })

  it('returns empty object when response body is not JSON', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue({
      ok: true,
      json: vi.fn().mockRejectedValue(new Error('invalid json')),
    })

    const res = await api('GET', '/health')
    expect(res).toEqual({})
  })

  it('returns parsed error payload for non-2xx responses', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue({
      ok: false,
      json: vi.fn().mockResolvedValue({ error: 'invalid credentials' }),
    })

    const res = await api('POST', '/auth/login', { email: 'bad@x' })
    expect(res).toEqual({ error: 'invalid credentials' })
  })
})
