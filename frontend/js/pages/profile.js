// ============================================
// Profile Page — User self-service management
// ============================================

import { getCurrentUser, getUser } from '/js/auth.js';
import { apiRequest, clearTokens, setCachedUser } from '/js/utils/api.js';

/**
 * Renders the profile page content.
 */
export function renderProfilePage() {
    return `
        <div class="profile-page">
            <!-- Personal Info Section -->
            <div class="profile-section" id="profile-info-section">
                <div class="profile-section-header">
                    <div class="profile-section-title">
                        <span class="profile-section-icon">👤</span>
                        <div>
                            <h3>Personal Information</h3>
                            <p>Manage your account details</p>
                        </div>
                    </div>
                    <button class="btn-edit" id="edit-profile-btn">
                        <span>✏️</span> Edit
                    </button>
                </div>

                <!-- View Mode -->
                <div id="profile-view-mode">
                    <div class="profile-info-loading" id="profile-loading">
                        <div class="spinner"></div>
                        <span>Loading profile...</span>
                    </div>
                    <div id="profile-info-grid" style="display:none;">
                        <div class="profile-field">
                            <span class="profile-field-label">Full Name</span>
                            <span class="profile-field-value" id="profile-name">—</span>
                        </div>
                        <div class="profile-field">
                            <span class="profile-field-label">Email Address</span>
                            <span class="profile-field-value" id="profile-email">—</span>
                        </div>
                        <div class="profile-field">
                            <span class="profile-field-label">Role</span>
                            <span class="profile-field-value" id="profile-role"></span>
                        </div>
                        <div class="profile-field">
                            <span class="profile-field-label">Subscription</span>
                            <span class="profile-field-value" id="profile-subscription"></span>
                        </div>
                        <div class="profile-field">
                            <span class="profile-field-label">Member Since</span>
                            <span class="profile-field-value" id="profile-since">—</span>
                        </div>
                    </div>
                </div>

                <!-- Edit Mode -->
                <div id="profile-edit-mode" style="display:none;">
                    <div id="edit-profile-alert" class="alert" style="display:none;"></div>
                    <form id="edit-profile-form">
                        <div class="form-group">
                            <label for="edit-name">Full Name</label>
                            <input type="text" id="edit-name" placeholder="Your full name" autocomplete="name">
                        </div>
                        <div class="form-group">
                            <label for="edit-email">Email Address</label>
                            <input type="email" id="edit-email" placeholder="you@example.com" autocomplete="email">
                        </div>
                        <div class="profile-edit-actions">
                            <button type="submit" class="btn btn-primary btn-sm" id="save-profile-btn">
                                Save Changes
                            </button>
                            <button type="button" class="btn btn-ghost btn-sm" id="cancel-edit-btn">
                                Cancel
                            </button>
                        </div>
                    </form>
                </div>
            </div>

            <!-- Change Password Section -->
            <div class="profile-section">
                <div class="profile-section-header">
                    <div class="profile-section-title">
                        <span class="profile-section-icon">🔒</span>
                        <div>
                            <h3>Change Password</h3>
                            <p>Update your password to keep your account secure</p>
                        </div>
                    </div>
                </div>
                <div id="password-alert" class="alert" style="display:none;"></div>
                <form id="change-password-form">
                    <div class="form-group">
                        <label for="current-password">Current Password</label>
                        <div class="input-wrapper">
                            <input type="password" id="current-password" placeholder="Enter current password" autocomplete="current-password">
                            <button type="button" class="toggle-password" data-target="current-password">👁️</button>
                        </div>
                    </div>
                    <div class="form-group">
                        <label for="new-password">New Password</label>
                        <div class="input-wrapper">
                            <input type="password" id="new-password" placeholder="Enter new password" autocomplete="new-password">
                            <button type="button" class="toggle-password" data-target="new-password">👁️</button>
                        </div>
                        <div class="password-strength">
                            <div class="strength-bar"><div class="strength-fill" id="new-pw-strength-fill"></div></div>
                            <span class="strength-text" id="new-pw-strength-text"></span>
                        </div>
                    </div>
                    <div class="form-group">
                        <label for="confirm-password">Confirm New Password</label>
                        <div class="input-wrapper">
                            <input type="password" id="confirm-password" placeholder="Re-enter new password" autocomplete="new-password">
                            <button type="button" class="toggle-password" data-target="confirm-password">👁️</button>
                        </div>
                    </div>
                    <button type="submit" class="btn btn-primary btn-sm" id="change-password-btn" style="width:auto;">
                        Update Password
                    </button>
                </form>
            </div>

            <!-- Danger Zone -->
            <div class="profile-section profile-danger-zone">
                <div class="profile-section-header">
                    <div class="profile-section-title">
                        <span class="profile-section-icon">⚠️</span>
                        <div>
                            <h3>Danger Zone</h3>
                            <p>Irreversible actions — proceed with caution</p>
                        </div>
                    </div>
                </div>
                <div class="danger-zone-content">
                    <div class="danger-zone-info">
                        <strong>Delete your account</strong>
                        <p>Once deleted, your account and all associated data will be permanently removed. This action cannot be undone.</p>
                    </div>
                    <button class="btn btn-danger btn-sm" id="delete-account-btn">
                        Delete Account
                    </button>
                </div>
            </div>

            <!-- Delete Confirmation Modal -->
            <div class="modal-overlay" id="delete-modal" style="display:none;">
                <div class="modal-card">
                    <div class="modal-icon">🗑️</div>
                    <h3>Delete Account</h3>
                    <p>This action is <strong>permanent</strong>. All your data will be erased.</p>
                    <p class="modal-instruction">Type <strong>DELETE</strong> below to confirm:</p>
                    <input type="text" id="delete-confirm-input" class="modal-input" placeholder="Type DELETE" autocomplete="off">
                    <div class="modal-actions">
                        <button class="btn btn-danger btn-sm" id="confirm-delete-btn" disabled>
                            Permanently Delete
                        </button>
                        <button class="btn btn-ghost btn-sm" id="cancel-delete-btn">
                            Cancel
                        </button>
                    </div>
                </div>
            </div>
        </div>
    `;
}

