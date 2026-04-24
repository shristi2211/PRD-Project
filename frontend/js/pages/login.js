// ============================================
// Login Page Component
// ============================================

import { login } from '/js/auth.js';

export function renderLoginPage() {
    return `
        <div class="page">
            <div class="auth-container">
                <div class="glass-card">
                    <div class="brand">
                        <div class="brand-icon">⛳</div>
                        <h1>Golf Score Lottery</h1>
                        <p>Sign in to your account</p>
                    </div>

                    <div id="login-alert" class="alert alert-error" style="display:none;"></div>

                    <form id="login-form" novalidate>
                        <div class="form-group">
                            <label for="login-email">Email Address</label>
                            <input
                                type="email"
                                id="login-email"
                                name="email"
                                placeholder="you@example.com"
                                autocomplete="email"
                                required
                            />
                            <span class="field-error" id="login-email-error"></span>
                        </div>

                        <div class="form-group">
                            <label for="login-password">Password</label>
                            <div class="input-wrapper">
                                <input
                                    type="password"
                                    id="login-password"
                                    name="password"
                                    placeholder="Enter your password"
                                    autocomplete="current-password"
                                    required
                                />
                                <button type="button" class="toggle-password" data-target="login-password" aria-label="Toggle password visibility">👁</button>
                            </div>
                            <span class="field-error" id="login-password-error"></span>
                        </div>

                        <button type="submit" class="btn btn-primary" id="login-submit">
                            <span class="btn-text">Sign In</span>
                        </button>
                    </form>

                    <div class="form-footer">
                        Don't have an account? <a href="#/register">Create one</a>
                    </div>
                </div>
            </div>
        </div>
    `;
}

export function initLoginPage() {
    const form = document.getElementById('login-form');
    const alertEl = document.getElementById('login-alert');
    const submitBtn = document.getElementById('login-submit');
    const emailInput = document.getElementById('login-email');
    const passwordInput = document.getElementById('login-password');

    // Password toggle
    document.querySelectorAll('.toggle-password').forEach(btn => {
        btn.addEventListener('click', () => {
            const input = document.getElementById(btn.dataset.target);
            const isPassword = input.type === 'password';
            input.type = isPassword ? 'text' : 'password';
            btn.textContent = isPassword ? '🙈' : '👁';
        });
    });

    // Clear field errors on input
    [emailInput, passwordInput].forEach(input => {
        input.addEventListener('input', () => {
            input.classList.remove('error');
            const errorEl = document.getElementById(`${input.id}-error`);
            if (errorEl) errorEl.textContent = '';
            alertEl.style.display = 'none';
        });
    });

    form.addEventListener('submit', async (e) => {
        e.preventDefault();

        // Client-side validation
        let valid = true;
        const email = emailInput.value.trim();
        const password = passwordInput.value;

        if (!email) {
            showFieldError('login-email', 'Email is required');
            valid = false;
        } else if (!isValidEmail(email)) {
            showFieldError('login-email', 'Please enter a valid email');
            valid = false;
        }

        if (!password) {
            showFieldError('login-password', 'Password is required');
            valid = false;
        }

        if (!valid) return;

        // Show loading
        setLoading(submitBtn, true);
        alertEl.style.display = 'none';

        try {
            const result = await login(email, password);
            // Regardless of user or admin role, both use the exact same dashboard core
            window.location.hash = '#/dashboard';
        } catch (err) {
            if (err.status === 403) {
                // Subscription required
                alertEl.innerHTML = `⚠ ${err.message}<br><a href="#/" style="color: var(--color-accent); text-decoration: underline; margin-top: 8px; display: inline-block;">→ Go to Plans</a>`;
            } else {
                alertEl.textContent = `⚠ ${err.message}`;
            }
            alertEl.style.display = 'flex';
        } finally {
            setLoading(submitBtn, false);
        }
    });
}

function showFieldError(inputId, message) {
    const input = document.getElementById(inputId);
    const errorEl = document.getElementById(`${inputId}-error`);
    if (input) input.classList.add('error');
    if (errorEl) errorEl.textContent = message;
}

function setLoading(btn, loading) {
    if (loading) {
        btn.disabled = true;
        btn.innerHTML = '<div class="spinner"></div><span>Signing in...</span>';
    } else {
        btn.disabled = false;
        btn.innerHTML = '<span class="btn-text">Sign In</span>';
    }
}

function isValidEmail(email) {
    return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
}
