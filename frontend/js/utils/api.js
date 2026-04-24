// ============================================
// API Client — Fetch wrapper with auto-refresh
// ============================================

const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080';

export async function getSetupStatus() {
    try {
        const response = await fetch(`${API_BASE}/api/auth/setup-status`);
        if (!response.ok) return { setup_complete: true }; // default to safe lock
        const data = await response.json();
        return data.data || data; // Handle JSON wrapper
    } catch {
        return { setup_complete: true };
    }
}

export async function setupAdmin(email, password, name) {
    const response = await fetch(`${API_BASE}/api/auth/setup`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password, name }),
    });

    const data = await response.json();
    if (!response.ok) {
        throw new Error(data.error || 'Admin setup failed');
    }
    return data;
}


/**
 * Makes an authenticated API request.
 * Automatically attaches the access token and handles 401 by refreshing.
 */
export async function apiRequest(endpoint, options = {}) {
    const url = `${API_BASE}${endpoint}`;

    // If we have a refresh token but no access token (e.g. page refresh),
    // proactively obtain a new access token before making the initial request.
    if (!getAccessToken() && getRefreshToken()) {
        await refreshAccessToken();
    }

    const token = getAccessToken();
    const headers = {
        'Content-Type': 'application/json',
        ...options.headers,
    };

    if (token) {
        headers['Authorization'] = `Bearer ${token}`;
    }

    let response = await fetch(url, {
        ...options,
        headers,
    });

    // If 401 even after the initial check (e.g. token expired mid-session)
    if (response.status === 401 && getRefreshToken()) {
        const refreshed = await refreshAccessToken();
        if (refreshed) {
            // Retry the original request with the new token
            headers['Authorization'] = `Bearer ${getAccessToken()}`;
            response = await fetch(url, {
                ...options,
                headers,
            });
        } else {
            // Refresh failed — force logout
            clearTokens();
            window.location.hash = '#/login';
            throw new ApiError('Session expired. Please log in again.', 401);
        }
    }

    const json = await response.json();

    if (!response.ok) {
        throw new ApiError(json.error || 'An unexpected error occurred', response.status);
    }

    // Auto-unwrap the `{ success: true, data: ... }` envelope
    return json.data !== undefined ? json.data : json;
}

/**
 * Refreshes the access token using the refresh token.
 * Returns true if successful, false otherwise.
 */
async function refreshAccessToken() {
    const refreshToken = getRefreshToken();
    if (!refreshToken) return false;

    try {
        const response = await fetch(`${API_BASE}/api/auth/refresh`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ refresh_token: refreshToken }),
        });

        if (!response.ok) {
            clearTokens();
            return false;
        }

        const data = await response.json();
        if (data.success && data.data) {
            setAccessToken(data.data.token);
            setRefreshToken(data.data.refresh_token);
            return true;
        }

        return false;
    } catch {
        clearTokens();
        return false;
    }
}

// ─── Token Storage ────────────────────────────
// Access token kept in memory (XSS-safe), refresh token in localStorage
let _accessToken = null;

export function getAccessToken() {
    return _accessToken;
}

export function setAccessToken(token) {
    _accessToken = token;
}

export function getRefreshToken() {
    try {
        return localStorage.getItem('gsl_refresh_token');
    } catch {
        return null;
    }
}

export function setRefreshToken(token) {
    try {
        localStorage.setItem('gsl_refresh_token', token);
    } catch {
        // localStorage unavailable (private browsing, etc.)
    }
}

export function clearTokens() {
    _accessToken = null;
    try {
        localStorage.removeItem('gsl_refresh_token');
        localStorage.removeItem('gsl_user');
    } catch {
        // ignore
    }
}

// ─── User Cache ───────────────────────────────
export function getCachedUser() {
    try {
        const user = localStorage.getItem('gsl_user');
        return user ? JSON.parse(user) : null;
    } catch {
        return null;
    }
}

export function setCachedUser(user) {
    try {
        localStorage.setItem('gsl_user', JSON.stringify(user));
    } catch {
        // ignore
    }
}

// ─── Custom Error ─────────────────────────────
export class ApiError extends Error {
    constructor(message, status) {
        super(message);
        this.name = 'ApiError';
        this.status = status;
    }
}

// ─── IP Subscription (Phase 6b) ───────────────
export async function submitIpSubscription(plan) {
    return await apiRequest('/api/public/ip-subscribe', {
        method: 'POST',
        body: JSON.stringify({ plan })
    });
}

export async function checkIpSubscription() {
    try {
        const data = await apiRequest('/api/public/ip-status');
        return data.active === true;
    } catch {
        return false;
    }
}
