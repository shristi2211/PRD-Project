import { apiRequest } from '../utils/api.js';

export function renderCharity() {
    return `
        <div class="charity-container list-card">
            <header class="page-header">
                <div class="header-content">
                    <h1>My Charity Portfolio</h1>
                    <p class="subtitle">Distribute your 30% platform cut across the charities below.</p>
                </div>
            </header>

            <div class="dashboard-card" style="margin-bottom:var(--space-xl); position:sticky; top:20px; z-index:10; border:1px solid var(--color-accent-light)">
                <div style="display:flex; justify-content:space-between; align-items:center;">
                    <div>
                        <h2 style="margin-bottom:4px; font-size:var(--font-size-lg)">Total Allocated</h2>
                        <p class="text-sm text-muted">You must allocate exactly 30% before saving.</p>
                    </div>
                    <div style="text-align:right">
                        <div style="font-size:var(--font-size-3xl); font-weight:800; color:var(--color-accent-light)"><span id="total-allocated">0</span> / 30%</div>
                    </div>
                </div>
                <div style="margin-top:var(--space-md); text-align:right;">
                    <button class="btn btn-primary" id="save-allocations-btn" disabled>Save Portfolio</button>
                    <span id="save-status" style="margin-left:var(--space-md); font-size:var(--font-size-sm);"></span>
                </div>
            </div>

            <div class="dashboard-grid" id="charities-grid">
                <div class="loading-spinner"></div>
            </div>
            
        </div>
    `;
}

export async function initCharity() {
    let charities = [];
    let allocations = {}; // Map of charityID -> percentage

    const loadData = async () => {
        try {
            const [charitiesData, selectionData] = await Promise.all([
                apiRequest('/api/charities'),
                apiRequest('/api/charity/my-selection')
            ]);
            
            charities = charitiesData || [];
            
            // Map existing selections to dict
            allocations = {};
            if (selectionData && Array.isArray(selectionData)) {
                selectionData.forEach(sel => {
                    allocations[sel.charity_id] = sel.contribution_percentage;
                });
            }
            
            renderGrid();
            updateTotal();
        } catch (err) {
            document.getElementById('charities-grid').innerHTML = `
                <div class="error-state">Failed to load data: ${err.message}</div>
            `;
        }
    };

    const updateTotal = () => {
        let sum = 0;
        Object.values(allocations).forEach(val => sum += parseInt(val || 0));
        
        const totalEl = document.getElementById('total-allocated');
        const saveBtn = document.getElementById('save-allocations-btn');
        
        totalEl.textContent = sum;
        
        if (sum === 30) {
            totalEl.style.color = 'var(--color-success)';
            saveBtn.disabled = false;
        } else if (sum > 30) {
            totalEl.style.color = 'var(--color-error)';
            saveBtn.disabled = true;
        } else {
            totalEl.style.color = 'var(--color-accent-light)';
            saveBtn.disabled = true;
        }
    };

    const handleSliderInput = (charityId, valStr) => {
        let val = parseInt(valStr || 0);
        
        // Safety cap check (prevent jumping way over 30 via manual typing)
        let otherSum = 0;
        Object.keys(allocations).forEach(k => {
            if (k !== charityId) otherSum += parseInt(allocations[k] || 0);
        });
        
        if (otherSum + val > 30) {
            val = 30 - otherSum;
        }
        
        if (val < 0) val = 0;
        
        allocations[charityId] = val;
        
        // Sync UI inputs
        document.getElementById('slider-' + charityId).value = val;
        document.getElementById('badge-' + charityId).textContent = val;
        
        updateTotal();
    };

    const renderGrid = () => {
        const grid = document.getElementById('charities-grid');
        if (charities.length === 0) {
            grid.innerHTML = '<div class="empty-state">No charities available at the moment.</div>';
            return;
        }

        grid.innerHTML = charities.map(c => {
            const currentVal = allocations[c.id] || 0;
            return `
            <div class="dashboard-card" style="display:flex; flex-direction:column; justify-content:space-between;">
                <div>
                    <div class="charity-header" style="display:flex; align-items:center; gap:var(--space-md); margin-bottom:var(--space-md)">
                        ${c.logo_url ? `<img src="${c.logo_url}" alt="${c.name}" class="charity-logo" style="width:40px;height:40px;border-radius:8px">` : `<div style="width:40px;height:40px;border-radius:8px;background:var(--color-bg-glass);display:flex;align-items:center;justify-content:center;font-weight:bold">${c.name.charAt(0)}</div>`}
                        <h4 style="margin:0">${c.name}</h4>
                    </div>
                    <p class="text-sm text-muted mb-4">${c.description}</p>
                </div>
                
                <div style="margin-top:auto; padding-top:var(--space-md); border-top:1px solid var(--color-border-glass)">
                    <div style="display:flex; justify-content:space-between; align-items:center; margin-bottom:var(--space-sm)">
                        <span class="text-xs text-muted">Allocation</span>
                        <span style="font-weight:bold; color:var(--color-accent-light)"><span id="badge-${c.id}">${currentVal}</span>%</span>
                    </div>
                    <input type="range" id="slider-${c.id}" data-id="${c.id}" min="0" max="30" step="1" value="${currentVal}" style="width:100%" class="allocation-slider">
                </div>
            </div>
            `;
        }).join('');

        // Listeners for ranges
        document.querySelectorAll('.allocation-slider').forEach(slider => {
            slider.addEventListener('input', (e) => {
                handleSliderInput(e.target.dataset.id, e.target.value);
            });
        });
    };

    document.getElementById('save-allocations-btn').addEventListener('click', async (e) => {
        const btn = e.target;
        const statusEl = document.getElementById('save-status');
        
        const payloadAllocations = Object.keys(allocations)
            .filter(id => allocations[id] > 0)
            .map(id => ({ charity_id: id, contribution_percentage: allocations[id] }));
            
        btn.disabled = true;
        btn.textContent = "Saving...";
        statusEl.textContent = "";
        
        try {
            await apiRequest('/api/charity/select', {
                method: 'POST',
                body: JSON.stringify({ allocations: payloadAllocations })
            });
            statusEl.style.color = "var(--color-success)";
            statusEl.textContent = "Portfolio saved successfully!";
        } catch (err) {
            statusEl.style.color = "var(--color-error)";
            statusEl.textContent = "Failed to save: " + err.message;
        } finally {
            btn.disabled = false;
            btn.textContent = "Save Portfolio";
            setTimeout(() => statusEl.textContent = "", 4000); // clear msg
        }
    });

    loadData();
}
