const API_BASE = 'http://localhost:8081/api/v1';

export async function fetchEscrows(user: string = 'Buyer') {
    const response = await fetch(`${API_BASE}/escrows?user=${user}`);
    if (!response.ok) throw new Error('Failed to fetch escrows');
    return response.json();
}

export async function fetchProposals(user: string = 'Buyer') {
    const response = await fetch(`${API_BASE}/escrows/proposals?user=${user}`);
    if (!response.ok) throw new Error('Failed to fetch proposals');
    return response.json();
}

export async function fetchSettlements() {
    const response = await fetch(`${API_BASE}/settlements`);
    if (!response.ok) throw new Error('Failed to fetch settlements');
    return response.json();
}

export async function fetchWallets(user: string = 'Buyer') {
    const response = await fetch(`${API_BASE}/wallets?user=${user}`);
    if (!response.ok) throw new Error('Failed to fetch wallets');
    return response.json();
}

export async function fetchMetrics(user: string = 'CentralBank') {
    const response = await fetch(`${API_BASE}/metrics?user=${user}`);
    if (!response.ok) throw new Error('Failed to fetch metrics');
    return response.json();
}

export async function fetchConfig(user: string, key: string) {
    const response = await fetch(`${API_BASE}/config?user=${user}&key=${key}`);
    if (!response.ok) {
        if (response.status === 404) return null;
        throw new Error('Failed to fetch config');
    }
    return response.json();
}

export async function saveConfig(user: string, key: string, value: any) {
    const response = await fetch(`${API_BASE}/config?user=${user}&key=${key}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(value)
    });
    if (!response.ok) throw new Error('Failed to save config');
}

export async function createEscrow(req: any) {
    const response = await fetch(`${API_BASE}/escrows`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(req)
    });
    if (!response.ok) throw new Error('Failed to create escrow');
    return response.json();
}

export async function proposeEscrow(req: any) {
    const response = await fetch(`${API_BASE}/escrows/propose`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(req)
    });
    if (!response.ok) throw new Error('Failed to propose escrow');
    return response.json();
}

export async function acceptProposal(id: string, user: string) {
    const response = await fetch(`${API_BASE}/escrows/${id}/accept?user=${user}`, {
        method: 'POST'
    });
    if (!response.ok) throw new Error('Failed to accept proposal');
    return response.json();
}

export async function releaseFunds(id: string) {
    const response = await fetch(`${API_BASE}/escrows/${id}/release`, {
        method: 'POST'
    });
    if (!response.ok) throw new Error('Failed to release funds');
}

export async function raiseDispute(id: string) {
    const response = await fetch(`${API_BASE}/escrows/${id}/resolve`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ payoutToBuyer: 0, payoutToSeller: 0 }) // Mock dispute start
    });
    if (!response.ok) throw new Error('Failed to raise dispute');
}

export async function settlePayment(id: string) {
    const response = await fetch(`${API_BASE}/settlements/${id}/settle`, {
        method: 'POST'
    });
    if (!response.ok) throw new Error('Failed to settle payment');
}

export async function createInvitation(inviterId: string, email: string, role: string, inviteeType: string, terms: any) {
    const response = await fetch(`${API_BASE}/invites?user=${inviterId}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            inviterId,
            inviteeEmail: email,
            inviteeRole: role,
            inviteeType,
            terms
        })
    });
    if (!response.ok) {
        const err = await response.json();
        throw new Error(err.message || 'Failed to create invitation');
    }
    return response.json();
}

export async function authenticateIdentity(jwt: string) {
    const response = await fetch(`${API_BASE}/auth/me`, {
        method: 'GET',
        headers: { 
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${jwt}`
        }
    });
    
    if (!response.ok) {
        throw new Error('Failed to authenticate and provision identity');
    }
    return response.json();
}
