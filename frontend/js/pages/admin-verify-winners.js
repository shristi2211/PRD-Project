import { apiRequest } from '../utils/api.js';

export function renderAdminVerifyWinners() {
    return `
        <div class="admin-users-page">
            <div class="admin-users-header">
                <div class="admin-users-title-group">
                    <h2>Verify Winners</h2>
                    <p>Review proof submissions and approve or reject claims</p>
                </div>
                <div class="admin-users-stat">
                    <span class="admin-users-stat-value" id="pending-count">—</span>
                    <span class="admin-users-stat-label">Pending</span>
                </div>
            </div>

            <div class="admin-table-container">
                <table class="admin-table">
                    <thead>
                        <tr>
                            <th>Winner</th>
                            <th>Draw</th>
                            <th>Prize</th>
                            <th>Proof</th>
                            <th>Status</th>
                            <th>Actions</th>
                        </tr>
                    </thead>
                    <tbody id="winners-tbody">
                        <tr><td colspan="6" class="table-loading"><div class="spinner"></div><span>Loading...</span></td></tr>
                    </tbody>
                </table>
            </div>
        </div>

        <div id="verify-modal" class="modal">
            <div class="modal-content">
                <h2 id="modal-title">Verify Winner</h2>
                <div id="modal-body"></div>
            </div>
        </div>
    `;
}

export async function initAdminVerifyWinners() {
    await loadWinners();
}

async function loadWinners() {
    const tbody = document.getElementById('winners-tbody');
    try {
        const winners = await apiRequest('/api/admin/winners/pending');
        const list = winners || [];
        document.getElementById('pending-count').textContent = list.length;
        renderWinners(list);
    } catch (err) {
        tbody.innerHTML = `<tr><td colspan="6" class="table-error"><span>⚠️ ${err.message}</span></td></tr>`;
    }
}

function renderWinners(winners) {
    const tbody = document.getElementById('winners-tbody');
    if (winners.length === 0) {
        tbody.innerHTML = '<tr><td colspan="6" class="table-empty"><div class="table-empty-icon">🏆</div><p>No pending verifications</p><span>All winners have been verified</span></td></tr>';
        return;
    }

    const fmt = new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' });

    tbody.innerHTML = winners.map(w => `
        <tr class="table-row">
            <td>
                <div class="table-user-cell">
                    <div class="table-avatar">${(w.user_name || '?').charAt(0).toUpperCase()}</div>
                    <div>
                        <span class="table-user-name">${esc(w.user_name)}</span>
                        <br><span class="text-muted text-xs">${esc(w.user_email)}</span>
                    </div>
                </div>
            </td>
            <td>${w.draw_month}/${w.draw_year}</td>
            <td><strong style="color:var(--color-accent-light)">${fmt.format(w.prize_amount)}</strong></td>
            <td>
                ${w.proof_url ? `<a href="${w.proof_url}" target="_blank" class="text-sm" style="color:var(--color-accent-light)">View Proof ↗</a>` : '<span class="text-muted">None</span>'}
                ${w.proof_notes ? `<br><span class="text-xs text-muted">${esc(w.proof_notes)}</span>` : ''}
            </td>
            <td><span class="status-badge status-${w.verification_status}">${w.verification_status}</span></td>
            <td>
                <div style="display:flex;gap:0.5rem">
                    <button class="btn-table-action btn-toggle-activate approve-btn" data-id="${w.id}">✓ Approve</button>
                    <button class="btn-table-action btn-toggle-deactivate reject-btn" data-id="${w.id}">✗ Reject</button>
                </div>
            </td>
        </tr>
    `).join('');

    tbody.querySelectorAll('.approve-btn').forEach(btn => {
        btn.addEventListener('click', () => verifyWinner(btn.dataset.id, 'approved', ''));
    });

    tbody.querySelectorAll('.reject-btn').forEach(btn => {
        btn.addEventListener('click', () => {
            const reason = prompt('Rejection reason:');
            if (reason !== null) {
                verifyWinner(btn.dataset.id, 'rejected', reason);
            }
        });
    });
}

async function verifyWinner(id, status, reason) {
    try {
        await apiRequest(`/api/admin/winners/${id}/verify`, {
            method: 'PUT',
            body: JSON.stringify({ status, rejection_reason: reason })
        });
        loadWinners();
    } catch (err) {
        alert('Error: ' + err.message);
    }
}

function esc(str) {
    if (!str) return '';
    const div = document.createElement('div');
    div.textContent = str;
    return div.innerHTML;
}
