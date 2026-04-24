// ============================================
// Dashboard Page Component (Protected)
// ============================================

import { getCurrentUser, getUser } from '/js/auth.js';
import { apiRequest } from '../utils/api.js';
/**
 * Renders dashboard content (without layout — layout is handled by router).
 * Shows different content based on user role.
 */
export function renderDashboardPage() {
    const user = getUser();
    const role = user?.role || 'user';
    const displayName = user?.name || 'User';

    if (role === 'admin') {
        return renderAdminDashboard(displayName);
    }
    return renderUserDashboard(displayName);
}

function renderUserDashboard(displayName) {
    return `
        <div class="welcome-section">
            <h2>Welcome back, ${displayName}!</h2>
            <p>Here's your Golf Score Lottery overview.</p>
        </div>

        <div class="stats-grid">
            <div class="stat-card">
                <div class="stat-icon">🏌️</div>
                <div class="stat-value" id="user-rounds">-</div>
                <div class="stat-label">Rounds Played</div>
            </div>
            <div class="stat-card">
                <div class="stat-icon">🎯</div>
                <div class="stat-value" id="user-best-score">-</div>
                <div class="stat-label">Best Score</div>
            </div>
            <div class="stat-card">
                <div class="stat-icon">🎟️</div>
                <div class="stat-value" id="user-entries">-</div>
                <div class="stat-label">Lottery Entries</div>
            </div>
            <div class="stat-card">
                <div class="stat-icon">🏆</div>
                <div class="stat-value" id="user-winnings">-</div>
                <div class="stat-label">Total Winnings</div>
            </div>
        </div>

        <div class="dashboard-grid-2col">
            <!-- Charts -->
            <div class="info-card">
                <h3>Score Trend</h3>
                <canvas id="scoreTrendChart"></canvas>
            </div>
            <div class="info-card">
                <h3>Charity Contributions</h3>
                <div style="height: 300px; display: flex; justify-content: center;">
                    <canvas id="charityContributionsChart"></canvas>
                </div>
            </div>

            <!-- Account Info -->
            <div class="info-card">
                <h3>Account Information</h3>
                <div id="user-info-loading" style="color: var(--color-text-muted); font-size: 0.875rem;">
                    Loading account details...
                </div>
                <div id="user-info-content" style="display:none;">
                    <div class="info-row">
                        <span class="info-label">Name</span>
                        <span class="info-value" id="info-name">—</span>
                    </div>
                    <div class="info-row">
                        <span class="info-label">Email</span>
                        <span class="info-value" id="info-email">—</span>
                    </div>
                    <div class="info-row">
                        <span class="info-label">Role</span>
                        <span class="info-value" id="info-role"></span>
                    </div>
                    <div class="info-row">
                        <span class="info-label">Subscription</span>
                        <span class="info-value" id="info-subscription"></span>
                    </div>
                    <div class="info-row">
                        <span class="info-label">Member Since</span>
                        <span class="info-value" id="info-created">—</span>
                    </div>
                </div>
            </div>

            <!-- Quick Actions -->
            <div class="info-card">
                <h3>Quick Actions</h3>
                <div class="quick-actions">
                    <a href="#/scores" class="quick-action-btn">
                        <span class="quick-action-icon">📈</span>
                        <span>Add New Score</span>
                    </a>
                    <a href="#/subscription" class="quick-action-btn">
                        <span class="quick-action-icon">💳</span>
                        <span>View Subscription</span>
                    </a>
                    <a href="#/charity" class="quick-action-btn">
                        <span class="quick-action-icon">❤️</span>
                        <span>Update Charity</span>
                    </a>
                </div>
            </div>
        </div>
    `;
}

