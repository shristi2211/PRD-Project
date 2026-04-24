import { apiRequest } from '../utils/api.js';

export function renderMyStats() {
    return `
        <div class="stats-container">
            <header class="page-header">
                <div class="header-content">
                    <h1>My Statistics</h1>
                    <p class="subtitle">Track your performance and winnings.</p>
                </div>
            </header>

            <div class="stats-overview-grid">
                <div class="stat-card">
                    <div class="stat-icon">⛳</div>
                    <div class="stat-content">
                        <h3>Rounds Played</h3>
                        <div class="stat-value" id="stat-rounds">-</div>
                    </div>
                </div>
                <div class="stat-card highlight">
                    <div class="stat-icon">🏆</div>
                    <div class="stat-content">
                        <h3>Best Score</h3>
                        <div class="stat-value"><span id="stat-best">-</span><span class="stat-unit">pts</span></div>
                    </div>
                </div>
                <div class="stat-card">
                    <div class="stat-icon">🎟️</div>
                    <div class="stat-content">
                        <h3>Draw Entries</h3>
                        <div class="stat-value" id="stat-entries">-</div>
                    </div>
                </div>
                <div class="stat-card success">
                    <div class="stat-icon">💰</div>
                    <div class="stat-content">
                        <h3>Total Winnings</h3>
                        <div class="stat-value" id="stat-winnings">-</div>
                    </div>
                </div>
            </div>

            <div class="dashboard-grid mt-4">
                <div style="display:flex;flex-direction:column;gap:var(--space-lg)">
                    <div class="dashboard-card chart-card">
                        <h2>Score History Trend</h2>
                        <div class="chart-container">
                            <canvas id="scoreTrendChart"></canvas>
                        </div>
                    </div>
                    <div class="dashboard-card chart-card">
                        <h2>Lottery Win Rate</h2>
                        <div class="chart-container" style="height:240px">
                            <canvas id="winRateChart"></canvas>
                        </div>
                    </div>
                </div>

                <div class="dashboard-card list-card">
                    <h2>My Winnings</h2>
                    <div id="winnings-list" class="winnings-list">
                        <div class="loading-spinner"></div>
                    </div>
                </div>
            </div>
        </div>

        <!-- Submit Proof Modal -->
        <div id="submit-proof-modal" class="modal">
            <div class="modal-content">
                <h2>Submit Score Proof</h2>
                <p>Congratulations! To claim your prize, please provide a link to your verified scorecard.</p>
                
                <form id="proof-form" class="standard-form">
                    <div id="proof-error" class="error-message" style="display: none;"></div>
                    
                    <div class="form-group">
                        <label for="proof-url">Proof URL</label>
                        <input type="url" id="proof-url" required placeholder="https://link-to-scorecard.com">
                    </div>
                    
                    <div class="form-group">
                        <label for="proof-notes">Additional Notes</label>
                        <textarea id="proof-notes" rows="3" placeholder="Any context helping admin verify..."></textarea>
                    </div>
                    
                    <div class="modal-actions" style="flex-direction:row;gap:0.75rem">
                        <button type="button" class="btn btn-secondary btn-sm" id="cancel-proof">Cancel</button>
                        <button type="submit" class="btn btn-primary btn-sm" id="confirm-proof">Submit Proof</button>
                    </div>
                </form>
            </div>
        </div>
    `;
}