/**
 * Initialize profile page interactivity.
 */
export function initProfilePage() {
    loadProfileData();
    setupEditProfile();
    setupChangePassword();
    setupDeleteAccount();
    setupPasswordToggles();
}

// ─── Load Profile Data ────────────────────────────
async function loadProfileData() {
    const loading = document.getElementById('profile-loading');
    const grid = document.getElementById('profile-info-grid');

    try {
        const user = await getCurrentUser();
        fillProfileFields(user);
        if (loading) loading.style.display = 'none';
        if (grid) grid.style.display = 'grid';
    } catch (err) {
        // Fall back to cached user
        const cached = getUser();
        if (cached) {
            fillProfileFields(cached);
            if (loading) loading.style.display = 'none';
            if (grid) grid.style.display = 'grid';
        } else {
            if (loading) {
                loading.innerHTML = '<span style="color:var(--color-error)">Failed to load profile data.</span>';
            }
        }
    }
}

function fillProfileFields(user) {
    const nameEl = document.getElementById('profile-name');
    const emailEl = document.getElementById('profile-email');
    const roleEl = document.getElementById('profile-role');
    const subEl = document.getElementById('profile-subscription');
    const sinceEl = document.getElementById('profile-since');

    if (nameEl) nameEl.textContent = user.name || '—';
    if (emailEl) emailEl.textContent = user.email || '—';
    if (roleEl) roleEl.innerHTML = `<span class="badge badge-role">${(user.role || 'user').toUpperCase()}</span>`;
    if (subEl) {
        subEl.innerHTML = user.subscription_active
            ? '<span class="badge badge-active">✓ Active</span>'
            : '<span class="badge badge-inactive">Inactive</span>';
    }
    if (sinceEl) {
        const date = new Date(user.created_at);
        sinceEl.textContent = date.toLocaleDateString('en-US', {
            year: 'numeric', month: 'long', day: 'numeric'
        });
    }
}

