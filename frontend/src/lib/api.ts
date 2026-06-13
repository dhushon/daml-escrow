// High-Assurance API Client
// authoritatively manages tripartite routing across local and GKE perimeters.

const DEFAULT_BASE = 'http://localhost:8080'; // Unified Gateway Port
const API_URL = (typeof import.meta !== 'undefined' && import.meta.env?.PUBLIC_API_URL) || DEFAULT_BASE;

function getAuthHeaders(email?: string, role?: string) {
    if (typeof window === 'undefined') {
        const headers: Record<string, string> = {
            'Content-Type': 'application/json'
        };
        if (email) {
            headers['X-Dev-User'] = email;
            headers['Authorization'] = `Bearer bypass-token-${email}`;
        }
        if (role) {
            headers['X-Assumed-Role'] = role;
        }
        return headers;
    }

    const data = localStorage.getItem('auth_session');
    if (!data) return { 'Content-Type': 'application/json' };
    const session = JSON.parse(data);

    const headers: Record<string, string> = {
        'Content-Type': 'application/json'
    };

    if (session.isBypass) {
        headers['X-Dev-User'] = session.identity.email; // Use email as the identifier
        headers['Authorization'] = `Bearer bypass-token-${session.identity.email}`;
    } else {
        headers['Authorization'] = `Bearer ${session.token}`;
    }

    if (session.assumedRole) {
        headers['X-Assumed-Role'] = session.assumedRole;
    }

    return headers;
}

/**
 * resolveApiPath authoritatively selects the correct Gateway path based on the user's institutional role.
 */
function resolveApiPath(path: string, email?: string): string {
    // Standalone Mode Detection: If port 8081 is used, we are hitting the API directly
    const isStandalone = API_URL.includes(':8081');

    if (typeof window === 'undefined') {
        let rolePath = 'bank'; // Default fallback
        const effEmail = email?.toLowerCase() || '';
        if (effEmail.includes('depositor')) {
            rolePath = 'depositor';
        } else if (effEmail.includes('beneficiary')) {
            rolePath = 'beneficiary';
        } else if (effEmail.includes('bank')) {
            rolePath = 'bank';
        }
        return isStandalone ? `${API_URL}/api/v1${path}` : `${API_URL}/${rolePath}/api/v1${path}`;
    }

    const data = localStorage.getItem('auth_session');
    if (!data || isStandalone) return `${API_URL}/api/v1${path}`;

    const session = JSON.parse(data);
    const effEmail = session.identity?.email?.toLowerCase() || '';
    let rolePath = 'bank'; // Default fallback

    if (effEmail.includes('depositor')) {
        rolePath = 'depositor';
    } else if (effEmail.includes('beneficiary')) {
        rolePath = 'beneficiary';
    } else if (effEmail.includes('bank')) {
        rolePath = 'bank';
    }

    return `${API_URL}/${rolePath}/api/v1${path}`;
}

export interface Milestone {
    label: string;
    amount: number;
    completed: boolean;
}

export interface EscrowResponse {
    id: string;
    depositor: string;
    beneficiary: string;
    issuer: string;
    mediator: string;
    amount: number;
    currency: string;
    state: string;
    currentMilestoneIndex: number;
    milestones: Milestone[];
}

export async function fetchEscrows(email?: string, role?: string) {
    const response = await fetch(resolveApiPath('/escrows', email), { headers: getAuthHeaders(email, role) });
    if (!response.ok) throw new Error('Failed to fetch escrows');
    return response.json();
}

export async function fetchProposals(email?: string, role?: string) {
    const response = await fetch(resolveApiPath('/escrows?tab=proposals', email), { headers: getAuthHeaders(email, role) });
    if (!response.ok) throw new Error('Failed to fetch proposals');
    return response.json();
}

export async function fetchSettlements() {
    const response = await fetch(resolveApiPath('/settlements'), { headers: getAuthHeaders() });
    if (!response.ok) throw new Error('Failed to fetch settlements');
    return response.json();
}

export async function fetchWallets() {
    const response = await fetch(resolveApiPath('/wallets'), { headers: getAuthHeaders() });
    if (!response.ok) throw new Error('Failed to fetch wallets');
    return response.json();
}

export async function fetchMetrics() {
    const response = await fetch(resolveApiPath('/metrics'), { headers: getAuthHeaders() });
    if (!response.ok) throw new Error('Failed to fetch metrics');
    return response.json();
}

