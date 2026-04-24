import { apiRequest } from '../utils/api.js';

export function renderAdminCharities() {
    return `
        <div class="admin-users-page">
            <div class="admin-users-header">
                <div class="admin-users-title-group">
                    <h2>Charity Management</h2>
                    <p>Add, edit, and manage registered charities</p>
                </div>
                <div style="display:flex;gap:0.75rem;align-items:center">
                    <div class="admin-users-stat">
                        <span class="admin-users-stat-value" id="charity-count">—</span>
                        <span class="admin-users-stat-label">Charities</span>
                    </div>
                    <button class="btn-generate" id="add-charity-btn">+ Add Charity</button>
                </div>
            </div>

            <div class="admin-table-container">
                <table class="admin-table">
                    <thead>
                        <tr>
                            <th>Charity</th>
                            <th>Website</th>
                            <th>Status</th>
                            <th>Created</th>
                            <th>Actions</th>
                        </tr>
                    </thead>
                    <tbody id="charities-tbody">
                        <tr><td colspan="5" class="table-loading"><div class="spinner"></div><span>Loading charities...</span></td></tr>
                    </tbody>
                </table>
            </div>
        </div>

        <div id="charity-modal" class="modal">
            <div class="modal-content">
                <h2 id="charity-modal-title">Add Charity</h2>
                <form id="charity-form" class="standard-form">
                    <div id="charity-form-error" class="error-message" style="display:none"></div>
                    <div class="form-group">
                        <label for="charity-name">Name</label>
                        <input type="text" id="charity-name" required placeholder="Charity name">
                    </div>
                    <div class="form-group">
                        <label for="charity-desc">Description</label>
                        <textarea id="charity-desc" rows="3" placeholder="Brief description"></textarea>
                    </div>
                    <div class="form-group">
                        <label for="charity-website">Website</label>
                        <input type="url" id="charity-website" placeholder="https://...">
                    </div>
                    <div class="form-group">
                        <label for="charity-logo">Logo URL</label>
                        <input type="url" id="charity-logo" placeholder="https://logo-url...">
                    </div>
                    <div class="modal-actions" style="flex-direction:row;gap:0.75rem">
                        <button type="button" class="btn btn-ghost btn-sm" id="cancel-charity">Cancel</button>
                        <button type="submit" class="btn btn-primary btn-sm" id="save-charity">Save</button>
                    </div>
                </form>
            </div>
        </div>
    `;
}

export async function initAdminCharities() {
    document.getElementById('add-charity-btn').addEventListener('click', () => openModal('add'));
    document.getElementById('cancel-charity').addEventListener('click', closeModal);
    document.getElementById('charity-form').addEventListener('submit', handleSave);
    await loadCharities();
}

let editingId = null;

async function loadCharities() {
    const tbody = document.getElementById('charities-tbody');
    try {
        const charities = await apiRequest('/api/charities');
        const list = charities || [];
        document.getElementById('charity-count').textContent = list.length;
        renderTable(list);
    } catch (err) {
        tbody.innerHTML = `<tr><td colspan="5" class="table-error"><span>⚠️ ${err.message}</span></td></tr>`;
    }
}

