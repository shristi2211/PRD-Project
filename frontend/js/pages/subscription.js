import { getCurrentUser } from '/js/auth.js';
import { apiRequest, setCachedUser, clearTokens } from '/js/utils/api.js';

export function renderSubscription() {
    return `
        <div class="subscription-container list-card">
            <header class="page-header">
                <div class="header-content">
                    <h1>My Subscription</h1>
                    <p class="subtitle">Manage your Golf Score Lottery membership.</p>
                </div>
            </header>

            <div class="dashboard-card" style="max-width: 700px; margin: 0 auto;">
                <div id="sub-loading">
                    <div class="loading-spinner"></div>
                </div>
                
                <div id="sub-content" style="display:none;">
                    <div id="status-badge" style="margin-bottom: var(--space-lg); text-align: center;"></div>
                    
                    <div class="sub-plan-info" id="sub-plan-info"></div>

                    <div id="sub-actions" style="margin-top: var(--space-xl); border-top: 1px solid var(--color-border-glass); padding-top: var(--space-xl);"></div>
                </div>
            </div>
            
            <div style="text-align:center; margin-top:var(--space-2xl);">
                <p class="text-xs text-muted">Stripe payment gateway is currently running in mock sandbox mode.</p>
            </div>
        </div>
    `;
}

export async function initSubscription() {
    let user = null;

    const loadUser = async () => {
        try {
            user = await getCurrentUser();
            updateUI();
        } catch (err) {
            console.error(err);
        }
    };

    const updateUI = () => {
        document.getElementById('sub-loading').style.display = 'none';
        document.getElementById('sub-content').style.display = 'block';

        const planType = user.subscription_type || 'free';
        const isActive = user.subscription_active;

        // Status badge
        let badgeHtml = '';
        if (isActive) {
            const planLabel = planType === 'monthly' ? 'Monthly Pro' : 'Yearly Elite';
            badgeHtml = `<span class="badge badge-success" style="font-size: var(--font-size-md); padding: var(--space-xs) var(--space-md)">✓ ${planLabel} — Active</span>`;
        } else {
            badgeHtml = '<span class="badge badge-inactive" style="font-size: var(--font-size-md); padding: var(--space-xs) var(--space-md)">Inactive — Free Plan</span>';
        }
        document.getElementById('status-badge').innerHTML = badgeHtml;

        // Plan info
        const planInfo = document.getElementById('sub-plan-info');
        if (isActive) {
            const price = planType === 'monthly' ? '$9.99/mo' : '$99.99/yr';
            planInfo.innerHTML = `
                <div style="text-align: center;">
                    <div style="font-size: var(--font-size-4xl); font-weight: 800; margin-bottom: var(--space-sm);">
                        ${planType === 'monthly' ? 'Monthly Pro' : 'Yearly Elite'}
                    </div>
                    <div style="font-size: var(--font-size-2xl); color: var(--color-accent); font-weight: 700; margin-bottom: var(--space-md);">
                        ${price}
                    </div>
                    <p class="text-muted">Full access to score logging, lottery entries, charity portfolio, and leaderboard tracking.</p>
                </div>
            `;
        } else {
            planInfo.innerHTML = `
                <div style="text-align: center;">
                    <div style="font-size: var(--font-size-4xl); font-weight: 800; margin-bottom: var(--space-sm); opacity: 0.5;">
                        Free Plan
                    </div>
                    <p class="text-muted">You currently have no active subscription. Upgrade to access all features.</p>
                </div>
            `;
        }

        // Actions
        const actionsContainer = document.getElementById('sub-actions');
        if (isActive) {
            const switchTo = planType === 'monthly' ? 'yearly' : 'monthly';
            const switchLabel = planType === 'monthly' ? 'Switch to Yearly Elite ($99.99/yr)' : 'Switch to Monthly Pro ($9.99/mo)';
            actionsContainer.innerHTML = `
                <div style="display: flex; gap: var(--space-md); justify-content: center; flex-wrap: wrap;">
                    <button class="btn btn-outline" id="switch-plan-btn">${switchLabel}</button>
                    <button class="btn btn-outline" style="border-color: #ef4444; color: #ef4444;" id="cancel-sub-btn">Cancel Subscription</button>
                </div>
            `;

            document.getElementById('switch-plan-btn').addEventListener('click', async () => {
                const btn = document.getElementById('switch-plan-btn');
                btn.disabled = true;
                btn.textContent = 'Switching...';
                try {
                    const data = await apiRequest('/api/subscriptions/start', {
                        method: 'POST',
                        body: JSON.stringify({ plan: switchTo }),
                    });
                    user = { ...user, ...data };
                    setCachedUser(user);
                    updateUI();
                } catch (err) {
                    alert('Failed to switch plan: ' + err.message);
                    btn.disabled = false;
                }
            });

            document.getElementById('cancel-sub-btn').addEventListener('click', async () => {
                if (!confirm('Are you sure you want to cancel your subscription? You will be logged out immediately.')) return;
                const btn = document.getElementById('cancel-sub-btn');
                btn.disabled = true;
                btn.textContent = 'Cancelling...';
                try {
                    await apiRequest('/api/subscriptions/cancel', { method: 'PUT' });
                    clearTokens();
                    window.location.hash = '#/';
                } catch (err) {
                    alert('Failed to cancel: ' + err.message);
                    btn.disabled = false;
                }
            });
        } else {
            actionsContainer.innerHTML = `
                <div style="display: flex; gap: var(--space-md); justify-content: center; flex-wrap: wrap;">
                    <button class="btn btn-primary" id="activate-monthly-btn">Activate Monthly ($9.99/mo)</button>
                    <button class="btn btn-primary" id="activate-yearly-btn">Activate Yearly ($99.99/yr)</button>
                </div>
            `;

            document.getElementById('activate-monthly-btn').addEventListener('click', () => activatePlan('monthly'));
            document.getElementById('activate-yearly-btn').addEventListener('click', () => activatePlan('yearly'));
        }
    };

    const activatePlan = async (plan) => {
        try {
            const data = await apiRequest('/api/subscriptions/start', {
                method: 'POST',
                body: JSON.stringify({ plan }),
            });
            user = { ...user, ...data };
            setCachedUser(user);
            updateUI();
        } catch (err) {
            alert('Failed to activate plan: ' + err.message);
        }
    };

    loadUser();
}
