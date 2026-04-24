import { apiRequest } from '../utils/api.js';

export function renderScores() {
    return `
        <div class="scores-container list-card">
            <header class="page-header">
                <div class="header-content">
                    <h1>My Scores</h1>
                    <p class="subtitle">Log your Stableford scores to enter the monthly draw.</p>
                </div>
            </header>

            <div class="dashboard-grid">
                <!-- Submit Form -->
                <div class="dashboard-card">
                    <div style="display:flex; justify-content:space-between; align-items:center; margin-bottom:var(--space-md);">
                        <h2 id="form-title">Submit New Score</h2>
                        <button type="button" id="cancel-edit-btn" class="btn btn-ghost btn-sm" style="display:none;">Cancel Edit</button>
                    </div>
                    <form id="score-form" class="standard-form">
                        <div id="score-error" class="error-message" style="display: none;"></div>
                        
                        <div class="form-group">
                            <label for="score">Stableford Points</label>
                            <input type="number" id="score" min="1" max="45" required placeholder="Enter points (1-45)">
                        </div>
                        
                        <div class="form-group">
                            <label for="round-date">Round Date</label>
                            <input type="date" id="round-date" required>
                        </div>
                        
                        <div class="form-group">
                            <label for="notes">Notes/Course (Optional)</label>
                            <input type="text" id="notes" placeholder="e.g. Pebble Beach">
                        </div>
                        
                        <button type="submit" class="btn btn-primary btn-block mt-4" id="submit-score-btn">
                            Submit Score
                        </button>
                        <p class="text-xs text-muted" style="margin-top:var(--space-md);text-align:center;">
                            You can store a maximum of 5 scores. New submissions will automatically roll over and replace your oldest score. Only 1 score is allowed per date.
                        </p>
                    </form>
                </div>

                <!-- History -->
                <div class="dashboard-card">
                    <div class="flex-between align-center mb-4">
                        <h2>Score History</h2>
                        <span class="text-sm text-muted"><span id="score-count">0</span> / 5 Scores</span>
                    </div>
                    <div class="progress-bar mb-4">
                        <div class="progress-fill" id="score-progress" style="width: 0%"></div>
                    </div>
                    
                    <div id="scores-list" class="drill-list">
                        <div class="loading-spinner"></div>
                    </div>
                </div>
            </div>
        </div>

        <!-- Delete Modal -->
        <div id="delete-score-modal" class="modal">
            <div class="modal-content">
                <h2>Delete Score</h2>
                <p>Are you sure you want to delete this score? This action cannot be undone.</p>
                <div id="delete-error" class="error-message" style="display: none;"></div>
                <div class="modal-actions" style="flex-direction:row;gap:0.75rem;margin-top:var(--space-lg)">
                    <button class="btn btn-secondary btn-sm" id="cancel-delete">Cancel</button>
                    <button class="btn btn-danger btn-sm" id="confirm-delete">Delete</button>
                </div>
            </div>
        </div>
    `;
}

