import { apiRequest } from '../utils/api.js';

let currentPage = 1;
let currentSearch = '';
let searchTimer = null;

export function renderAdminScores() {
    return `
        <div class="admin-users-page">
            <div class="admin-users-header">
                <div class="admin-users-title-group">
                    <h2>Score Management</h2>
                    <p>View and manage all user scores</p>
                </div>
                <div class="admin-users-stat">
                    <span class="admin-users-stat-value" id="total-scores-count">—</span>
                    <span class="admin-users-stat-label">Total Scores</span>
                </div>
            </div>

            <div class="admin-users-toolbar">
                <div class="admin-users-search-wrapper">
                    <span class="search-icon">🔍</span>
                    <input type="text" id="score-search" class="admin-users-search" placeholder="Search by name or email..." autocomplete="off">
                </div>
            </div>

            <div class="admin-table-container">
                <table class="admin-table">
                    <thead>
                        <tr>
                            <th>User</th>
                            <th>Score</th>
                            <th>Round Date</th>
                            <th>Notes</th>
                            <th>Submitted</th>
                            <th>Actions</th>
                        </tr>
                    </thead>
                    <tbody id="scores-tbody">
                        <tr><td colspan="6" class="table-loading"><div class="spinner"></div><span>Loading scores...</span></td></tr>
                    </tbody>
                </table>
            </div>

            <div class="admin-pagination" id="scores-pagination"></div>
        </div>
    `;
}

export async function initAdminScores() {
    currentPage = 1;
    currentSearch = '';

    const searchInput = document.getElementById('score-search');
    if (searchInput) {
        searchInput.addEventListener('input', () => {
            clearTimeout(searchTimer);
            searchTimer = setTimeout(() => {
                currentSearch = searchInput.value.trim();
                currentPage = 1;
                loadData();
            }, 350);
        });
    }

    await loadData();
}

async function loadData() {
    const tbody = document.getElementById('scores-tbody');
    if (!tbody) return;
    tbody.innerHTML = '<tr><td colspan="6" class="table-loading"><div class="spinner"></div><span>Loading scores...</span></td></tr>';

    try {
        const params = new URLSearchParams({ page: currentPage, size: 20 });
        if (currentSearch) params.set('search', currentSearch);

        const data = await apiRequest(`/api/admin/scores?${params}`);
        const scores = data.scores || data || [];
        const total = data.total || scores.length;
        const page = data.page || 1;
        const pageSize = data.page_size || 20;

        document.getElementById('total-scores-count').textContent = total;
        renderTable(scores);
        renderPagination(total, page, pageSize);
    } catch (err) {
        tbody.innerHTML = `<tr><td colspan="6" class="table-error"><span>⚠️ ${err.message}</span></td></tr>`;
    }
}

function renderTable(scores) {
    const tbody = document.getElementById('scores-tbody');
    if (scores.length === 0) {
        tbody.innerHTML = '<tr><td colspan="6" class="table-empty"><div class="table-empty-icon">📋</div><p>No scores found</p></td></tr>';
        return;
    }

    tbody.innerHTML = scores.map(s => `
        <tr class="table-row">
            <td>
                <div class="table-user-cell">
                    <div class="table-avatar">${(s.user_name || '?').charAt(0).toUpperCase()}</div>
                    <div>
                        <span class="table-user-name">${esc(s.user_name || 'N/A')}</span>
                        <br><span class="text-muted text-xs">${esc(s.user_email || '')}</span>
                    </div>
                </div>
            </td>
            <td><span class="badge badge-role">${s.score} pts</span></td>
            <td>${s.round_date}</td>
            <td class="text-muted text-sm">${esc(s.notes || '-')}</td>
            <td class="table-date">${new Date(s.created_at).toLocaleDateString()}</td>
            <td>
                <button class="btn-table-action btn-toggle-deactivate delete-score-btn" data-id="${s.id}">Delete</button>
            </td>
        </tr>
    `).join('');

    tbody.querySelectorAll('.delete-score-btn').forEach(btn => {
        btn.addEventListener('click', async () => {
            if (!confirm('Delete this score?')) return;
            btn.disabled = true;
            btn.textContent = '...';
            try {
                await apiRequest(`/api/admin/scores/${btn.dataset.id}`, { method: 'DELETE' });
                loadData();
            } catch (err) {
                alert('Error: ' + err.message);
                btn.disabled = false;
                btn.textContent = 'Delete';
            }
        });
    });
}

function renderPagination(total, page, pageSize) {
    const container = document.getElementById('scores-pagination');
    const totalPages = Math.ceil(total / pageSize);
    if (totalPages <= 1) { container.innerHTML = ''; return; }

    const start = (page - 1) * pageSize + 1;
    const end = Math.min(page * pageSize, total);

    let btns = `<button class="pagination-btn ${page <= 1 ? 'disabled' : ''}" data-page="${page - 1}" ${page <= 1 ? 'disabled' : ''}>← Prev</button>`;
    for (let i = Math.max(1, page - 2); i <= Math.min(totalPages, page + 2); i++) {
        btns += `<button class="pagination-btn ${i === page ? 'active' : ''}" data-page="${i}">${i}</button>`;
    }
    btns += `<button class="pagination-btn ${page >= totalPages ? 'disabled' : ''}" data-page="${page + 1}" ${page >= totalPages ? 'disabled' : ''}>Next →</button>`;

    container.innerHTML = `
        <span class="pagination-info">Showing ${start}–${end} of ${total}</span>
        <div class="pagination-controls">${btns}</div>
    `;

    container.querySelectorAll('.pagination-btn:not(.disabled)').forEach(btn => {
        btn.addEventListener('click', () => {
            currentPage = parseInt(btn.dataset.page);
            loadData();
        });
    });
}

function esc(str) {
    if (!str) return '';
    const div = document.createElement('div');
    div.textContent = str;
    return div.innerHTML;
}