export async function fetchEscrowLifecycle(id: string) {
    const response = await fetch(resolveApiPath(`/escrows/${id}/lifecycle`), { headers: getAuthHeaders() });
    if (!response.ok) throw new Error('Failed to fetch lifecycle');
    return response.json();
}

export async function fetchIdentities() {
    const response = await fetch(resolveApiPath('/identities'), { headers: getAuthHeaders() });
    if (!response.ok) throw new Error('Failed to fetch identities');
    return response.json();
}

// --- Phase 11: Draft & Negotiation ---

export async function saveDraft(req: any) {
    const response = await fetch(resolveApiPath('/drafts'), {
        method: 'POST',
        headers: getAuthHeaders(),
        body: JSON.stringify(req)
    });
    if (!response.ok) {
        const err = await response.json();
        throw new Error(err.error || 'Failed to save draft agreement');
    }
    return response.json();
}

export async function fetchDrafts(email?: string, role?: string) {
    const response = await fetch(resolveApiPath('/drafts', email), { headers: getAuthHeaders(email, role) });
    if (!response.ok) throw new Error('Failed to fetch draft agreements');
    return await response.json();
}

export async function fetchDraft(draftID: string) {
    const response = await fetch(resolveApiPath(`/drafts/${draftID}`), { headers: getAuthHeaders() });
    if (!response.ok) throw new Error('Failed to fetch draft agreement');
    return await response.json();
}

export async function amendDraft(draftID: string, req: any) {
    const response = await fetch(resolveApiPath(`/drafts/${draftID}/amend`), {
        method: 'POST',
        headers: getAuthHeaders(),
        body: JSON.stringify(req)
    });
    if (!response.ok) {
        const err = await response.json();
        throw new Error(err.error || 'Failed to amend draft');
    }
    return await response.json();
}

export async function approveDraft(draftID: string) {
    const response = await fetch(resolveApiPath(`/drafts/${draftID}/approve`), {
        method: 'POST',
        headers: getAuthHeaders()
    });
    if (!response.ok) throw new Error('Failed to approve draft');
    return true;
}

export async function promoteDraftToLedger(draftID: string) {

    const response = await fetch(resolveApiPath(`/drafts/${draftID}/promote`), {
        method: 'POST',
        headers: getAuthHeaders()
    });
    if (!response.ok) throw new Error('Failed to promote draft to ledger');
}

export async function ingestContract(files: File[]) {
    const formData = new FormData();
    files.forEach(file => {
        formData.append('agreement', file);
    });

    const headers = getAuthHeaders();
    // Remove Content-Type to let browser set boundary
    delete headers['Content-Type'];

    const response = await fetch(resolveApiPath('/ingest'), {
        method: 'POST',
        headers: headers,
        body: formData
    });
    if (!response.ok) {
        const err = await response.json();
        throw new Error(err.error || 'Failed to ingest contract');
    }
    return response.json();
}

export async function discoverAuth(email: string) {
    const response = await fetch(resolveApiPath('/auth/discover'), {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email })
    });
    if (!response.ok) throw new Error('Failed to discover auth provider');
    return response.json();
}

// Lifecycle Actions (Phase 5 High-Assurance)

export async function proposeEscrow(req: any) {
    const response = await fetch(resolveApiPath('/escrows/propose'), {
        method: 'POST',
        headers: getAuthHeaders(),
        body: JSON.stringify(req)
    });
    if (!response.ok) {
        const err = await response.json();
        throw new Error(err.error || 'Failed to propose escrow');
    }
    return response.json();
}

export async function fundEscrow(id: string, custodyRef: string) {
    const response = await fetch(resolveApiPath(`/escrows/${id}/fund`), {
        method: 'POST',
        headers: getAuthHeaders(),
        body: JSON.stringify({ custodyRef })
    });
    if (!response.ok) throw new Error('Failed to fund escrow');
}

export async function activateEscrow(id: string) {
    const response = await fetch(resolveApiPath(`/escrows/${id}/activate`), {
        method: 'POST',
        headers: getAuthHeaders()
    });
    if (!response.ok) throw new Error('Failed to activate escrow');
}

export async function confirmConditions(id: string) {
    const response = await fetch(resolveApiPath(`/escrows/${id}/confirm`), {
        method: 'POST',
        headers: getAuthHeaders()
    });
    if (!response.ok) throw new Error('Failed to confirm conditions');
}