// ─── Edit Profile ─────────────────────────────────
function setupEditProfile() {
    const editBtn = document.getElementById('edit-profile-btn');
    const cancelBtn = document.getElementById('cancel-edit-btn');
    const form = document.getElementById('edit-profile-form');
    const viewMode = document.getElementById('profile-view-mode');
    const editMode = document.getElementById('profile-edit-mode');

    if (editBtn) {
        editBtn.addEventListener('click', () => {
            const user = getUser();
            document.getElementById('edit-name').value = user?.name || '';
            document.getElementById('edit-email').value = user?.email || '';
            viewMode.style.display = 'none';
            editMode.style.display = 'block';
            editBtn.style.display = 'none';
            hideAlert('edit-profile-alert');
        });
    }

    if (cancelBtn) {
        cancelBtn.addEventListener('click', () => {
            viewMode.style.display = 'block';
            editMode.style.display = 'none';
            editBtn.style.display = 'inline-flex';
        });
    }

    if (form) {
        form.addEventListener('submit', async (e) => {
            e.preventDefault();
            const saveBtn = document.getElementById('save-profile-btn');
            const name = document.getElementById('edit-name').value.trim();
            const email = document.getElementById('edit-email').value.trim();

            if (!name || !email) {
                showAlert('edit-profile-alert', 'Name and email are required.', 'error');
                return;
            }

            saveBtn.disabled = true;
            saveBtn.innerHTML = '<div class="spinner"></div> Saving...';

            try {
                const data = await apiRequest('/api/users/me', {
                    method: 'PUT',
                    body: JSON.stringify({ name, email }),
                });

                const updatedUser = data.data;
                setCachedUser(updatedUser);
                fillProfileFields(updatedUser);

                viewMode.style.display = 'block';
                editMode.style.display = 'none';
                editBtn.style.display = 'inline-flex';

                showAlert('edit-profile-alert', 'Profile updated successfully!', 'success');
                // Brief flash of success, then hide
                setTimeout(() => hideAlert('edit-profile-alert'), 3000);
            } catch (err) {
                showAlert('edit-profile-alert', err.message || 'Failed to update profile.', 'error');
            } finally {
                saveBtn.disabled = false;
                saveBtn.innerHTML = 'Save Changes';
            }
        });
    }
}

// ─── Change Password ──────────────────────────────
function setupChangePassword() {
    const form = document.getElementById('change-password-form');
    const newPwInput = document.getElementById('new-password');

    // Password strength meter
    if (newPwInput) {
        newPwInput.addEventListener('input', () => {
            updatePasswordStrength(newPwInput.value);
        });
    }

    if (form) {
        form.addEventListener('submit', async (e) => {
            e.preventDefault();
            const btn = document.getElementById('change-password-btn');
            const currentPw = document.getElementById('current-password').value;
            const newPw = document.getElementById('new-password').value;
            const confirmPw = document.getElementById('confirm-password').value;

            hideAlert('password-alert');

            if (!currentPw || !newPw || !confirmPw) {
                showAlert('password-alert', 'All password fields are required.', 'error');
                return;
            }
            if (newPw !== confirmPw) {
                showAlert('password-alert', 'New passwords do not match.', 'error');
                return;
            }
            if (newPw.length < 8) {
                showAlert('password-alert', 'New password must be at least 8 characters.', 'error');
                return;
            }

            btn.disabled = true;
            btn.innerHTML = '<div class="spinner"></div> Updating...';

            try {
                await apiRequest('/api/users/me/password', {
                    method: 'PUT',
                    body: JSON.stringify({
                        current_password: currentPw,
                        new_password: newPw,
                    }),
                });

                showAlert('password-alert', 'Password changed successfully!', 'success');
                form.reset();
                updatePasswordStrength('');
                setTimeout(() => hideAlert('password-alert'), 4000);
            } catch (err) {
                showAlert('password-alert', err.message || 'Failed to change password.', 'error');
            } finally {
                btn.disabled = false;
                btn.innerHTML = 'Update Password';
            }
        });
    }
}

