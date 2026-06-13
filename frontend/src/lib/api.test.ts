import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { fetchEscrows, discoverAuth, saveConfig, fetchConfig } from './api';

describe('API Client Tests', () => {
  const mockFetch = vi.fn();

  beforeEach(() => {
    vi.stubGlobal('fetch', mockFetch);
    vi.stubGlobal('window', {});
    
    // Setup local storage session bypass mock
    const session = {
      isBypass: true,
      identity: {
        email: 'depositor@bank.com',
        damlPartyId: 'Depositor::123'
      },
      token: 'dummy-token'
    };
    vi.stubGlobal('localStorage', {
      getItem: vi.fn().mockReturnValue(JSON.stringify(session)),
      setItem: vi.fn(),
      removeItem: vi.fn(),
      clear: vi.fn()
    });
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('should fetch active escrows using getAuthHeaders and resolveApiPath', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => [{ id: 'escrow-1', state: 'ACTIVE' }]
    });

    const escrows = await fetchEscrows();
    expect(escrows).toEqual([{ id: 'escrow-1', state: 'ACTIVE' }]);
    expect(mockFetch).toHaveBeenLastCalledWith(
      expect.stringContaining('/depositor/api/v1/escrows'),
      expect.objectContaining({
        headers: expect.objectContaining({
          'Authorization': 'Bearer bypass-token-depositor@bank.com',
          'X-Dev-User': 'depositor@bank.com'
        })
      })
    );
  });

  it('should discover auth provider via POST method and body parameter (no URL query)', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({ provider: 'google', bypass: true })
    });

    const res = await discoverAuth('user@bank.com');
    expect(res).toEqual({ provider: 'google', bypass: true });
    expect(mockFetch).toHaveBeenLastCalledWith(
      expect.stringContaining('/api/v1/auth/discover'),
      expect.objectContaining({
        method: 'POST',
        headers: expect.objectContaining({
          'Content-Type': 'application/json'
        }),
        body: JSON.stringify({ email: 'user@bank.com' })
      })
    );
  });

  it('should save user configs via POST method and body payload (no URL query)', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true
    });

    await saveConfig('depositor@bank.com', 'stablecoin_provider', 'bitgo');
    expect(mockFetch).toHaveBeenLastCalledWith(
      expect.stringContaining('/depositor/api/v1/config'),
      expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({ key: 'stablecoin_provider', value: 'bitgo' })
      })
    );
  });

  it('should fetch config using GET method and load key from response (no URL query)', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({ stablecoin_provider: 'bitgo', okta_org: 'dev-123' })
    });

    const val = await fetchConfig('depositor@bank.com', 'stablecoin_provider');
    expect(val).toBe('bitgo');
    expect(mockFetch).toHaveBeenLastCalledWith(
      expect.stringContaining('/depositor/api/v1/config'),
      expect.objectContaining({
        headers: expect.objectContaining({
          'Authorization': 'Bearer bypass-token-depositor@bank.com'
        })
      })
    );
  });
});