function renderAdminDashboard(displayName) {
    return `
        <div class="welcome-section">
            <h2>Admin Panel</h2>
            <p>Welcome back, ${displayName}. Here's the system overview.</p>
        </div>

        <div class="stats-grid">
            <div class="stat-card">
                <div class="stat-icon">👥</div>
                <div class="stat-value" id="admin-total-users">-</div>
                <div class="stat-label">Total Users</div>
            </div>
            <div class="stat-card">
                <div class="stat-icon">💰</div>
                <div class="stat-value" id="admin-revenue">-</div>
                <div class="stat-label">Monthly Revenue</div>
            </div>
            <div class="stat-card">
                <div class="stat-icon">📋</div>
                <div class="stat-value" id="admin-active-subs">-</div>
                <div class="stat-label">Active Subscriptions</div>
            </div>
            <div class="stat-card">
                <div class="stat-icon">⏳</div>
                <div class="stat-value" id="admin-pending">-</div>
                <div class="stat-label">Pending Verifications</div>
            </div>
        </div>

        <div class="dashboard-grid-2col">
            <!-- Charts -->
            <div class="info-card">
                <h3>User Growth</h3>
                <canvas id="userGrowthChart"></canvas>
            </div>
            <div class="info-card">
                <h3>Revenue Trends</h3>
                <canvas id="revenueChart"></canvas>
            </div>

            <!-- System Info -->
            <div class="info-card">
                <h3>System Status</h3>
                <div id="user-info-loading" style="color: var(--color-text-muted); font-size: 0.875rem;">
                    Loading system data...
                </div>
                <div id="user-info-content" style="display:none;">
                    <div class="info-row">
                        <span class="info-label">Admin Name</span>
                        <span class="info-value" id="info-name">—</span>
                    </div>
                    <div class="info-row">
                        <span class="info-label">Admin Email</span>
                        <span class="info-value" id="info-email">—</span>
                    </div>
                    <div class="info-row">
                        <span class="info-label">Role</span>
                        <span class="info-value" id="info-role"></span>
                    </div>
                    <div class="info-row">
                        <span class="info-label">System Since</span>
                        <span class="info-value" id="info-created">—</span>
                    </div>
                </div>
            </div>

            <!-- Critical Actions -->
            <div class="info-card">
                <h3>Critical Actions</h3>
                <div class="quick-actions">
                    <a href="#/admin/draws" class="quick-action-btn admin-action">
                        <span class="quick-action-icon">🔀</span>
                        <span>Run Monthly Draw</span>
                    </a>
                    <a href="#/admin/verify-winners" class="quick-action-btn admin-action">
                        <span class="quick-action-icon">🏆</span>
                        <span>Verify Winners</span>
                    </a>
                    <a href="#/admin/reports" class="quick-action-btn admin-action">
                        <span class="quick-action-icon">📉</span>
                        <span>Generate Reports</span>
                    </a>
                </div>
            </div>
        </div>
    `;
}

// Chart.js global defaults
if (window.Chart) {
    Chart.defaults.color = '#94a3b8';
    Chart.defaults.font.family = "'Inter', sans-serif";
    Chart.defaults.plugins.tooltip.backgroundColor = 'rgba(15, 23, 42, 0.9)';
    Chart.defaults.plugins.tooltip.titleColor = '#f1f5f9';
    Chart.defaults.plugins.tooltip.bodyColor = '#cbd5e1';
    Chart.defaults.plugins.tooltip.borderColor = '#334155';
    Chart.defaults.plugins.tooltip.borderWidth = 1;
}

export function initDashboardPage() {
    const user = getUser();
    const role = user?.role || 'user';

    // Fetch fresh user data from API
    loadUserInfo();
    
    setTimeout(() => {
        if (role === 'admin') {
            loadAdminStats();
            initAdminCharts();
        } else {
            loadUserStats();
            initUserCharts();
        }
    }, 100);
}

const formatter = new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' });

async function loadUserStats() {
    try {
        const stats = await apiRequest('/api/stats/dashboard');
        if (stats) {
            document.getElementById('user-rounds').textContent = stats.rounds_played;
            document.getElementById('user-best-score').textContent = stats.best_score || '-';
            document.getElementById('user-entries').textContent = stats.lottery_entries;
            document.getElementById('user-winnings').textContent = formatter.format(stats.total_winnings);
        }
    } catch (err) {
        console.error('Failed to load user stats:', err);
    }
}

async function loadAdminStats() {
    try {
        const stats = await apiRequest('/api/admin/stats/dashboard');
        if (stats) {
            document.getElementById('admin-total-users').textContent = stats.total_users;
            document.getElementById('admin-revenue').textContent = formatter.format(stats.total_revenue);
            document.getElementById('admin-active-subs').textContent = stats.active_subscriptions;
            document.getElementById('admin-pending').textContent = stats.pending_verifications;
        }
    } catch (err) {
        console.error('Failed to load admin stats:', err);
    }
}