export async function initScores() {
    document.getElementById('round-date').valueAsDate = new Date();
    
    let currentScores = [];
    let scoreToDelete = null;
    let editingScoreId = null;

    const loadScores = async () => {
        try {
            const data = await apiRequest('/api/scores');
            currentScores = data || [];
            updateUI();
        } catch (err) {
            document.getElementById('scores-list').innerHTML = `
                <div class="empty-state">
                    <p class="error-text">Failed to load scores: ${err.message}</p>
                </div>
            `;
        }
    };

    const updateUI = () => {
        const count = currentScores.length;
        document.getElementById('score-count').textContent = count;
        document.getElementById('score-progress').style.width = `${(count / 5) * 100}%`;

        const listEl = document.getElementById('scores-list');
        if (count === 0) {
            listEl.innerHTML = `
                <div class="empty-state">
                    <div class="empty-icon">⛳</div>
                    <h3>No scores yet</h3>
                    <p>Submit your first score to enter the lottery.</p>
                </div>
            `;
            return;
        }

        const bestScore = Math.max(...currentScores.map(s => s.score));

        listEl.innerHTML = currentScores.map(score => `
            <div class="score-item ${score.score === bestScore ? 'best-score' : ''}" style="display:flex; justify-content:space-between; align-items:center; padding:var(--space-md); border-bottom:1px solid var(--color-border-glass);">
                <div class="score-details">
                    <div style="font-size:var(--font-size-2xl); font-weight:800; color:${score.score === bestScore ? 'var(--color-gold)' : 'var(--color-accent)'}">
                        ${score.score}<span style="font-size:var(--font-size-xs); color:var(--color-text-muted); margin-left:2px">pts</span>
                    </div>
                    <div class="score-info" style="margin-top:2px;">
                        <div class="score-date text-sm">${score.round_date}</div>
                        ${score.notes ? `<div class="score-notes text-xs text-muted">${score.notes}</div>` : ''}
                    </div>
                </div>
                <div class="score-actions" style="display:flex; gap:0.5rem; flex-direction:column; align-items:flex-end;">
                    ${score.score === bestScore ? '<span class="badge badge-success mb-2">Best Entry</span>' : ''}
                    <div style="display:flex; gap:0.5rem;">
                        <button class="btn btn-icon edit-btn" style="padding:4px" data-id="${score.id}" aria-label="Edit">✏️</button>
                        <button class="btn btn-icon delete-btn" style="padding:4px" data-id="${score.id}" aria-label="Delete">🗑️</button>
                    </div>
                </div>
            </div>
        `).join('');

        // Edit handlers
        document.querySelectorAll('.edit-btn').forEach(btn => {
            btn.addEventListener('click', (e) => {
                const id = e.currentTarget.dataset.id;
                const score = currentScores.find(s => s.id === id);
                if (score) {
                    editingScoreId = id;
                    document.getElementById('score').value = score.score;
                    document.getElementById('round-date').value = score.round_date;
                    document.getElementById('notes').value = score.notes || '';
                    
                    document.getElementById('form-title').textContent = 'Edit Score';
                    document.getElementById('submit-score-btn').textContent = 'Update Score';
                    document.getElementById('cancel-edit-btn').style.display = 'block';
                    
                    document.getElementById('score').focus();
                }
            });
        });

        // Delete handlers
        document.querySelectorAll('.delete-btn').forEach(btn => {
            btn.addEventListener('click', (e) => {
                scoreToDelete = e.currentTarget.dataset.id;
                document.getElementById('delete-score-modal').classList.add('active');
            });
        });
    };

    document.getElementById('cancel-edit-btn').addEventListener('click', () => {
        editingScoreId = null;
        document.getElementById('score-form').reset();
        document.getElementById('round-date').valueAsDate = new Date();
        document.getElementById('form-title').textContent = 'Submit New Score';
        document.getElementById('submit-score-btn').textContent = 'Submit Score';
        document.getElementById('cancel-edit-btn').style.display = 'none';
        document.getElementById('score-error').style.display = 'none';
    });

    document.getElementById('score-form').addEventListener('submit', async (e) => {
        e.preventDefault();
        const errorEl = document.getElementById('score-error');
        const btn = document.getElementById('submit-score-btn');
        errorEl.style.display = 'none';
        
        const scoreVal = parseInt(document.getElementById('score').value, 10);
        const dateVal = document.getElementById('round-date').value;
        const notesVal = document.getElementById('notes').value;

        btn.disabled = true;
        btn.textContent = editingScoreId ? 'Updating...' : 'Submitting...';

        try {
            if (editingScoreId) {
                await apiRequest(`/api/scores/${editingScoreId}`, {
                    method: 'PUT',
                    body: JSON.stringify({ score: scoreVal, round_date: dateVal, notes: notesVal })
                });
            } else {
                await apiRequest('/api/scores', {
                    method: 'POST',
                    body: JSON.stringify({ score: scoreVal, round_date: dateVal, notes: notesVal })
                });
            }
            
            document.getElementById('cancel-edit-btn').click();
            await loadScores();
        } catch (err) {
            errorEl.textContent = err.message;
            errorEl.style.display = 'block';
        } finally {
            btn.disabled = false;
        }
    });

    const deleteModal = document.getElementById('delete-score-modal');
    
    document.getElementById('cancel-delete').addEventListener('click', () => {
        deleteModal.classList.remove('active');
        scoreToDelete = null;
    });

    document.getElementById('confirm-delete').addEventListener('click', async () => {
        if (!scoreToDelete) return;
        
        const errorEl = document.getElementById('delete-error');
        const confirmBtn = document.getElementById('confirm-delete');
        errorEl.style.display = 'none';
        
        confirmBtn.disabled = true;
        confirmBtn.textContent = 'Deleting...';

        try {
            await apiRequest(`/api/scores/${scoreToDelete}`, { method: 'DELETE' });
            deleteModal.classList.remove('active');
            await loadScores();
        } catch (err) {
            errorEl.textContent = err.message;
            errorEl.style.display = 'block';
        } finally {
            confirmBtn.disabled = false;
            confirmBtn.textContent = 'Delete';
            scoreToDelete = null;
        }
    });

    await loadScores();
}
