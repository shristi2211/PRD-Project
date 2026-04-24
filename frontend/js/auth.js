// ============================================
// Auth Module — Login, Register, Logout, State
// ============================================

import {
    apiRequest,
    setAccessToken,
    setRefreshToken,
    getRefreshToken,
    clearTokens,
    getCachedUser,
    setCachedUser,
    getAccessToken,
    getSetupStatus,
    setupAdmin as apiSetupAdmin
} from './utils/api.js';

export async function checkSetupStatus() {
    return await getSetupStatus();
}

export async function setupSystemAdmin(email, password, name) {
    return await apiSetupAdmin(email, password, name);
}

/**
 * Register a new user.
 * @returns {Object} User data on success
 */
export async function register(email, password, name) {
    const data = await apiRequest('/api/auth/register', {
        method: 'POST',
        body: JSON.stringify({ email, password, name }),
    });
    return data;
}

/**
 * Login with email and password.
 * Stores tokens and user data.
 * @returns {Object} { token, refresh_token, user }
 */
export async function login(email, password) {
    const data = await apiRequest('/api/auth/login', {
        method: 'POST',
        body: JSON.stringify({ email, password }),
    });

    const { token, refresh_token, user } = data;

    setAccessToken(token);
    setRefreshToken(refresh_token);
    setCachedUser(user);

    return data;
}

/**
 * Logout the current user.
 */
export async function logout() {
    const refreshToken = getRefreshToken();

    try {
        if (refreshToken) {
            await apiRequest('/api/auth/logout', {
                method: 'POST',
                body: JSON.stringify({ refresh_token: refreshToken }),
            });
        }
    } catch {
        // Silently fail — we'll clear local state regardless
    } finally {
        clearTokens();
    }
}

/**
 * Get the current user profile from the server.
 * @returns {Object} User data
 */
export async function getCurrentUser() {
    const data = await apiRequest('/api/users/me');
    setCachedUser(data);
    return data;
}

/**
 * Check if the user is currently authenticated.
 * @returns {boolean}
 */
export function isAuthenticated() {
    return !!(getAccessToken() || getRefreshToken());
}

/**
 * Get cached user or null.
 * @returns {Object|null}
 */
export function getUser() {
    return getCachedUser();
}
