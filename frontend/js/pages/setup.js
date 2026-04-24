// ============================================
// Admin Setup Component (One-time use)
// ============================================

import { setupSystemAdmin, checkSetupStatus } from '../auth.js';

export function renderSetupPage() {
    return `
        <div class="page">
            <div class="auth-container">
                <div class="glass-card">
                    <div class="brand">
                        <div class="brand-icon">🛡️</div>
                        <h1>SuperAdmin Setup</h1>
                        <p>Register the master administrator account. This can only be done once.</p>
                    </div>

                    <div id="error-message" class="alert" style="display:none;"></div>

                    <form id="setup-form" novalidate>
                        <div class="form-group">
                            <label for="name">Admin Full Name</label>
                            <input type="text" id="name" placeholder="John Doe" autocomplete="name" required />
                        </div>

                        <div class="form-group">
                            <label for="email">Admin Email Address</label>
                            <input type="email" id="email" placeholder="admin@example.com" autocomplete="email" required />
                        </div>

                        <div class="form-group">
                            <label for="password">Master Password</label>
                            <div class="input-wrapper">
                                <input type="password" id="password" placeholder="SuperSecure123!" autocomplete="new-password" required />
                                <button type="button" class="toggle-password" data-target="password" aria-label="Toggle password visibility">👁</button>
                            </div>
                        </div>

                        <button type="submit" class="btn btn-primary" id="submit-btn">
                            <span class="btn-text">Initialize System Admin</span>
                        </button>
                    </form>
                </div>
            </div>
        </div>
    `;
}

export async function initSetupPage() {
    // 1. HARD SECURITY CHECK: Lockout if setup already complete
    try {
        const { setup_complete } = await checkSetupStatus();
        if (setup_complete) {
            window.location.hash = '#/login'; // Boot them out
            return;
        }
    } catch {
        window.location.hash = '#/login'; // Boot them out securely on failure
        return;
    }

    // 2. Setup the UI interactions
    const form = document.getElementById('setup-form');
    const togglePassword = document.querySelector('.toggle-password');
    const passwordInput = document.getElementById('password');
    const errorMsg = document.getElementById('error-message');
    const submitBtn = document.getElementById('submit-btn');

    // Toggle password visibility (Microinteraction)
    if (togglePassword) {
        togglePassword.addEventListener('click', () => {
            const type = passwordInput.getAttribute('type') === 'password' ? 'text' : 'password';
            passwordInput.setAttribute('type', type);
            togglePassword.textContent = type === 'password' ? 'visibility' : 'visibility_off';
        });
    }

    // Form Submission
    form.addEventListener('submit', async (e) => {
        e.preventDefault();
        
        const name = document.getElementById('name').value;
        const email = document.getElementById('email').value;
        const password = passwordInput.value;

        // Reset state
        errorMsg.style.display = 'none';
        submitBtn.classList.add('loading');
        submitBtn.disabled = true;

        try {
            await setupSystemAdmin(email, password, name);
            // On success, redirect to login so they can authenticate normally
            window.location.hash = '#/login';
        } catch (err) {
            errorMsg.textContent = err.message || 'Setup failed. Please try again.';
            errorMsg.style.display = 'block';
            
            // Jiggle animation
            errorMsg.classList.add('jiggle');
            setTimeout(() => errorMsg.classList.remove('jiggle'), 300);
        } finally {
            submitBtn.classList.remove('loading');
            submitBtn.disabled = false;
        }
    });
}
