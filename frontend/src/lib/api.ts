const API_BASE = 'http://localhost:8081/api/v1';

function getAuthHeaders() {
    const data = localStorage.getItem('auth_session');
    if (!data) return { 'Content-Type': 'application/json' };
    const session = JSON.parse(data);
    
    const headers: Record<string, string> = {
        'Content-Type': 'application/json'
    };

    if (session.isBypass) {
        headers['X-Dev-User'] = session.identity.displayName;
        // Backend middleware will skip verification if isDevBypass is true
        headers['Authorization'] = `Bearer bypass-token-${session.identity.displayName}`;
    } else {
        headers['Authorization'] = `Bearer ${session.token}`;
    }

    return headers;
}

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

export async function fetchEscrows() {
    const response = await fetch(`${API_BASE}/escrows`, { headers: getAuthHeaders() });
    if (!response.ok) throw new Error('Failed to fetch escrows');
    return response.json();
}

export async function fetchProposals() {
    const response = await fetch(`${API_BASE}/escrows?tab=proposals`, { headers: getAuthHeaders() });
    if (!response.ok) throw new Error('Failed to fetch proposals');
    return response.json();
}

export async function fetchSettlements() {
    const response = await fetch(`${API_BASE}/settlements`, { headers: getAuthHeaders() });
    if (!response.ok) throw new Error('Failed to fetch settlements');
    return response.json();
}

export async function fetchWallets() {
    const response = await fetch(`${API_BASE}/wallets`, { headers: getAuthHeaders() });
    if (!response.ok) throw new Error('Failed to fetch wallets');
    return response.json();
}

export async function fetchMetrics() {
    const response = await fetch(`${API_BASE}/metrics`, { headers: getAuthHeaders() });
    if (!response.ok) throw new Error('Failed to fetch metrics');
    return response.json();
}

export async function fetchEscrowLifecycle(id: string) {
    const response = await fetch(`${API_BASE}/escrows/${id}/lifecycle`, { headers: getAuthHeaders() });
    if (!response.ok) throw new Error('Failed to fetch lifecycle');
    return response.json();
}

export async function discoverAuth(email: string) {
    const response = await fetch(`${API_BASE}/auth/discover?email=${encodeURIComponent(email)}`);
    if (!response.ok) throw new Error('Failed to discover auth provider');
    return response.json();
}

// Lifecycle Actions (Phase 5 High-Assurance)

export async function proposeEscrow(req: any) {
    const response = await fetch(`${API_BASE}/escrows/propose`, {
        method: 'POST',
        headers: getAuthHeaders(),
        body: JSON.stringify(req)
    });
    if (!response.ok) {
        const err = await response.json();
        throw new Error(err.message || 'Failed to propose escrow');
    }
    return response.json();
}

export async function fundEscrow(id: string, custodyRef: string) {
    const response = await fetch(`${API_BASE}/escrows/${id}/fund`, {
        method: 'POST',
        headers: getAuthHeaders(),
        body: JSON.stringify({ custodyRef })
    });
    if (!response.ok) throw new Error('Failed to fund escrow');
}

export async function activateEscrow(id: string) {
    const response = await fetch(`${API_BASE}/escrows/${id}/activate`, {
        method: 'POST',
        headers: getAuthHeaders()
    });
    if (!response.ok) throw new Error('Failed to activate escrow');
}

export async function confirmConditions(id: string) {
    const response = await fetch(`${API_BASE}/escrows/${id}/confirm`, {
        method: 'POST',
        headers: getAuthHeaders()
    });
    if (!response.ok) throw new Error('Failed to confirm conditions');
}

export async function raiseDispute(id: string) {
    const response = await fetch(`${API_BASE}/escrows/${id}/dispute`, {
        method: 'POST',
        headers: getAuthHeaders()
    });
    if (!response.ok) throw new Error('Failed to raise dispute');
}

export async function ratifySettlement(id: string) {
    const response = await fetch(`${API_BASE}/escrows/${id}/ratify`, {
        method: 'POST',
        headers: getAuthHeaders()
    });
    if (!response.ok) throw new Error('Failed to ratify settlement');
}

export async function finalizeSettlement(id: string) {
    const response = await fetch(`${API_BASE}/escrows/${id}/finalize`, {
        method: 'POST',
        headers: getAuthHeaders()
    });
    if (!response.ok) throw new Error('Failed to finalize settlement');
}

// Invitation Actions

export async function createInvitation(email: string, role: string, inviteeType: string, asset: any, terms: any) {
    const response = await fetch(`${API_BASE}/invites`, {
        method: 'POST',
        headers: getAuthHeaders(),
        body: JSON.stringify({
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

export async function claimInvitation(token: string) {
    const response = await fetch(`${API_BASE}/invites/token/${token}/claim`, {
        method: 'POST',
        headers: getAuthHeaders()
    });
    if (!response.ok) throw new Error('Failed to claim invitation');
    return response.json();
}

export async function fetchHealth() {
    const response = await fetch(`${API_BASE}/health`);
    if (!response.ok) throw new Error('Failed to fetch health');
    return response.json();
}

// Identity

export async function authenticateIdentity(jwt: string) {
    const response = await fetch(`${API_BASE}/auth/me`, {
        method: 'GET',
        headers: getAuthHeaders()
    });
    
    if (!response.ok) {
        throw new Error('Failed to authenticate and provision identity');
    }
    return response.json();
}