export async function raiseDispute(id: string) {
    const response = await fetch(resolveApiPath(`/escrows/${id}/dispute`), {
        method: 'POST',
        headers: getAuthHeaders()
    });
    if (!response.ok) throw new Error('Failed to raise dispute');
}

export async function ratifySettlement(id: string) {
    const response = await fetch(resolveApiPath(`/escrows/${id}/ratify`), {
        method: 'POST',
        headers: getAuthHeaders()
    });
    if (!response.ok) throw new Error('Failed to ratify settlement');
}

export async function finalizeSettlement(id: string) {
    const response = await fetch(resolveApiPath(`/escrows/${id}/finalize`), {
        method: 'POST',
        headers: getAuthHeaders()
    });
    if (!response.ok) throw new Error('Failed to finalize settlement');
}

// Invitation Actions

export async function createInvitation(email: string, role: string, inviteeType: string, asset: any, terms: any) {
    const response = await fetch(resolveApiPath('/invites'), {
        method: 'POST',
        headers: getAuthHeaders(),
        body: JSON.stringify({
            inviteeEmail: email,
            inviteeRole: role,
            inviteeType,
            assetType: asset.assetType,
            assetId: asset.assetId,
            amount: asset.amount,
            currency: asset.currency,
            conditionDescription: terms.conditionDescription,
            conditionType: terms.conditionType || 'Default',
            evidenceRequired: terms.evidenceRequired || 'Default',
            expiryDate: terms.expiryDate,
            gracePeriodDays: terms.gracePeriodDays || 0,
            disputeWindowDays: terms.disputeWindowDays || 0
        })
    });
    if (!response.ok) {
        const err = await response.json();
        throw new Error(err.error || 'Failed to create invitation');
    }
    return response.json();
}

export async function fetchInvitationByToken(token: string) {
    const response = await fetch(resolveApiPath(`/invites/token/${token}`));
    if (!response.ok) throw new Error('Invitation not found or expired');
    return response.json();
}

export async function claimInvitation(token: string) {
    const response = await fetch(resolveApiPath(`/invites/token/${token}/claim`), {
        method: 'POST',
        headers: getAuthHeaders()
    });
    if (!response.ok) throw new Error('Failed to claim invitation');
    return response.json();
}

export async function fetchHealth() {
    const response = await fetch(resolveApiPath('/health'));
    if (!response.ok) throw new Error('Failed to fetch health');
    return response.json();
}

export async function fetchConfig(user: string, key: string) {
    const response = await fetch(resolveApiPath('/config'), {
        headers: getAuthHeaders()
    });
    if (response.status === 404) return null;
    if (!response.ok) throw new Error('Failed to fetch config');
    const configs = await response.json();
    return configs ? configs[key] : null;
}

export async function saveConfig(user: string, key: string, value: any) {
    const response = await fetch(resolveApiPath('/config'), {
        method: 'POST',
        headers: getAuthHeaders(),
        body: JSON.stringify({ key, value })
    });
    if (!response.ok) throw new Error('Failed to save config');
}

// Identity

export async function authenticateIdentity(jwt: string) {
    const headers = getAuthHeaders();
    if (jwt) {
        headers['Authorization'] = `Bearer ${jwt}`;
    }
    if (typeof window !== 'undefined') {
        const assumedRole = localStorage.getItem('pending_assumed_role');
        if (assumedRole) {
            headers['X-Assumed-Role'] = assumedRole;
        }
    }

    const response = await fetch(resolveApiPath('/auth/me'), {
        method: 'GET',
        headers
    });

    if (!response.ok) {
        const error = await response.json();
        throw new Error(`Auth failed (${response.status}): ${error.error || 'Unknown error'}`);
    }
    return response.json();
}

export async function fetchNonce() {
    const response = await fetch(resolveApiPath('/auth/nonce'));
    if (!response.ok) throw new Error('Failed to fetch authentication nonce');
    return response.json();
}

export async function verifyWallet(req: { nonce: string; signature: string; publicKey: string; damlPartyId: string; assumedRole?: string }) {
    const response = await fetch(resolveApiPath('/auth/wallet/verify'), {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(req)
    });
    if (!response.ok) {
        const err = await response.json();
        throw new Error(err.error || 'Wallet verification failed');
    }
    return response.json();
}
