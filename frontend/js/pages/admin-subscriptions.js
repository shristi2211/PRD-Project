import { apiRequest } from '/js/utils/api.js';

export function renderAdminSubscriptions() {
    return `
        <div class="admin-subscriptions list-card">
            <header class="page-header">
                <div class="header-content">
                    <h1>Subscription Tracking</h1>
                    <p class="subtitle">Monitor all user subscription plans and statuses.</p>
                </div>
            </header>

            <div class="dashboard-card">
                <div class="filter-bar" style="display: flex; gap: var(--space-md); margin-bottom: var(--space-xl); flex-wrap: wrap; align-items: center;">
                    <input type="text" id="sub-search" placeholder="Search by name or email..." class="form-input" style="flex: 1; min-width: 200px;" />
                    <select id="sub-plan-filter" class="form-input" style="width: 180px;">
                        <option value="">All Plans</option>
                        <option value="free">Free</option>
                        <option value="monthly">Monthly</option>
                        <option value="yearly">Yearly</option>
                    </select>
                    <select id="sub-status-filter" class="form-input" style="width: 180px;">
                        <option value="">All Statuses</option>
                        <option value="active">Active</option>
                        <option value="inactive">Inactive</option>
                    </select>
                </div>

                <div id="sub-loading" style="text-align: center; padding: var(--space-2xl);">
                    <div class="loading-spinner"></div>
                </div>

                <div id="sub-table-container" style="display: none;">
                    <div style="overflow-x: auto;">
                        <table class="data-table" id="sub-table">
                            <thead>
                                <tr>
                                    <th>User</th>
                                    <th>Email</th>
                                    <th>Plan</th>
                                    <th>Status</th>
                                    <th>Joined</th>
                                </tr>
                            </thead>
                            <tbody id="sub-tbody"></tbody>
                        </table>
                    </div>

                    <div class="pagination" id="sub-pagination"></div>
                </div>

                <div id="sub-empty" style="display: none; text-align: center; padding: var(--space-2xl);">
                    <p class="text-muted">No subscription records found.</p>
                </div>
            </div>
        </div>
    `;
}

export async function initAdminSubscriptions() {
    let currentPage = 1;
    const pageSize = 10;
    let searchTimeout = null;

    const searchInput = document.getElementById('sub-search');
    const planFilter = document.getElementById('sub-plan-filter');
    const statusFilter = document.getElementById('sub-status-filter');

    async function loadData() {
        const loading = document.getElementById('sub-loading');
        const tableContainer = document.getElementById('sub-table-container');
        const emptyState = document.getElementById('sub-empty');

        loading.style.display = 'block';
        tableContainer.style.display = 'none';
        emptyState.style.display = 'none';

        try {
            const search = searchInput.value.trim();
            const status = statusFilter.value;
            const plan = planFilter.value;

            let url = `/api/admin/subscriptions?page=${currentPage}&size=${pageSize}`;
            if (search) url += `&search=${encodeURIComponent(search)}`;
            if (status) url += `&status=${status}`;

            const data = await apiRequest(url);
            const users = data.users || [];
            const total = data.total || 0;

            // Client-side plan filter (since backend doesn't have plan filter yet)
            const filtered = plan ? users.filter(u => (u.subscription_type || 'free') === plan) : users;

            loading.style.display = 'none';

            if (filtered.length === 0) {
                emptyState.style.display = 'block';
                return;
            }

            tableContainer.style.display = 'block';
            const tbody = document.getElementById('sub-tbody');
            tbody.innerHTML = filtered.map(user => {
                const planType = user.subscription_type || 'free';
                const isActive = user.subscription_active;
                const planBadgeClass = planType === 'free' ? 'badge-inactive' : (planType === 'monthly' ? 'badge-info' : 'badge-success');
                const planLabel = planType.charAt(0).toUpperCase() + planType.slice(1);
                const statusBadge = isActive
                    ? '<span class="badge badge-success">Active</span>'
                    : '<span class="badge badge-inactive">Inactive</span>';
                const joinDate = new Date(user.created_at).toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' });

                return `
                    <tr>
                        <td><strong>${escapeHtml(user.name)}</strong></td>
                        <td>${escapeHtml(user.email)}</td>
                        <td><span class="badge ${planBadgeClass}">${planLabel}</span></td>
                        <td>${statusBadge}</td>
                        <td>${joinDate}</td>
                    </tr>
                `;
            }).join('');

            // Pagination
            const totalPages = Math.ceil(total / pageSize);
            const paginationEl = document.getElementById('sub-pagination');
            if (totalPages > 1) {
                let paginationHtml = '';
                for (let i = 1; i <= totalPages; i++) {
                    paginationHtml += `<button class="btn btn-sm ${i === currentPage ? 'btn-primary' : 'btn-outline'}" data-page="${i}">${i}</button>`;
                }
                paginationEl.innerHTML = paginationHtml;
                paginationEl.querySelectorAll('button').forEach(btn => {
                    btn.addEventListener('click', () => {
                        currentPage = parseInt(btn.dataset.page);
                        loadData();
                    });
                });
            } else {
                paginationEl.innerHTML = '';
            }
        } catch (err) {
            loading.style.display = 'none';
            console.error('Failed to load subscriptions:', err);
        }
    }

    // Event listeners
    searchInput.addEventListener('input', () => {
        clearTimeout(searchTimeout);
        searchTimeout = setTimeout(() => { currentPage = 1; loadData(); }, 300);
    });
    planFilter.addEventListener('change', () => { currentPage = 1; loadData(); });
    statusFilter.addEventListener('change', () => { currentPage = 1; loadData(); });

    loadData();
}

function escapeHtml(str) {
    const div = document.createElement('div');
    div.textContent = str;
    return div.innerHTML;
}
