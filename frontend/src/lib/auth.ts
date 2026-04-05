import { discoverAuth, authenticateIdentity } from './api';

export interface AuthSession {
    token: string;
    email: string;
    isBypass?: boolean;
    identity: {
        oktaSub: string;
        damlUserId: string;
        damlPartyId: string;
        email: string;
        displayName: string;
    };
    scopes: string[];
}

export function setSession(session: AuthSession) {
    localStorage.setItem('auth_session', JSON.stringify(session));
}

export function getSession(): AuthSession | null {
    if (typeof window === 'undefined') return null;
    const data = localStorage.getItem('auth_session');
    if (!data) return null;
    try {
        return JSON.parse(data);
    } catch (e) {
        return null;
    }
}

export function clearSession() {
    localStorage.removeItem('auth_session');
    document.cookie = "user_email=; path=/; expires=Thu, 01 Jan 1970 00:00:00 UTC;";
    document.cookie = "user_scopes=; path=/; expires=Thu, 01 Jan 1970 00:00:00 UTC;";
    document.cookie = "admin-mode=; path=/; expires=Thu, 01 Jan 1970 00:00:00 UTC;";
}

export function hasScope(scope: string): boolean {
    const session = getSession();
    return session?.scopes.includes(scope) || false;
}

export function setRedirect(url: string) {
    if (url.includes('/login') || url.includes('/callback')) return;
    sessionStorage.setItem('post_login_redirect', url);
}

export function getRedirect(): string {
    const url = sessionStorage.getItem('post_login_redirect');
    sessionStorage.removeItem('post_login_redirect');
    return url || '/';
}

export async function handleLoginDiscovery(email: string) {
    const provider = await discoverAuth(email);
    
    // Save intended email
    localStorage.setItem('pending_email', email);

    if (provider.type === 'OIDC') {
        const clientId = '0oa11kvi7mxaHbQaV698'; // Okta Client ID
        const issuer = provider.issuer;
        const redirectUri = window.location.origin + '/callback';
        const scope = 'openid profile email escrow:read escrow:write escrow:accept system:admin';
        
        // Authorization Code Flow with PKCE (simplified for this walkthrough)
        const authUrl = `${issuer}/v1/authorize?client_id=${clientId}&response_type=token&scope=${encodeURIComponent(scope)}&redirect_uri=${encodeURIComponent(redirectUri)}&state=xyz&nonce=abc`;
        
        window.location.href = authUrl;
    } else if (provider.type === 'SAML') {
        window.location.href = provider.loginUrl;
    }
}

export async function loginAsRole(role: string) {
    // Map roles to specific emails
    const roleMap: Record<string, string> = {
        'Buyer': 'joey@buyer.com',
        'Seller': 'jimmy@seller.com',
        'EscrowMediator': 'sally@mediator.com',
        'CentralBank': 'bob@banker.com'
    };
    
    const email = roleMap[role] || `${role.toLowerCase()}@dev.local`;
    const devToken = `dev-token-${role}`;
    
    // Set partial session so getAuthHeaders includes X-Dev-User
    const partialSession = {
        token: devToken,
        isBypass: true,
        identity: { displayName: role, email: email }
    };
    localStorage.setItem('auth_session', JSON.stringify(partialSession));

    try {
        const identity = await authenticateIdentity(devToken);
        
        const session: AuthSession = {
            token: devToken,
            email: identity.email,
            isBypass: true,
            identity,
            scopes: ['escrow:read', 'escrow:write', 'escrow:accept', 'system:admin']
        };

        setSession(session);
        document.cookie = `user_email=${identity.email}; path=/; max-age=3600`;
        document.cookie = `user_scopes=${session.scopes.join(',')}; path=/; max-age=3600`;
        
        return session;
    } catch (err) {
        localStorage.removeItem('auth_session');
        throw err;
    }
}

export async function finalizeAuthentication(token: string) {
    const identity = await authenticateIdentity(token);
    
    // Extract scopes from JWT (only if it's a real 3-part token)
    let scopes: string[] = [];
    const parts = token.split('.');
    if (parts.length === 3) {
        try {
            const payload = JSON.parse(atob(parts[1]));
            scopes = payload.scp || [];
        } catch (e) {
            console.warn('Failed to parse JWT scopes, using defaults');
        }
    } else {
        // Default scopes for bypass/dummy tokens
        scopes = ['escrow:read', 'escrow:write', 'escrow:accept', 'system:admin'];
    }

    const session: AuthSession = {
        token,
        email: identity.email,
        identity,
        scopes
    };

    setSession(session);
    
    // Set cookies for Astro SSR compatibility
    document.cookie = `user_email=${identity.email}; path=/; max-age=3600`;
    document.cookie = `user_scopes=${scopes.join(',')}; path=/; max-age=3600`;
    
    return session;
}
