import { apiRequest } from '../utils/api.js';

// State
let currentLevel = 'users'; // users | years | months | events
let selectedUser = null;
let selectedYear = null;
let selectedMonth = null;

export function renderAdminActivityLogs() {
    return `
        <div class="admin-users-page">
            <div class="admin-users-header">
                <div class="admin-users-title-group">
                    <h2>Activity Logs</h2>
                    <p>Drill down into system-wide user activity</p>
                </div>
            </div>

            <div id="drill-breadcrumb" class="drill-breadcrumb"></div>
            <div id="drill-content">
                <div class="table-loading" style="padding: 3rem; text-align: center;">
                    <div class="spinner"></div>
                    <span>Loading activity data...</span>
                </div>
            </div>
        </div>
    `;
}

export async function initAdminActivityLogs() {
    currentLevel = 'users';
    selectedUser = null;
    selectedYear = null;
    selectedMonth = null;
    await loadLevel();
}

async function loadLevel() {
    renderBreadcrumb();
    const container = document.getElementById('drill-content');
    container.innerHTML = '<div style="padding:3rem;text-align:center"><div class="spinner"></div><span style="color:var(--color-text-muted);margin-top:1rem;display:block">Loading...</span></div>';

    try {
        if (currentLevel === 'users') {
            const users = await apiRequest('/api/admin/activity-logs/users');
            renderUsersList(users || []);
        } else if (currentLevel === 'years') {
            const years = await apiRequest(`/api/admin/activity-logs/users/${selectedUser.id}/years`);
            renderYearsList(years || []);
        } else if (currentLevel === 'months') {
            const months = await apiRequest(`/api/admin/activity-logs/users/${selectedUser.id}/years/${selectedYear}/months`);
            renderMonthsList(months || []);
        } else if (currentLevel === 'events') {
            const events = await apiRequest(`/api/admin/activity-logs/users/${selectedUser.id}/years/${selectedYear}/months/${selectedMonth}`);
            renderEventsList(events || []);
        }
    } catch (err) {
        container.innerHTML = `<div class="empty-state">⚠️ Error: ${err.message}</div>`;
    }
}

function renderBreadcrumb() {
    const bc = document.getElementById('drill-breadcrumb');
    if (!bc) return;

    let html = '';

    if (currentLevel === 'users') {
        html = '<span class="drill-breadcrumb-current">📄 All Users</span>';
    } else {
        html += '<span class="drill-breadcrumb-item" data-level="users">📄 All Users</span>';

        if (currentLevel === 'years') {
            html += '<span class="drill-breadcrumb-sep">›</span>';
            html += `<span class="drill-breadcrumb-current">👤 ${selectedUser.name}</span>`;
        } else if (currentLevel === 'months') {
            html += '<span class="drill-breadcrumb-sep">›</span>';
            html += `<span class="drill-breadcrumb-item" data-level="years">👤 ${selectedUser.name}</span>`;
            html += '<span class="drill-breadcrumb-sep">›</span>';
            html += `<span class="drill-breadcrumb-current">📅 ${selectedYear}</span>`;
        } else if (currentLevel === 'events') {
            html += '<span class="drill-breadcrumb-sep">›</span>';
            html += `<span class="drill-breadcrumb-item" data-level="years">👤 ${selectedUser.name}</span>`;
            html += '<span class="drill-breadcrumb-sep">›</span>';
            html += `<span class="drill-breadcrumb-item" data-level="months">📅 ${selectedYear}</span>`;
            html += '<span class="drill-breadcrumb-sep">›</span>';
            html += `<span class="drill-breadcrumb-current">🗓️ Month ${selectedMonth}</span>`;
        }
    }

    bc.innerHTML = html;

    bc.querySelectorAll('.drill-breadcrumb-item').forEach(item => {
        item.addEventListener('click', () => {
            const level = item.dataset.level;
            currentLevel = level;
            if (level === 'users') { selectedUser = null; selectedYear = null; selectedMonth = null; }
            if (level === 'years') { selectedYear = null; selectedMonth = null; }
            if (level === 'months') { selectedMonth = null; }
            loadLevel();
        });
    });
}

