import { apiRequest } from '../utils/api.js';

export function renderAdminDraws() {
    return `
        <div class="admin-users-page">
            <div class="admin-users-header">
                <div class="admin-users-title-group">
                    <h2>Lottery Draws</h2>
                    <p>Run monthly draws and view past results</p>
                </div>
                <div style="display:flex;gap:0.75rem;align-items:center">
                    <button class="btn-generate" id="new-draw-btn">▶ Run New Draw</button>
                </div>
            </div>

            <div class="admin-table-container">
                <table class="admin-table">
                    <thead>
                        <tr>
                            <th>Month/Year</th>
                            <th>Date</th>
                            <th>Total Pool</th>
                            <th>Winner Prize (60%)</th>
                            <th>Status</th>
                            <th>Entries</th>
                            <th>Actions</th>
                        </tr>
                    </thead>
                    <tbody id="draws-table-body">
                        <tr><td colspan="7" class="table-loading"><div class="spinner"></div><span>Loading...</span></td></tr>
                    </tbody>
                </table>
            </div>
        </div>

        <!-- Run Draw Modal -->
        <div id="run-draw-modal" class="modal">
            <div class="modal-content">
                <h2>Run Monthly Draw</h2>
                
                <form id="run-draw-form" class="standard-form">
                    <div id="draw-error" class="error-message" style="display: none;"></div>
                    
                    <div class="form-row">
                        <div class="form-group">
                            <label>Month</label>
                            <select id="draw-month" required>
                                ${Array.from({length: 12}, (_, i) => `<option value="${i+1}" ${i+1 === new Date().getMonth()+1 ? 'selected' : ''}>${new Date(2000, i).toLocaleString('default', { month: 'long' })}</option>`).join('')}
                            </select>
                        </div>
                        <div class="form-group">
                            <label>Year</label>
                            <input type="number" id="draw-year" required value="${new Date().getFullYear()}">
                        </div>
                    </div>
                    
                    <div class="form-group">
                        <label>Total Pool Amount ($)</label>
                        <input type="number" id="base-pool" required min="1" step="0.01" placeholder="e.g. 1000.00">
                    </div>

                    <div class="simulation-results card bg-dark mt-4" id="simulation-panel" style="display: none;">
                        <h3>Simulation Results</h3>
                        <div class="grid grid-3 mt-2">
                            <div class="stat-box">
                                <div class="text-muted">Eligible Users</div>
                                <div class="text-xl" id="sim-users">-</div>
                            </div>
                            <div class="stat-box">
                                <div class="text-muted">Winner Prize</div>
                                <div class="text-xl text-success" id="sim-winner">-</div>
                            </div>
                            <div class="stat-box">
                                <div class="text-muted">Platform Fee</div>
                                <div class="text-xl" id="sim-platform">-</div>
                            </div>
                        </div>
                        <p class="text-sm text-muted mt-2">This is a dry-run. Proceed to execute the actual crypto-random draw.</p>
                    </div>
                    
                    <div class="modal-actions mt-4">
                        <button type="button" class="btn btn-secondary" id="cancel-draw">Cancel</button>
                        <button type="button" class="btn btn-outline" id="simulate-btn">Simulate</button>
                        <button type="submit" class="btn btn-primary" id="execute-btn" disabled>Execute Draw</button>
                    </div>
                </form>
            </div>
        </div>

        <!-- Draw Detail Modal -->
        <div id="draw-detail-modal" class="modal">
            <div class="modal-content modal-lg" id="draw-detail-content">
                Loading...
            </div>
        </div>
    `;
}