function renderTable(charities) {
    const tbody = document.getElementById('charities-tbody');
    if (charities.length === 0) {
        tbody.innerHTML = '<tr><td colspan="5" class="table-empty"><div class="table-empty-icon">🏛️</div><p>No charities yet</p><span>Click "Add Charity" to create one</span></td></tr>';
        return;
    }

    tbody.innerHTML = charities.map(c => `
        <tr class="table-row">
            <td>
                <div class="table-user-cell">
                    <div class="table-avatar" style="background:linear-gradient(135deg,#ca8a04,#f59e0b)">🏛</div>
                    <div>
                        <span class="table-user-name">${esc(c.name)}</span>
                        <br><span class="text-muted text-xs">${esc(c.description || '').substring(0, 60)}</span>
                    </div>
                </div>
            </td>
            <td>${c.website ? `<a href="${c.website}" target="_blank" style="color:var(--color-accent-light)">${c.website.replace('https://', '').substring(0, 30)}</a>` : '-'}</td>
            <td>${c.active ? '<span class="badge badge-active">Active</span>' : '<span class="badge badge-inactive">Inactive</span>'}</td>
            <td class="table-date">${new Date(c.created_at).toLocaleDateString()}</td>
            <td>
                <div style="display:flex;gap:0.5rem">
                    <button class="btn-table-action btn-toggle-activate edit-btn" data-id="${c.id}" data-name="${escAttr(c.name)}" data-desc="${escAttr(c.description || '')}" data-website="${escAttr(c.website || '')}" data-logo="${escAttr(c.logo_url || '')}">Edit</button>
                    <button class="btn-table-action ${c.active ? 'btn-toggle-deactivate' : 'btn-toggle-activate'} toggle-btn" data-id="${c.id}" data-active="${c.active}">${c.active ? 'Deactivate' : 'Activate'}</button>
                </div>
            </td>
        </tr>
    `).join('');

    tbody.querySelectorAll('.edit-btn').forEach(btn => {
        btn.addEventListener('click', () => {
            openModal('edit', {
                id: btn.dataset.id,
                name: btn.dataset.name,
                description: btn.dataset.desc,
                website: btn.dataset.website,
                logo_url: btn.dataset.logo
            });
        });
    });

    tbody.querySelectorAll('.toggle-btn').forEach(btn => {
        btn.addEventListener('click', async () => {
            const active = btn.dataset.active !== 'true';
            btn.disabled = true;
            btn.textContent = '...';
            try {
                await apiRequest(`/api/admin/charities/${btn.dataset.id}/toggle`, {
                    method: 'PUT',
                    body: JSON.stringify({ active })
                });
                loadCharities();
            } catch (err) {
                alert('Error: ' + err.message);
                btn.disabled = false;
            }
        });
    });
}

function openModal(mode, data = {}) {
    editingId = mode === 'edit' ? data.id : null;
    document.getElementById('charity-modal-title').textContent = mode === 'edit' ? 'Edit Charity' : 'Add Charity';
    document.getElementById('charity-name').value = data.name || '';
    document.getElementById('charity-desc').value = data.description || '';
    document.getElementById('charity-website').value = data.website || '';
    document.getElementById('charity-logo').value = data.logo_url || '';
    document.getElementById('charity-form-error').style.display = 'none';
    document.getElementById('charity-modal').classList.add('active');
}

function closeModal() {
    document.getElementById('charity-modal').classList.remove('active');
    document.getElementById('charity-form').reset();
    editingId = null;
}

async function handleSave(e) {
    e.preventDefault();
    const errEl = document.getElementById('charity-form-error');
    const btn = document.getElementById('save-charity');
    errEl.style.display = 'none';
    btn.disabled = true;
    btn.textContent = 'Saving...';

    const body = {
        name: document.getElementById('charity-name').value,
        description: document.getElementById('charity-desc').value,
        website: document.getElementById('charity-website').value,
        logo_url: document.getElementById('charity-logo').value,
    };

    try {
        if (editingId) {
            await apiRequest(`/api/admin/charities/${editingId}`, { method: 'PUT', body: JSON.stringify(body) });
        } else {
            await apiRequest('/api/admin/charities', { method: 'POST', body: JSON.stringify(body) });
        }
        closeModal();
        loadCharities();
    } catch (err) {
        errEl.textContent = err.message;
        errEl.style.display = 'block';
    } finally {
        btn.disabled = false;
        btn.textContent = 'Save';
    }
}

function esc(str) { if (!str) return ''; const d = document.createElement('div'); d.textContent = str; return d.innerHTML; }
function escAttr(str) { if (!str) return ''; return str.replace(/"/g, '&quot;').replace(/'/g, '&#39;'); }