async function initUserCharts() {
    try {
        const [trendData, charityData] = await Promise.all([
            apiRequest('/api/stats/score-trend'),
            apiRequest('/api/stats/charity-distribution')
        ]);

        const scoreCtx = document.getElementById('scoreTrendChart');
        const charityCtx = document.getElementById('charityContributionsChart');

        if (scoreCtx && window.Chart && trendData?.length > 0) {
            const primaryColor = getComputedStyle(document.documentElement).getPropertyValue('--bg-primary').trim() || '#FF4F00';
            new Chart(scoreCtx, {
                type: 'line',
                data: {
                    labels: trendData.map(d => d.label),
                    datasets: [{
                        label: 'Stableford Score',
                        data: trendData.map(d => d.value),
                        borderColor: primaryColor,
                        backgroundColor: primaryColor + '20',
                        borderWidth: 2,
                        tension: 0.4,
                        fill: true
                    }]
                },
                options: {
                    responsive: true,
                    plugins: { legend: { display: false } },
                    scales: {
                        y: { grid: { color: 'rgba(255,255,255,0.05)' }, max: 45 },
                        x: { grid: { color: 'rgba(255,255,255,0.05)' } }
                    }
                }
            });
        }

        if (charityCtx && window.Chart && charityData?.length > 0) {
            new Chart(charityCtx, {
                type: 'doughnut',
                data: {
                    labels: charityData.map(d => d.charity_name),
                    datasets: [{
                        data: charityData.map(d => d.percentage),
                        backgroundColor: ['#ca8a04', '#65a30d', '#0284c7', '#9333ea', '#e11d48'],
                        borderWidth: 0,
                        hoverOffset: 4
                    }]
                },
                options: {
                    responsive: true,
                    maintainAspectRatio: false,
                    plugins: {
                        legend: { position: 'bottom' }
                    },
                    cutout: '70%'
                }
            });
        }
    } catch (err) {
        console.error('Failed to init charts:', err);
    }
}

async function initAdminCharts() {
    try {
        const [growthData, revenueData] = await Promise.all([
            apiRequest('/api/admin/stats/user-growth'),
            apiRequest('/api/admin/stats/revenue')
        ]);

        const userCtx = document.getElementById('userGrowthChart');
        const revenueCtx = document.getElementById('revenueChart');

        if (userCtx && window.Chart && growthData?.length > 0) {
            new Chart(userCtx, {
                type: 'line',
                data: {
                    labels: growthData.map(d => d.label),
                    datasets: [{
                        label: 'New Users',
                        data: growthData.map(d => d.value),
                        borderColor: '#0284c7',
                        backgroundColor: 'rgba(2, 132, 199, 0.1)',
                        borderWidth: 2,
                        tension: 0.4,
                        fill: true
                    }]
                },
                options: {
                    responsive: true,
                    plugins: { legend: { display: false } },
                    scales: {
                        y: { grid: { color: 'rgba(255,255,255,0.05)' } },
                        x: { grid: { color: 'rgba(255,255,255,0.05)' } }
                    }
                }
            });
        }

        if (revenueCtx && window.Chart && revenueData?.length > 0) {
            new Chart(revenueCtx, {
                type: 'bar',
                data: {
                    labels: revenueData.map(d => d.label),
                    datasets: [{
                        label: 'Revenue ($)',
                        data: revenueData.map(d => d.value),
                        backgroundColor: '#ca8a04',
                        borderRadius: 4
                    }]
                },
                options: {
                    responsive: true,
                    plugins: { legend: { display: false } },
                    scales: {
                        y: { grid: { color: 'rgba(255,255,255,0.05)' } },
                        x: { grid: { display: false } }
                    }
                }
            });
        }
    } catch (err) {
        console.error('Failed to load admin charts:', err);
    }
}


async function loadUserInfo() {
    const loadingEl = document.getElementById('user-info-loading');
    const contentEl = document.getElementById('user-info-content');

    if (!loadingEl || !contentEl) return;

    try {
        const user = await getCurrentUser();

        document.getElementById('info-name').textContent = user.name;
        document.getElementById('info-email').textContent = user.email;

        // Role badge
        document.getElementById('info-role').innerHTML =
            `<span class="badge badge-role">${user.role.toUpperCase()}</span>`;

        // Subscription badge (only for user dashboard)
        const subEl = document.getElementById('info-subscription');
        if (subEl) {
            subEl.innerHTML = user.subscription_active
                ? '<span class="badge badge-active">✓ Active</span>'
                : '<span class="badge badge-inactive">Inactive</span>';
        }

        // Member since
        const date = new Date(user.created_at);
        document.getElementById('info-created').textContent = date.toLocaleDateString('en-US', {
            year: 'numeric',
            month: 'long',
            day: 'numeric',
        });

        // Update welcome section with fresh name
        const welcomeH2 = document.querySelector('.welcome-section h2');
        if (welcomeH2 && user.role !== 'admin') {
            welcomeH2.textContent = `Welcome back, ${user.name}!`;
        }

        loadingEl.style.display = 'none';
        contentEl.style.display = 'block';
    } catch (err) {
        loadingEl.textContent = 'Failed to load account details.';
        loadingEl.style.color = 'var(--color-error)';
    }
}