function updatePasswordStrength(password) {
    const fill = document.getElementById('new-pw-strength-fill');
    const text = document.getElementById('new-pw-strength-text');
    if (!fill || !text) return;

    if (!password) {
        fill.className = 'strength-fill';
        text.textContent = '';
        return;
    }

    let score = 0;
    if (password.length >= 8) score++;
    if (password.length >= 12) score++;
    if (/[A-Z]/.test(password) && /[a-z]/.test(password)) score++;
    if (/[0-9]/.test(password)) score++;
    if (/[^A-Za-z0-9]/.test(password)) score++;

    const levels = [
        { cls: 'weak', label: 'Weak' },
        { cls: 'weak', label: 'Weak' },
        { cls: 'fair', label: 'Fair' },
        { cls: 'good', label: 'Good' },
        { cls: 'strong', label: 'Strong' },
        { cls: 'strong', label: 'Very Strong' },
    ];

    const level = levels[score] || levels[0];
    fill.className = `strength-fill ${level.cls}`;
    text.textContent = level.label;
}

// ─── Delete Account ───────────────────────────────
function setupDeleteAccount() {
    const deleteBtn = document.getElementById('delete-account-btn');
    const modal = document.getElementById('delete-modal');
    const confirmInput = document.getElementById('delete-confirm-input');
    const confirmBtn = document.getElementById('confirm-delete-btn');
    const cancelBtn = document.getElementById('cancel-delete-btn');

    if (deleteBtn) {
        deleteBtn.addEventListener('click', () => {
            modal.style.display = 'flex';
            confirmInput.value = '';
            confirmBtn.disabled = true;
        });
    }

    if (cancelBtn) {
        cancelBtn.addEventListener('click', () => {
            modal.style.display = 'none';
        });
    }

    // Close on overlay click
    if (modal) {
        modal.addEventListener('click', (e) => {
            if (e.target === modal) {
                modal.style.display = 'none';
            }
        });
    }

    if (confirmInput) {
        confirmInput.addEventListener('input', () => {
            confirmBtn.disabled = confirmInput.value.trim() !== 'DELETE';
        });
    }

    if (confirmBtn) {
        confirmBtn.addEventListener('click', async () => {
            if (confirmInput.value.trim() !== 'DELETE') return;

            confirmBtn.disabled = true;
            confirmBtn.innerHTML = '<div class="spinner"></div> Deleting...';

            try {
                await apiRequest('/api/users/me', { method: 'DELETE' });
                clearTokens();
                window.location.hash = '#/login';
            } catch (err) {
                confirmBtn.disabled = false;
                confirmBtn.innerHTML = 'Permanently Delete';
                alert(err.message || 'Failed to delete account.');
            }
        });
    }
}

// ─── Password Toggle ──────────────────────────────
function setupPasswordToggles() {
    document.querySelectorAll('.toggle-password').forEach(btn => {
        btn.addEventListener('click', () => {
            const target = document.getElementById(btn.dataset.target);
            if (target) {
                const isPassword = target.type === 'password';
                target.type = isPassword ? 'text' : 'password';
                btn.textContent = isPassword ? '🙈' : '👁️';
            }
        });
    });
}

// ─── Alert Helpers ────────────────────────────────
function showAlert(id, message, type) {
    const el = document.getElementById(id);
    if (!el) return;
    el.className = `alert alert-${type}`;
    el.textContent = message;
    el.style.display = 'flex';
}

function hideAlert(id) {
    const el = document.getElementById(id);
    if (el) el.style.display = 'none';
}
