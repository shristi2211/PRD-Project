// ============================================
// Admin Users Page — User management data table
// ============================================

import { apiRequest } from '/js/utils/api.js';

// State
let currentPage = 1;
let currentPageSize = 10;
let currentSearch = '';
let currentStatus = '';
let searchDebounceTimer = null;

/**
 * Renders the admin users page content.
 */
export function renderAdminUsersPage() {
    return `
        <div class="admin-users-page">
            <!-- Header -->
            <div class="admin-users-header">
                <div class="admin-users-title-group">
                    <h2>User Management</h2>
                    <p>View, search, and manage all registered users</p>
                </div>
                <div class="admin-users-stat" id="admin-users-total-stat">
                    <span class="admin-users-stat-value" id="total-users-count">—</span>
                    <span class="admin-users-stat-label">Total Users</span>
                </div>
            </div>

            <!-- Toolbar -->
            <div class="admin-users-toolbar">
                <div class="admin-users-search-wrapper">
                    <span class="search-icon">🔍</span>
                    <input
                        type="text"
                        id="admin-users-search"
                        class="admin-users-search"
                        placeholder="Search by name or email..."
                        autocomplete="off"
                    >
                </div>
                <div class="admin-users-filters">
                    <select id="admin-users-status-filter" class="admin-users-select">
                        <option value="">All Status</option>
                        <option value="active">Active</option>
                        <option value="inactive">Inactive</option>
                    </select>
                    <select id="admin-users-pagesize" class="admin-users-select">
                        <option value="5">5 per page</option>
                        <option value="10" selected>10 per page</option>
                        <option value="25">25 per page</option>
                        <option value="50">50 per page</option>
                    </select>
                </div>
            </div>

            <!-- Data Table -->
            <div class="admin-table-container">
                <table class="admin-table" id="admin-users-table">
                    <thead>
                        <tr>
                            <th>User</th>
                            <th>Email</th>
                            <th>Role</th>
                            <th>Status</th>
                            <th>Joined</th>
                            <th>Actions</th>
                        </tr>
                    </thead>
                    <tbody id="admin-users-tbody">
                        <tr>
                            <td colspan="6" class="table-loading">
                                <div class="spinner"></div>
                                <span>Loading users...</span>
                            </td>
                        </tr>
                    </tbody>
                </table>
            </div>

            <!-- Pagination -->
            <div class="admin-pagination" id="admin-users-pagination"></div>
        </div>
    `;
}

/**
 * Initialize admin users page interactivity.
 */
export function initAdminUsersPage() {
    // Reset state
    currentPage = 1;
    currentSearch = '';
    currentStatus = '';

    const searchInput = document.getElementById('admin-users-search');
    const statusFilter = document.getElementById('admin-users-status-filter');
    const pageSizeSelect = document.getElementById('admin-users-pagesize');

    // Search — debounced
    if (searchInput) {
        searchInput.addEventListener('input', () => {
            clearTimeout(searchDebounceTimer);
            searchDebounceTimer = setTimeout(() => {
                currentSearch = searchInput.value.trim();
                currentPage = 1;
                loadUsers();
            }, 350);
        });
    }

    // Status filter
    if (statusFilter) {
        statusFilter.addEventListener('change', () => {
            currentStatus = statusFilter.value;
            currentPage = 1;
            loadUsers();
        });
    }

    // Page size
    if (pageSizeSelect) {
        pageSizeSelect.value = currentPageSize.toString();
        pageSizeSelect.addEventListener('change', () => {
            currentPageSize = parseInt(pageSizeSelect.value) || 10;
            currentPage = 1;
            loadUsers();
        });
    }

    // Initial load
    loadUsers();
}

// ─── Data Loading ─────────────────────────────────
async function loadUsers() {
    const tbody = document.getElementById('admin-users-tbody');
    if (!tbody) return;

    tbody.innerHTML = `
        <tr>
            <td colspan="6" class="table-loading">
                <div class="spinner"></div>
                <span>Loading users...</span>
            </td>
        </tr>
    `;

    try {
        const params = new URLSearchParams({
            page: currentPage.toString(),
            size: currentPageSize.toString(),
        });
        if (currentSearch) params.set('search', currentSearch);
        if (currentStatus) params.set('status', currentStatus);

        const response = await apiRequest(`/api/admin/users?${params.toString()}`);
        const data = response;

        // Update total count
        const totalEl = document.getElementById('total-users-count');
        if (totalEl) totalEl.textContent = (data.total || 0).toLocaleString();

        renderUsersTable(data.users || []);
        renderPagination(data.total || 0, data.page || 1, data.page_size || 10);
    } catch (err) {
        tbody.innerHTML = `
            <tr>
                <td colspan="6" class="table-error">
                    <span>⚠️ ${err.message || 'Failed to load users'}</span>
                    <button class="btn btn-ghost btn-sm" onclick="location.reload()">Retry</button>
                </td>
            </tr>
        `;
    }
}

