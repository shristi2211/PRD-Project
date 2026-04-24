import { apiRequest } from '../utils/api.js';

let currentReportType = 'users';
let reportData = [];

export function renderAdminReports() {
    return `
        <div class="admin-users-page">
            <div class="admin-users-header">
                <div class="admin-users-title-group">
                    <h2>System Reports</h2>
                    <p>Generate and export detailed reports</p>
                </div>
            </div>

            <div class="report-controls">
                <div class="form-group">
                    <label for="report-type">Report Type</label>
                    <select id="report-type">
                        <option value="users" selected>User Summary</option>
                        <option value="revenue">Revenue Breakdown</option>
                        <option value="draws">Draw History</option>
                        <option value="charities">Charity Allocation</option>
                    </select>
                </div>
                <div class="form-group">
                    <label for="report-from">From Date</label>
                    <input type="date" id="report-from">
                </div>
                <div class="form-group">
                    <label for="report-to">To Date</label>
                    <input type="date" id="report-to">
                </div>
                <div class="report-actions">
                    <button class="btn-generate" id="generate-btn">📊 Generate</button>
                    <button class="btn-export" id="export-btn" disabled>📥 Export CSV</button>
                </div>
            </div>

            <div id="report-summary" class="report-summary" style="display:none"></div>

            <div id="report-table-area">
                <div class="empty-state">
                    <p style="font-size:48px;margin-bottom:1rem">📉</p>
                    <p style="color:var(--color-text-secondary);font-weight:600">Select a report type and click Generate</p>
                    <p>Reports will appear here with data preview and CSV export</p>
                </div>
            </div>
        </div>
    `;
}

export async function initAdminReports() {
    document.getElementById('generate-btn').addEventListener('click', generateReport);
    document.getElementById('export-btn').addEventListener('click', exportCSV);
    document.getElementById('report-type').addEventListener('change', () => {
        document.getElementById('export-btn').disabled = true;
        reportData = [];
    });
}

async function generateReport() {
    const type = document.getElementById('report-type').value;
    const from = document.getElementById('report-from').value;
    const to = document.getElementById('report-to').value;
    currentReportType = type;

    const btn = document.getElementById('generate-btn');
    const tableArea = document.getElementById('report-table-area');
    btn.disabled = true;
    btn.textContent = '⏳ Loading...';
    tableArea.innerHTML = '<div style="padding:3rem;text-align:center"><div class="spinner"></div></div>';

    try {
        let params = '';
        if (from && to) params = `?from=${from}&to=${to}`;

        const endpoint = `/api/admin/reports/${type}${type === 'charities' ? '' : params}`;
        reportData = await apiRequest(endpoint) || [];

        renderSummary(type, reportData);
        renderTable(type, reportData);

        document.getElementById('export-btn').disabled = reportData.length === 0;
    } catch (err) {
        tableArea.innerHTML = `<div class="empty-state">⚠️ Error: ${err.message}</div>`;
        document.getElementById('report-summary').style.display = 'none';
    } finally {
        btn.disabled = false;
        btn.textContent = '📊 Generate';
    }
}

function renderSummary(type, data) {
    const container = document.getElementById('report-summary');
    if (data.length === 0) {
        container.style.display = 'none';
        return;
    }

    let cards = '';
    if (type === 'users') {
        const total = data.length;
        const active = data.filter(d => d.subscription_active).length;
        const totalWinnings = data.reduce((s, d) => s + d.total_winnings, 0);
        cards = `
            <div class="report-summary-card"><div class="report-summary-value">${total}</div><div class="report-summary-label">Total Users</div></div>
            <div class="report-summary-card"><div class="report-summary-value">${active}</div><div class="report-summary-label">Active Subs</div></div>
            <div class="report-summary-card"><div class="report-summary-value">$${totalWinnings.toFixed(2)}</div><div class="report-summary-label">Total Winnings</div></div>
        `;
    } else if (type === 'revenue') {
        const totalPool = data.reduce((s, d) => s + d.total_pool, 0);
        const totalFee = data.reduce((s, d) => s + d.platform_fee, 0);
        const totalCharity = data.reduce((s, d) => s + d.charity_amount, 0);
        cards = `
            <div class="report-summary-card"><div class="report-summary-value">$${totalPool.toFixed(2)}</div><div class="report-summary-label">Total Pool</div></div>
            <div class="report-summary-card"><div class="report-summary-value">$${totalFee.toFixed(2)}</div><div class="report-summary-label">Platform Revenue</div></div>
            <div class="report-summary-card"><div class="report-summary-value">$${totalCharity.toFixed(2)}</div><div class="report-summary-label">Charity Amount</div></div>
        `;
    } else if (type === 'draws') {
        const completed = data.filter(d => d.status === 'completed').length;
        const totalEntries = data.reduce((s, d) => s + d.total_entries, 0);
        cards = `
            <div class="report-summary-card"><div class="report-summary-value">${data.length}</div><div class="report-summary-label">Total Draws</div></div>
            <div class="report-summary-card"><div class="report-summary-value">${completed}</div><div class="report-summary-label">Completed</div></div>
            <div class="report-summary-card"><div class="report-summary-value">${totalEntries}</div><div class="report-summary-label">Total Entries</div></div>
        `;
    } else if (type === 'charities') {
        const totalUsers = data.reduce((s, d) => s + d.total_users, 0);
        cards = `
            <div class="report-summary-card"><div class="report-summary-value">${data.length}</div><div class="report-summary-label">Total Charities</div></div>
            <div class="report-summary-card"><div class="report-summary-value">${totalUsers}</div><div class="report-summary-label">Users Selecting</div></div>
        `;
    }

    container.innerHTML = cards;
    container.style.display = 'grid';
}

