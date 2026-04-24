// ============================================
// Register Page Component
// ============================================

import { register } from '/js/auth.js';

export function renderRegisterPage() {
    return `
        <div class="page">
            <div class="auth-container">
                <div class="glass-card">
                    <div class="brand">
                        <div class="brand-icon">⛳</div>
                        <h1>Golf Score Lottery</h1>
                        <p>Create your account</p>
                    </div>

                    <div id="register-alert" class="alert" style="display:none;"></div>

                    <form id="register-form" novalidate>
                        <div class="form-group">
                            <label for="register-name">Full Name</label>
                            <input
                                type="text"
                                id="register-name"
                                name="name"
                                placeholder="John Doe"
                                autocomplete="name"
                                required
                            />
                            <span class="field-error" id="register-name-error"></span>
                        </div>

                        <div class="form-group">
                            <label for="register-email">Email Address</label>
                            <input
                                type="email"
                                id="register-email"
                                name="email"
                                placeholder="you@example.com"
                                autocomplete="email"
                                required
                            />
                            <span class="field-error" id="register-email-error"></span>
                        </div>

                        <div class="form-group">
                            <label for="register-password">Password</label>
                            <div class="input-wrapper">
                                <input
                                    type="password"
                                    id="register-password"
                                    name="password"
                                    placeholder="Min 8 characters"
                                    autocomplete="new-password"
                                    required
                                    minlength="8"
                                />
                                <button type="button" class="toggle-password" data-target="register-password" aria-label="Toggle password visibility">👁</button>
                            </div>
                            <span class="field-error" id="register-password-error"></span>
                            <div class="password-strength">
                                <div class="strength-bar">
                                    <div class="strength-fill" id="strength-fill"></div>
                                </div>
                                <div class="strength-text" id="strength-text"></div>
                            </div>
                        </div>

                        <div class="form-group">
                            <label for="register-confirm">Confirm Password</label>
                            <div class="input-wrapper">
                                <input
                                    type="password"
                                    id="register-confirm"
                                    name="confirm_password"
                                    placeholder="Repeat your password"
                                    autocomplete="new-password"
                                    required
                                />
                                <button type="button" class="toggle-password" data-target="register-confirm" aria-label="Toggle password visibility">👁</button>
                            </div>
                            <span class="field-error" id="register-confirm-error"></span>
                        </div>

                        <button type="submit" class="btn btn-primary" id="register-submit">
                            <span class="btn-text">Create Account</span>
                        </button>
                    </form>

                    <div class="form-footer">
                        Already have an account? <a href="#/login">Sign in</a>
                    </div>
                </div>
            </div>
        </div>
    `;
}

export function initRegisterPage() {
    const form = document.getElementById('register-form');
    const alertEl = document.getElementById('register-alert');
    const submitBtn = document.getElementById('register-submit');
    const nameInput = document.getElementById('register-name');
    const emailInput = document.getElementById('register-email');
    const passwordInput = document.getElementById('register-password');
    const confirmInput = document.getElementById('register-confirm');
    const strengthFill = document.getElementById('strength-fill');
    const strengthText = document.getElementById('strength-text');

    // Password toggle
    document.querySelectorAll('.toggle-password').forEach(btn => {
        btn.addEventListener('click', () => {
            const input = document.getElementById(btn.dataset.target);
            const isPassword = input.type === 'password';
            input.type = isPassword ? 'text' : 'password';
            btn.textContent = isPassword ? '🙈' : '👁';
        });
    });

    // Password strength meter
    passwordInput.addEventListener('input', () => {
        const strength = getPasswordStrength(passwordInput.value);
        strengthFill.className = `strength-fill ${strength.level}`;
        strengthText.textContent = strength.text;
        strengthText.style.color = strength.color;
    });

    // Clear field errors on input
    [nameInput, emailInput, passwordInput, confirmInput].forEach(input => {
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
        const name = nameInput.value.trim();
        const email = emailInput.value.trim();
        const password = passwordInput.value;
        const confirm = confirmInput.value;

        if (!name || name.length < 2) {
            showFieldError('register-name', 'Name must be at least 2 characters');
            valid = false;
        }

        if (!email) {
            showFieldError('register-email', 'Email is required');
            valid = false;
        } else if (!isValidEmail(email)) {
            showFieldError('register-email', 'Please enter a valid email');
            valid = false;
        }

        if (!password) {
            showFieldError('register-password', 'Password is required');
            valid = false;
        } else if (password.length < 8) {
            showFieldError('register-password', 'Must be at least 8 characters');
            valid = false;
        } else if (!hasComplexity(password)) {
            showFieldError('register-password', 'Must include uppercase, lowercase, and a number');
            valid = false;
        }

        if (password !== confirm) {
            showFieldError('register-confirm', 'Passwords do not match');
            valid = false;
        }

        if (!valid) return;

        setLoading(submitBtn, true);
        alertEl.style.display = 'none';

        try {
            await register(email, password, name);

            // Check if a plan was selected from the landing page
            const pendingPlan = localStorage.getItem('gsl_pending_plan');
            if (pendingPlan && (pendingPlan === 'monthly' || pendingPlan === 'yearly')) {
                try {
                    const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080';
                    await fetch(`${API_BASE}/api/auth/subscribe`, {
                        method: 'POST',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify({ email, plan: pendingPlan }),
                    });
                    localStorage.removeItem('gsl_pending_plan');
                } catch (subErr) {
                    console.warn('Failed to activate pending plan:', subErr);
                }
            }

            // Show success message
            alertEl.className = 'alert alert-success';
            const planMsg = pendingPlan ? ` Your ${pendingPlan} plan is now active!` : '';
            alertEl.textContent = `✅ Account created!${planMsg} Redirecting to login...`;
            alertEl.style.display = 'flex';

            setTimeout(() => {
                window.location.hash = '#/login';
            }, 1500);
        } catch (err) {
            alertEl.className = 'alert alert-error';
            alertEl.textContent = `⚠ ${err.message}`;
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
        btn.innerHTML = '<div class="spinner"></div><span>Creating account...</span>';
    } else {
        btn.disabled = false;
        btn.innerHTML = '<span class="btn-text">Create Account</span>';
    }
}

function isValidEmail(email) {
    return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
}

function hasComplexity(password) {
    return /[a-z]/.test(password) && /[A-Z]/.test(password) && /\d/.test(password);
}

function getPasswordStrength(password) {
    if (!password) return { level: '', text: '', color: '' };

    let score = 0;
    if (password.length >= 8) score++;
    if (password.length >= 12) score++;
    if (/[a-z]/.test(password) && /[A-Z]/.test(password)) score++;
    if (/\d/.test(password)) score++;
    if (/[^a-zA-Z0-9]/.test(password)) score++;

    if (score <= 1) return { level: 'weak', text: 'Weak', color: '#ef4444' };
    if (score <= 2) return { level: 'fair', text: 'Fair', color: '#f59e0b' };
    if (score <= 3) return { level: 'good', text: 'Good', color: '#34d399' };
    return { level: 'strong', text: 'Strong', color: '#10b981' };
}