function renderUsersTable(users) {
    const tbody = document.getElementById('admin-users-tbody');
    if (!tbody) return;

    if (users.length === 0) {
        tbody.innerHTML = `
            <tr>
                <td colspan="6" class="table-empty">
                    <div class="table-empty-icon">📋</div>
                    <p>No users found</p>
                    <span>Try adjusting your search or filters</span>
                </td>
            </tr>
        `;
        return;
    }

    tbody.innerHTML = users.map(user => {
        const initial = user.name ? user.name.charAt(0).toUpperCase() : '?';
        const joinDate = new Date(user.created_at).toLocaleDateString('en-US', {
            year: 'numeric', month: 'short', day: 'numeric'
        });
        const isActive = user.subscription_active;
        const statusBadge = isActive
            ? '<span class="badge badge-active">Active</span>'
            : '<span class="badge badge-inactive">Inactive</span>';
        const roleBadge = user.role === 'admin'
            ? '<span class="badge badge-role">ADMIN</span>'
            : '<span class="badge badge-role-user">USER</span>';
        const toggleLabel = isActive ? 'Deactivate' : 'Activate';
        const toggleClass = isActive ? 'btn-toggle-deactivate' : 'btn-toggle-activate';

        return `
            <tr class="table-row" data-user-id="${user.id}">
                <td>
                    <div class="table-user-cell">
                        <div class="table-avatar">${initial}</div>
                        <span class="table-user-name">${escapeHtml(user.name)}</span>
                    </div>
                </td>
                <td class="table-email">${escapeHtml(user.email)}</td>
                <td>${roleBadge}</td>
                <td class="table-status-cell" id="status-${user.id}">${statusBadge}</td>
                <td class="table-date">${joinDate}</td>
                <td>
                    <button
                        class="btn-table-action ${toggleClass}"
                        data-user-id="${user.id}"
                        data-active="${isActive}"
                        id="toggle-btn-${user.id}"
                        ${user.role === 'admin' ? 'disabled title="Cannot toggle admin"' : ''}
                    >
                        ${toggleLabel}
                    </button>
                </td>
            </tr>
        `;
    }).join('');

    // Attach toggle listeners
    tbody.querySelectorAll('.btn-table-action').forEach(btn => {
        btn.addEventListener('click', () => handleToggle(btn));
    });
}

function renderPagination(total, page, pageSize) {
    const container = document.getElementById('admin-users-pagination');
    if (!container) return;

    const totalPages = Math.ceil(total / pageSize);
    if (totalPages <= 1) {
        container.innerHTML = '';
        return;
    }

    let pagesHTML = '';

    // Previous button
    pagesHTML += `
        <button class="pagination-btn ${page <= 1 ? 'disabled' : ''}"
                data-page="${page - 1}" ${page <= 1 ? 'disabled' : ''}>
            ← Prev
        </button>
    `;

    // Page numbers
    const maxVisible = 5;
    let startPage = Math.max(1, page - Math.floor(maxVisible / 2));
    let endPage = Math.min(totalPages, startPage + maxVisible - 1);
    if (endPage - startPage < maxVisible - 1) {
        startPage = Math.max(1, endPage - maxVisible + 1);
    }

    if (startPage > 1) {
        pagesHTML += `<button class="pagination-btn" data-page="1">1</button>`;
        if (startPage > 2) pagesHTML += `<span class="pagination-ellipsis">…</span>`;
    }

    for (let i = startPage; i <= endPage; i++) {
        pagesHTML += `
            <button class="pagination-btn ${i === page ? 'active' : ''}" data-page="${i}">
                ${i}
            </button>
        `;
    }

    if (endPage < totalPages) {
        if (endPage < totalPages - 1) pagesHTML += `<span class="pagination-ellipsis">…</span>`;
        pagesHTML += `<button class="pagination-btn" data-page="${totalPages}">${totalPages}</button>`;
    }

    // Next button
    pagesHTML += `
        <button class="pagination-btn ${page >= totalPages ? 'disabled' : ''}"
                data-page="${page + 1}" ${page >= totalPages ? 'disabled' : ''}>
            Next →
        </button>
    `;

    // Info
    const startRecord = (page - 1) * pageSize + 1;
    const endRecord = Math.min(page * pageSize, total);

    container.innerHTML = `
        <span class="pagination-info">Showing ${startRecord}–${endRecord} of ${total}</span>
        <div class="pagination-controls">${pagesHTML}</div>
    `;

    // Attach page click listeners
    container.querySelectorAll('.pagination-btn:not(.disabled)').forEach(btn => {
        btn.addEventListener('click', () => {
            currentPage = parseInt(btn.dataset.page);
            loadUsers();
        });
    });
}

// ─── Toggle Activation ───────────────────────────
async function handleToggle(btn) {
    const userId = btn.dataset.userId;
    const currentlyActive = btn.dataset.active === 'true';
    const newActive = !currentlyActive;

    btn.disabled = true;
    btn.textContent = '...';

    try {
        await apiRequest(`/api/admin/users/${userId}/activation`, {
            method: 'PUT',
            body: JSON.stringify({ active: newActive }),
        });

        // Update UI without full reload
        btn.dataset.active = newActive.toString();
        btn.textContent = newActive ? 'Deactivate' : 'Activate';
        btn.className = `btn-table-action ${newActive ? 'btn-toggle-deactivate' : 'btn-toggle-activate'}`;

        const statusCell = document.getElementById(`status-${userId}`);
        if (statusCell) {
            statusCell.innerHTML = newActive
                ? '<span class="badge badge-active">Active</span>'
                : '<span class="badge badge-inactive">Inactive</span>';
        }
    } catch (err) {
        btn.textContent = currentlyActive ? 'Deactivate' : 'Activate';
        alert(err.message || 'Failed to update user status');
    } finally {
        btn.disabled = false;
    }
}

// ─── Helpers ─────────────────────────────────────
function escapeHtml(str) {
    const div = document.createElement('div');
    div.textContent = str;
    return div.innerHTML;
}