function renderTable(type, data) {
    const tableArea = document.getElementById('report-table-area');

    if (data.length === 0) {
        tableArea.innerHTML = '<div class="empty-state">No data found for the selected criteria.</div>';
        return;
    }

    const configs = {
        users: {
            headers: ['Name', 'Email', 'Role', 'Subscription', 'Rounds', 'Best Score', 'Winnings', 'Joined'],
            row: d => `
                <td><span class="table-user-name">${esc(d.name)}</span></td>
                <td class="table-email">${esc(d.email)}</td>
                <td><span class="badge badge-role">${d.role.toUpperCase()}</span></td>
                <td>${d.subscription_active ? '<span class="badge badge-active">Active</span>' : '<span class="badge badge-inactive">Inactive</span>'}</td>
                <td>${d.rounds_played}</td>
                <td>${d.best_score}</td>
                <td>$${d.total_winnings.toFixed(2)}</td>
                <td class="table-date">${new Date(d.created_at).toLocaleDateString()}</td>
            `
        },
        revenue: {
            headers: ['Month', 'Year', 'Total Pool', 'Winner Prize', 'Charity', 'Platform Fee', 'Entries'],
            row: d => `
                <td>${d.month}</td>
                <td>${d.year}</td>
                <td>$${d.total_pool.toFixed(2)}</td>
                <td>$${d.winner_prize.toFixed(2)}</td>
                <td>$${d.charity_amount.toFixed(2)}</td>
                <td>$${d.platform_fee.toFixed(2)}</td>
                <td>${d.total_entries}</td>
            `
        },
        draws: {
            headers: ['Date', 'Month/Year', 'Status', 'Pool', 'Winner', 'Winner Email', 'Entries'],
            row: d => `
                <td class="table-date">${new Date(d.draw_date).toLocaleDateString()}</td>
                <td>${d.month}/${d.year}</td>
                <td><span class="badge ${d.status === 'completed' ? 'badge-active' : 'badge-inactive'}">${d.status}</span></td>
                <td>$${d.total_pool.toFixed(2)}</td>
                <td>${esc(d.winner_name)}</td>
                <td class="table-email">${esc(d.winner_email)}</td>
                <td>${d.total_entries}</td>
            `
        },
        charities: {
            headers: ['Charity', 'Website', 'Active', 'Users', 'Avg %', 'Total Received'],
            row: d => `
                <td><span class="table-user-name">${esc(d.charity_name)}</span></td>
                <td>${d.website ? `<a href="${d.website}" target="_blank" class="text-sm">${d.website}</a>` : '-'}</td>
                <td>${d.active ? '<span class="badge badge-active">Yes</span>' : '<span class="badge badge-inactive">No</span>'}</td>
                <td>${d.total_users}</td>
                <td>${d.avg_percentage}%</td>
                <td>$${d.total_received.toFixed(2)}</td>
            `
        }
    };

    const cfg = configs[type];
    tableArea.innerHTML = `
        <div class="admin-table-container">
            <table class="admin-table">
                <thead><tr>${cfg.headers.map(h => `<th>${h}</th>`).join('')}</tr></thead>
                <tbody>${data.map(d => `<tr class="table-row">${cfg.row(d)}</tr>`).join('')}</tbody>
            </table>
        </div>
    `;
}

function exportCSV() {
    if (reportData.length === 0) return;

    const keys = Object.keys(reportData[0]);
    const header = keys.join(',');
    const rows = reportData.map(row =>
        keys.map(k => {
            let val = row[k];
            if (val === null || val === undefined) val = '';
            val = String(val).replace(/"/g, '""');
            return `"${val}"`;
        }).join(',')
    );

    const csv = [header, ...rows].join('\n');
    const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `${currentReportType}_report_${new Date().toISOString().slice(0, 10)}.csv`;
    a.click();
    URL.revokeObjectURL(url);
}

function esc(str) {
    if (!str) return '';
    const div = document.createElement('div');
    div.textContent = str;
    return div.innerHTML;
}