function renderUsersList(users) {
    const container = document.getElementById('drill-content');
    if (users.length === 0) {
        container.innerHTML = '<div class="empty-state">No activity logs found yet.</div>';
        return;
    }

    container.innerHTML = `<div class="drill-list">${users.map(u => `
        <div class="drill-card" data-id="${u.user_id}" data-name="${escapeAttr(u.user_name)}" data-email="${escapeAttr(u.user_email)}">
            <div class="drill-card-title">${escapeHtml(u.user_name)}</div>
            <div class="drill-card-subtitle">${escapeHtml(u.user_email)}</div>
            <div class="drill-card-meta">
                <div>
                    <div class="drill-card-count">${u.activity_count}</div>
                    <div class="drill-card-label">Activities</div>
                </div>
                <span class="drill-card-arrow">→</span>
            </div>
        </div>
    `).join('')}</div>`;

    container.querySelectorAll('.drill-card').forEach(card => {
        card.addEventListener('click', () => {
            selectedUser = { id: card.dataset.id, name: card.dataset.name, email: card.dataset.email };
            currentLevel = 'years';
            loadLevel();
        });
    });
}

function renderYearsList(years) {
    const container = document.getElementById('drill-content');
    if (years.length === 0) {
        container.innerHTML = '<div class="empty-state">No activity found for this user.</div>';
        return;
    }

    container.innerHTML = `<div class="drill-list">${years.map(y => `
        <div class="drill-card" data-year="${y.year}">
            <div class="drill-card-title">📅 ${y.year}</div>
            <div class="drill-card-meta">
                <div>
                    <div class="drill-card-count">${y.activity_count}</div>
                    <div class="drill-card-label">Activities</div>
                </div>
                <span class="drill-card-arrow">→</span>
            </div>
        </div>
    `).join('')}</div>`;

    container.querySelectorAll('.drill-card').forEach(card => {
        card.addEventListener('click', () => {
            selectedYear = parseInt(card.dataset.year);
            currentLevel = 'months';
            loadLevel();
        });
    });
}

function renderMonthsList(months) {
    const container = document.getElementById('drill-content');
    if (months.length === 0) {
        container.innerHTML = '<div class="empty-state">No activity found for this year.</div>';
        return;
    }

    container.innerHTML = `<div class="drill-list">${months.map(m => `
        <div class="drill-card" data-month="${m.month}">
            <div class="drill-card-title">🗓️ ${m.month_name.trim()}</div>
            <div class="drill-card-meta">
                <div>
                    <div class="drill-card-count">${m.activity_count}</div>
                    <div class="drill-card-label">Activities</div>
                </div>
                <span class="drill-card-arrow">→</span>
            </div>
        </div>
    `).join('')}</div>`;

    container.querySelectorAll('.drill-card').forEach(card => {
        card.addEventListener('click', () => {
            selectedMonth = parseInt(card.dataset.month);
            currentLevel = 'events';
            loadLevel();
        });
    });
}

function renderEventsList(events) {
    const container = document.getElementById('drill-content');
    if (events.length === 0) {
        container.innerHTML = '<div class="empty-state">No events found for this month.</div>';
        return;
    }

    container.innerHTML = `
        <div class="admin-table-container">
            <table class="admin-table">
                <thead>
                    <tr>
                        <th>Timestamp</th>
                        <th>Action</th>
                        <th>Entity</th>
                        <th>Details</th>
                        <th>IP</th>
                    </tr>
                </thead>
                <tbody>
                    ${events.map(e => `
                        <tr class="table-row">
                            <td class="table-date">${new Date(e.created_at).toLocaleString()}</td>
                            <td><span class="badge badge-role">${escapeHtml(e.action)}</span></td>
                            <td>${escapeHtml(e.entity_type)} <span class="text-xs text-muted">[${e.entity_id ? e.entity_id.substring(0, 8) : '-'}]</span></td>
                            <td><pre style="margin:0;font-size:11px;background:transparent;padding:0;color:var(--color-text-muted);white-space:pre-wrap;max-width:300px">${JSON.stringify(e.metadata, null, 2)}</pre></td>
                            <td class="text-muted text-xs">${e.ip_address || '-'}</td>
                        </tr>
                    `).join('')}
                </tbody>
            </table>
        </div>
    `;
}

function escapeHtml(str) {
    if (!str) return '';
    const div = document.createElement('div');
    div.textContent = str;
    return div.innerHTML;
}

function escapeAttr(str) {
    if (!str) return '';
    return str.replace(/"/g, '&quot;').replace(/'/g, '&#39;');
}
