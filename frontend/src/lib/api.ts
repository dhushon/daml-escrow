const API_BASE = 'http://localhost:8081/api/v1';

export interface Milestone {
    label: string;
    amount: number;
    completed: boolean;
}

export interface EscrowResponse {
    id: string;
    buyer: string;
    seller: string;
    issuer: string;
    mediator: string;
    amount: number;
    currency: string;
    state: string;
    currentMilestoneIndex: number;
    milestones: Milestone[];
}

export async function fetchEscrows(user: string = 'Buyer') {
    const response = await fetch(`${API_BASE}/escrows?user=${user}`);
    if (!response.ok) throw new Error('Failed to fetch escrows');
    return response.json();
}

export async function fetchProposals(user: string = 'Buyer') {
    const response = await fetch(`${API_BASE}/escrows?user=${user}&tab=proposals`); // Adjusted to use common endpoint if needed, or update backend
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

export async function fetchEscrowLifecycle(id: string, user: string = 'Buyer') {
    const response = await fetch(`${API_BASE}/escrows/${id}/lifecycle?user=${user}`);
    if (!response.ok) throw new Error('Failed to fetch lifecycle');
    return response.json();
}

// Lifecycle Actions (Phase 5 High-Assurance)

export async function proposeEscrow(req: any) {
    const response = await fetch(`${API_BASE}/escrows/propose`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(req)
    });
    if (!response.ok) {
        const err = await response.json();
        throw new Error(err.message || 'Failed to propose escrow');
    }
    return response.json();
}

export async function fundEscrow(id: string, custodyRef: string, user: string) {
    const response = await fetch(`${API_BASE}/escrows/${id}/fund?user=${user}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ custodyRef })
    });
    if (!response.ok) throw new Error('Failed to fund escrow');
}

export async function activateEscrow(id: string, user: string) {
    const response = await fetch(`${API_BASE}/escrows/${id}/activate?user=${user}`, {
        method: 'POST'
    });
    if (!response.ok) throw new Error('Failed to activate escrow');
}

export async function confirmConditions(id: string, user: string) {
    const response = await fetch(`${API_BASE}/escrows/${id}/confirm?user=${user}`, {
        method: 'POST'
    });
    if (!response.ok) throw new Error('Failed to confirm conditions');
}

export async function raiseDispute(id: string, user: string) {
    const response = await fetch(`${API_BASE}/escrows/${id}/dispute?user=${user}`, {
        method: 'POST'
    });
    if (!response.ok) throw new Error('Failed to raise dispute');
}

export async function ratifySettlement(id: string, user: string) {
    const response = await fetch(`${API_BASE}/escrows/${id}/ratify?user=${user}`, {
        method: 'POST'
    });
    if (!response.ok) throw new Error('Failed to ratify settlement');
}

export async function finalizeSettlement(id: string, user: string) {
    const response = await fetch(`${API_BASE}/escrows/${id}/finalize?user=${user}`, {
        method: 'POST'
    });
    if (!response.ok) throw new Error('Failed to finalize settlement');
}

// Invitation Actions

export async function createInvitation(inviterId: string, email: string, role: string, inviteeType: string, asset: any, terms: any) {
    const response = await fetch(`${API_BASE}/invites?user=${inviterId}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            inviterId,
            inviteeEmail: email,
            inviteeRole: role,
            inviteeType,
            asset,
            terms
        })
    });
    if (!response.ok) {
        const err = await response.json();
        throw new Error(err.message || 'Failed to create invitation');
    }
    return response.json();
}

export async function fetchInvitationByToken(token: string) {
    const response = await fetch(`${API_BASE}/invites/token/${token}`);
    if (!response.ok) throw new Error('Invitation not found or expired');
    return response.json();
}

export async function claimInvitation(inviteId: string, token: string, user: string) {
    const response = await fetch(`${API_BASE}/invites/token/${token}/claim?user=${user}`, {
        method: 'POST'
    });
    if (!response.ok) throw new Error('Failed to claim invitation');
    return response.json();
}

// Identity

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