export async function initAdminDraws() {
    const formatter = new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' });

    const loadData = async () => {
        try {
            const data = await apiRequest('/api/admin/draws?page=1&size=50');
            renderTable(data.draws || []);
        } catch (err) {
            document.getElementById('draws-table-body').innerHTML = `
                <tr><td colspan="7" class="text-center text-danger">Error: ${err.message}</td></tr>
            `;
        }
    };

    const renderTable = (draws) => {
        const tbody = document.getElementById('draws-table-body');
        if (draws.length === 0) {
            tbody.innerHTML = '<tr><td colspan="7" class="text-center">No draws found.</td></tr>';
            return;
        }

        tbody.innerHTML = draws.map(d => `
            <tr>
                <td><strong>${d.month}/${d.year}</strong></td>
                <td>${new Date(d.draw_date).toLocaleDateString()}</td>
                <td>${formatter.format(d.total_pool)}</td>
                <td class="text-success">${formatter.format(d.winner_prize)}</td>
                <td><span class="badge status-${d.status}">${d.status}</span></td>
                <td>${d.total_entries}</td>
                <td>
                    <button class="btn btn-sm btn-secondary view-detail-btn" data-id="${d.id}">View</button>
                </td>
            </tr>
        `).join('');

        // View detail handlers
        document.querySelectorAll('.view-detail-btn').forEach(btn => {
            btn.addEventListener('click', async (e) => {
                const id = e.currentTarget.dataset.id;
                await openDetailModal(id);
            });
        });
    };

    const openDetailModal = async (id) => {
        const modal = document.getElementById('draw-detail-modal');
        const content = document.getElementById('draw-detail-content');
        modal.classList.add('active');
        content.innerHTML = '<div class="loading-spinner"></div>';

        try {
            const data = await apiRequest(`/api/admin/draws/${id}`);
            
            content.innerHTML = `
                <h2>Draw Results: ${data.draw.month}/${data.draw.year}</h2>
                
                <div class="grid grid-4 mt-4 mb-4">
                    <div class="stat-card p-3">
                        <div class="text-sm text-muted">Total Pool</div>
                        <div class="text-lg">${formatter.format(data.draw.total_pool)}</div>
                    </div>
                    <div class="stat-card p-3">
                        <div class="text-sm text-muted">Winner (60%)</div>
                        <div class="text-lg text-success">${formatter.format(data.draw.winner_prize)}</div>
                    </div>
                    <div class="stat-card p-3">
                        <div class="text-sm text-muted">Charity (30%)</div>
                        <div class="text-lg text-primary">${formatter.format(data.draw.charity_amount)}</div>
                    </div>
                    <div class="stat-card p-3">
                        <div class="text-sm text-muted">Platform (10%)</div>
                        <div class="text-lg">${formatter.format(data.draw.platform_fee)}</div>
                    </div>
                </div>

                <div class="winner-splash mb-4 p-4 card highlight">
                    <h3>🏆 Winner</h3>
                    ${data.winner ? `
                        <div class="mt-2">
                            <strong>${data.winner.user_name}</strong> (${data.winner.user_email})<br>
                            Verification: <span class="badge status-${data.winner.verification_status}">${data.winner.verification_status}</span>
                        </div>
                    ` : '<p>No winner recorded.</p>'}
                </div>

                <h3>All Entries</h3>
                <div class="table-responsive mt-2">
                    <table class="data-table">
                        <thead>
                            <tr>
                                <th>User</th>
                                <th>Entry Score</th>
                                <th>Result</th>
                            </tr>
                        </thead>
                        <tbody>
                            ${data.entries.map(e => `
                                <tr class="${e.is_winner ? 'bg-success-light' : ''}">
                                    <td>${e.user_name} <span class="text-sm text-muted">(${e.user_email})</span></td>
                                    <td>${e.entry_score} pts</td>
                                    <td>${e.is_winner ? '⭐ Winner' : 'Participant'}</td>
                                </tr>
                            `).join('')}
                        </tbody>
                    </table>
                </div>

                <div class="modal-actions mt-4">
                    <button class="btn btn-secondary" onclick="document.getElementById('draw-detail-modal').classList.remove('active')">Close</button>
                </div>
            `;
        } catch (err) {
            content.innerHTML = `
                <div class="error-message">Failed to load details: ${err.message}</div>
                <button class="btn btn-secondary mt-4" onclick="document.getElementById('draw-detail-modal').classList.remove('active')">Close</button>
            `;
        }
    };

    // Run Draw Modal setup
    document.getElementById('new-draw-btn').addEventListener('click', () => {
        document.getElementById('run-draw-form').reset();
        document.getElementById('simulation-panel').style.display = 'none';
        document.getElementById('execute-btn').disabled = true;
        document.getElementById('draw-error').style.display = 'none';
        
        // Default to last month
        let m = new Date().getMonth();
        let y = new Date().getFullYear();
        if (m === 0) { m = 12; y--; }
        document.getElementById('draw-month').value = m;
        document.getElementById('draw-year').value = y;
        
        document.getElementById('run-draw-modal').classList.add('active');
    });

    document.getElementById('cancel-draw').addEventListener('click', () => {
        document.getElementById('run-draw-modal').classList.remove('active');
    });

    // Simulate functionality
    document.getElementById('simulate-btn').addEventListener('click', async () => {
        const errorEl = document.getElementById('draw-error');
        errorEl.style.display = 'none';
        
        const reqBody = {
            month: parseInt(document.getElementById('draw-month').value),
            year: parseInt(document.getElementById('draw-year').value),
            pool_amount: parseFloat(document.getElementById('base-pool').value)
        };

        if (isNaN(reqBody.pool_amount) || reqBody.pool_amount <= 0) {
            errorEl.textContent = 'Please enter a valid pool amount.';
            errorEl.style.display = 'block';
            return;
        }

        try {
            const result = await apiRequest('/api/admin/draws/simulate', {
                method: 'POST',
                body: JSON.stringify(reqBody)
            });
            
            document.getElementById('sim-users').textContent = result.eligible_users;
            document.getElementById('sim-winner').textContent = formatter.format(result.winner_prize);
            document.getElementById('sim-platform').textContent = formatter.format(result.platform_fee);
            
            document.getElementById('simulation-panel').style.display = 'block';
            
            if (result.eligible_users > 0) {
                document.getElementById('execute-btn').disabled = false;
            } else {
                errorEl.textContent = 'No eligible users found for drawing.';
                errorEl.style.display = 'block';
            }
        } catch (err) {
            errorEl.textContent = err.message;
            errorEl.style.display = 'block';
        }
    });

    // Execute functionality
    document.getElementById('run-draw-form').addEventListener('submit', async (e) => {
        e.preventDefault();
        
        if (!confirm('Are you sure you want to run this draw? This will randomly select a winner, assign prizes, and record the results into the database. THIS CANNOT BE UNDONE.')) return;

        const errorEl = document.getElementById('draw-error');
        const btn = document.getElementById('execute-btn');
        btn.disabled = true;
        btn.textContent = 'Executing crypto-draw...';
        
        try {
            const reqBody = {
                month: parseInt(document.getElementById('draw-month').value),
                year: parseInt(document.getElementById('draw-year').value),
                pool_amount: parseFloat(document.getElementById('base-pool').value)
            };

            await apiRequest('/api/admin/draws/run', {
                method: 'POST',
                body: JSON.stringify(reqBody)
            });
            
            document.getElementById('run-draw-modal').classList.remove('active');
            await loadData();
        } catch (err) {
            errorEl.textContent = err.message;
            errorEl.style.display = 'block';
            btn.disabled = false;
            btn.textContent = 'Execute Draw';
        }
    });

    await loadData();
}