export async function initMyStats() {
    let scoreChart = null;
    let winChart = null;
    let selectedWinnerId = null;

    const loadData = async () => {
        try {
            const [stats, trendData, winnings] = await Promise.all([
                apiRequest('/api/stats/dashboard'),
                apiRequest('/api/stats/score-trend'),
                apiRequest('/api/winners/me')
            ]);
            
            updateOverviewStats(stats);
            renderCharts(stats, trendData);
            updateWinningsList(winnings || []);
        } catch (err) {
            console.error('Failed to load stats:', err);
        }
    };

    const updateOverviewStats = (stats) => {
        if (!stats) return;
        document.getElementById('stat-rounds').textContent = stats.rounds_played || 0;
        document.getElementById('stat-best').textContent = stats.best_score || '-';
        document.getElementById('stat-entries').textContent = stats.lottery_entries || 0;
        
        const formatter = new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' });
        document.getElementById('stat-winnings').textContent = formatter.format(stats.total_winnings || 0);
    };

    const renderCharts = (stats, trendData) => {
        if (!trendData || trendData.length === 0) return;
        
        // --- Line Chart ---
        const ctxTrend = document.getElementById('scoreTrendChart');
        if (ctxTrend) {
            if (scoreChart) scoreChart.destroy();
            const primaryColor = getComputedStyle(document.documentElement).getPropertyValue('--color-accent').trim() || '#FF4F00';
            
            scoreChart = new Chart(ctxTrend, {
                type: 'line',
                data: {
                    labels: trendData.map(d => d.label),
                    datasets: [{
                        label: 'Score',
                        data: trendData.map(d => d.value),
                        borderColor: primaryColor,
                        backgroundColor: primaryColor + '20',
                        borderWidth: 3,
                        tension: 0.4,
                        fill: true
                    }]
                },
                options: {
                    responsive: true,
                    maintainAspectRatio: false,
                    plugins: { legend: { display: false } },
                    scales: {
                        y: { 
                            beginAtZero: true,
                            max: 45,
                            grid: { color: 'rgba(255,255,255,0.05)' },
                            ticks: { color: '#94a3b8' }
                        },
                        x: {
                            grid: { display: false },
                            ticks: { color: '#94a3b8' }
                        }
                    }
                }
            });
        }

        // --- Donut Chart ---
        const ctxWin = document.getElementById('winRateChart');
        if (ctxWin && stats) {
            if (winChart) winChart.destroy();
            
            // Assume total winnings entries estimate win count (for presentation purposes, actual win count can be fetched from API later)
            // Just displaying participated vs not won for now based on stats
            const entries = stats.lottery_entries || 0;
            const wins = stats.total_winnings > 0 ? 1 : 0; // simplified
            
            const noWins = entries - wins;
            
            winChart = new Chart(ctxWin, {
                type: 'doughnut',
                data: {
                    labels: ['Won', 'Participated'],
                    datasets: [{
                        data: [wins, noWins < 0 ? 0 : noWins],
                        backgroundColor: ['#10b981', '#1e293b'],
                        borderWidth: 0,
                        hoverOffset: 4
                    }]
                },
                options: {
                    responsive: true,
                    maintainAspectRatio: false,
                    cutout: '75%',
                    plugins: {
                        legend: { position: 'bottom', labels: { color: '#e2e8f0'} }
                    }
                }
            });
        }
    };

    const updateWinningsList = (winnings) => {
        const listEl = document.getElementById('winnings-list');
        if (winnings.length === 0) {
            listEl.innerHTML = '<div class="empty-state">No winnings yet. Keep playing!</div>';
            return;
        }

        const formatter = new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' });

        listEl.innerHTML = winnings.map(w => `
            <div class="winning-item">
                <div class="winning-info">
                    <h4>Draw: ${w.draw_month}/${w.draw_year}</h4>
                    <div class="winning-amount">${formatter.format(w.prize_amount)}</div>
                </div>
                <div class="winning-status">
                    <span class="status-badge status-${w.verification_status}">${w.verification_status}</span>
                    ${w.verification_status === 'pending' ? `
                        <button class="btn-outline btn-sm submit-proof-btn mt-2" data-id="${w.id}">
                            Submit Proof
                        </button>
                    ` : ''}
                    ${w.verification_status === 'rejected' ? `
                        <div class="rejection-reason mt-1">${w.rejection_reason}</div>
                    ` : ''}
                </div>
            </div>
        `).join('');

        document.querySelectorAll('.submit-proof-btn').forEach(btn => {
            btn.addEventListener('click', (e) => {
                selectedWinnerId = e.target.dataset.id;
                document.getElementById('submit-proof-modal').classList.add('active');
            });
        });
    };

    document.getElementById('proof-form').addEventListener('submit', async (e) => {
        e.preventDefault();
        const errorEl = document.getElementById('proof-error');
        const btn = document.getElementById('confirm-proof');
        errorEl.style.display = 'none';
        
        btn.disabled = true;
        btn.textContent = 'Submitting...';

        try {
            await apiRequest(`/api/winners/${selectedWinnerId}/proof`, {
                method: 'PUT',
                body: JSON.stringify({
                    proof_url: document.getElementById('proof-url').value,
                    proof_notes: document.getElementById('proof-notes').value
                })
            });
            
            document.getElementById('submit-proof-modal').classList.remove('active');
            document.getElementById('proof-form').reset();
            await loadData();
        } catch (err) {
            errorEl.textContent = err.message;
            errorEl.style.display = 'block';
        } finally {
            btn.disabled = false;
            btn.textContent = 'Submit Proof';
        }
    });

    document.getElementById('cancel-proof').addEventListener('click', () => {
        document.getElementById('submit-proof-modal').classList.remove('active');
        document.getElementById('proof-form').reset();
    });

    await loadData();
}
