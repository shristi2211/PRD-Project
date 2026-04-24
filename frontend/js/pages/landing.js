import { isAuthenticated, getCurrentUser } from '../auth.js';
import { apiRequest, setCachedUser, submitIpSubscription } from '../utils/api.js';

export function renderLandingPage() {
    return `
        <div class="landing-page v2">
            <!-- Background Elements -->
            <div class="hero-bg-elements">
                <div class="bg-orb orb-1"></div>
                <div class="bg-orb orb-2"></div>
                <div class="bg-orb orb-3"></div>
                <div class="bg-orb orb-4"></div>
                <div class="bg-orb orb-5"></div>
            </div>

            <!-- Central Glass Page -->
            <div class="glass-page-container">
                <div class="hero-glass-pane">
                    <!-- Browser-like Top Bar -->
                    <div class="glass-header">
                        <div class="window-controls">
                            <span class="control red"></span>
                            <span class="control yellow"></span>
                            <span class="control green"></span>
                        </div>
                        <div class="window-address">golf-score-lottery.app</div>
                        <div class="window-search">🔍</div>
                    </div>

                    <!-- Inner Nav -->
                    <nav class="inner-nav">
                        <div class="inner-logo">
                            <span class="logo-circ">YL</span>
                        </div>
                        <div class="inner-links">
                            <a href="#mission">Mission</a>
                            <a href="#roles">Experience</a>
                            <a href="#pricing">Pricing</a>
                        </div>
                        <div class="inner-auth">
                            ${isAuthenticated() ? '<a href="#/dashboard" class="btn btn-primary btn-sm">Go to App</a>' : '<a href="#/login">Sign In</a><a href="#/register" class="btn-signup">Sign up</a>'}
                        </div>
                    </nav>

                    <!-- Hero Content Inside Glass -->
                    <div class="glass-hero-content">
                        <div class="hero-badge">Elevate Your Game</div>
                        <h1 class="hero-title">Play. Score. <br><span class="hero-text-accent">Win Big.</span></h1>
                        <p class="hero-description">
                            The ultimate high-stakes scoring lottery for golfers. Log your rounds, contribute to charities, and enter the monthly draw for exclusive rewards.
                        </p>
                        <div class="hero-btns">
                            <a href="#pricing" class="btn-glass-primary">Get Started</a>
                        </div>
                    </div>

                    <!-- Footer Mockup in Glass -->
                    <div class="glass-footer-mock">
                        <div class="mock-tag">4K PSD MOCKUP</div>
                    </div>
                </div>
            </div>

            <!-- Content Sections -->
            <div class="content-sections">
                <section id="mission" class="mission-section landing-section">
                    <div class="section-inner">
                        <div class="section-header">
                            <h2>Our Mission</h2>
                            <p>Connecting passion with purpose through every golf swing.</p>
                        </div>
                        <div class="mission-grid">
                            <div class="glass-card mission-item">
                                <div class="icon-circle">🎯</div>
                                <h3>The Goal</h3>
                                <p>To create the world's most transparent, charity-driven golf scoring lottery platform where skill meets luck.</p>
                            </div>
                            <div class="glass-card mission-item">
                                <div class="icon-circle">🌍</div>
                                <h3>The Purpose</h3>
                                <p>Democratizing professional-grade rewards for everyday players while supporting global causes that matter.</p>
                            </div>
                            <div class="glass-card mission-item">
                                <div class="icon-circle">⭐</div>
                                <h3>Achievements</h3>
                                <p>Over 10,000+ scores verified, $50,000+ raised for charities, and a thriving community of competitive golfers.</p>
                            </div>
                        </div>
                    </div>
                </section>

                <section id="roles" class="roles-section landing-section">
                    <div class="section-bg-elements">
                        <div class="bg-orb orb-roles-left"></div>
                        <div class="bg-orb orb-roles-center"></div>
                        <div class="bg-orb orb-roles-right"></div>
                    </div>
                    <div class="section-inner">
                        <div class="section-header">
                            <h2>The Ecosystem</h2>
                            <p>Roles designed for every type of member.</p>
                        </div>
                        <div class="roles-grid">
                            <div class="role-card glass-card">
                                <div class="role-icon">👀</div>
                                <h3>Public Visitor</h3>
                                <p>Start your journey by exploring our concept, discovering featured charities, and choosing your subscription path.</p>
                            </div>
                            <div class="role-card glass-card highlighted">
                                <div class="role-icon">🏌️</div>
                                <h3>Registered Subscriber</h3>
                                <p>Our core members who drive the platform. Enter scores, build your charity portfolio, and enter the lottery.</p>
                            </div>
                            <div class="role-card glass-card">
                                <div class="role-icon">🛡️</div>
                                <h3>Administrator</h3>
                                <p>Platform guardians who ensure transparency, manage global draws, and verify winner integrity.</p>
                            </div>
                        </div>
                    </div>
                </section>

                <section id="pricing" class="pricing-section landing-section">
                    <div class="section-inner">
                        <div class="section-header">
                            <h2>Choose Your Plan</h2>
                            <p>Unlock the full potential of your scorecard.</p>
                        </div>
                        <div class="pricing-grid">
                            <div class="glass-card pricing-card">
                                <div class="pricing-header">
                                    <h3>Public Visitor</h3>
                                    <div class="price">Free</div>
                                </div>
                                <ul class="plan-features">
                                    <li>✓ Browse landing page</li>
                                    <li>✓ View featured charities</li>
                                    <li>✗ Score logging</li>
                                    <li>✗ Lottery entries</li>
                                </ul>
                                <button class="btn btn-outline btn-block" id="plan-free-btn">Explore Free</button>
                            </div>
                            <div class="glass-card pricing-card featured">
                                <div class="popular-badge">Most Popular</div>
                                <div class="pricing-header">
                                    <h3>Monthly Pro</h3>
                                    <div class="price">$9.99<span>/mo</span></div>
                                </div>
                                <ul class="plan-features">
                                    <li>✓ Full score logging</li>
                                    <li>✓ Monthly lottery entries</li>
                                    <li>✓ Charity portfolio</li>
                                    <li>✓ Leaderboard access</li>
                                </ul>
                                <button class="btn btn-primary btn-block" id="plan-monthly-btn">Get Monthly</button>
                            </div>
                            <div class="glass-card pricing-card">
                                <div class="pricing-header">
                                    <h3>Yearly Elite</h3>
                                    <div class="price">$99.99<span>/yr</span></div>
                                </div>
                                <ul class="plan-features">
                                    <li>✓ Everything in Monthly</li>
                                    <li>✓ 2 months free</li>
                                    <li>✓ Priority support</li>
                                    <li>✓ Exclusive draws</li>
                                </ul>
                                <button class="btn btn-primary btn-block" id="plan-yearly-btn">Get Yearly</button>
                            </div>
                        </div>
                    </div>
                </section>
            </div>
            
            <footer class="landing-footer">
                <div class="footer-container">
                    <div class="footer-brand">
                        <span class="logo-text">Golf Lottery</span>
                        <p>© 2026 Premium Experiences. All rights reserved.</p>
                    </div>
                    <div class="footer-links">
                        <a href="#">Privacy</a>
                        <a href="#">Terms</a>
                        <a href="#">Support</a>
                    </div>
                </div>
            </footer>

            <!-- Panel Preview Modal -->
            <div class="modal-overlay" id="panel-preview-modal" style="display:none;">
                <div class="modal-glass">
                    <button class="modal-close" id="modal-close-btn">✕</button>
                    <h2 class="modal-title">Explore the Platform</h2>
                    <p class="modal-subtitle">Choose a panel to preview</p>
                    <div class="preview-options">
                        <div class="preview-card" id="preview-admin">
                            <div class="preview-icon">🛡️</div>
                            <h3>Admin Panel</h3>
                            <p>Manage users, execute draws, verify winners, and track subscriptions.</p>
                            <div class="preview-features">
                                <div class="pf-item"><span>📊</span> Dashboard Analytics</div>
                                <div class="pf-item"><span>👥</span> User Management</div>
                                <div class="pf-item"><span>🎰</span> Draw Execution</div>
                                <div class="pf-item"><span>💳</span> Subscription Tracking</div>
                                <div class="pf-item"><span>📝</span> Activity Logs</div>
                            </div>
                        </div>
                        <div class="preview-card" id="preview-user">
                            <div class="preview-icon">🏌️</div>
                            <h3>User Panel</h3>
                            <p>Log scores, select charities, track performance, and win prizes.</p>
                            <div class="preview-features">
                                <div class="pf-item"><span>⛳</span> Score Logging</div>
                                <div class="pf-item"><span>🌍</span> Charity Selection</div>
                                <div class="pf-item"><span>📈</span> Statistics Tracker</div>
                                <div class="pf-item"><span>🏆</span> Lottery Entries</div>
                                <div class="pf-item"><span>💰</span> Subscription Mgmt</div>
                            </div>
                        </div>
                    </div>
                    <div class="modal-footer-actions">
                        <a href="#/register" class="btn btn-primary">Continue to Register →</a>
                    </div>
                </div>
            </div>

            <!-- Payment Confirmation Modal -->
            <div class="modal-overlay" id="payment-modal" style="display:none;">
                <div class="modal-glass modal-sm">
                    <button class="modal-close" id="payment-close-btn">✕</button>
                    <div class="payment-icon">💳</div>
                    <h2 class="modal-title" id="payment-plan-title">Confirm Plan</h2>
                    <p class="modal-subtitle" id="payment-plan-desc">You are about to subscribe.</p>
                    <div class="payment-summary">
                        <div class="payment-row"><span>Plan</span><strong id="payment-plan-name">—</strong></div>
                        <div class="payment-row"><span>Amount</span><strong id="payment-plan-amount">—</strong></div>
                        <div class="payment-row"><span>Billing</span><strong id="payment-plan-billing">—</strong></div>
                    </div>
                    <p class="payment-note">⚡ Mock sandbox — no real charges</p>
                    <button class="btn btn-primary btn-block" id="payment-confirm-btn">Confirm & Subscribe</button>
                    <button class="btn btn-outline btn-block" id="payment-cancel-btn" style="margin-top:8px;">Cancel</button>
                </div>
            </div>
        </div>
    `;
}

export function initLandingPage() {
    // ─── Smooth scrolling ─────────────────────────────────
    document.querySelectorAll('a[href^="#"]').forEach(anchor => {
        anchor.addEventListener('click', function (e) {
            const href = this.getAttribute('href');
            if (href.startsWith('#/')) return;
            e.preventDefault();
            const target = document.querySelector(href);
            if (target) {
                target.scrollIntoView({ behavior: 'smooth' });
            }
        });
    });

    // ─── Intersection Observer for fade-in ────────────────
    const observer = new IntersectionObserver((entries) => {
        entries.forEach(entry => {
            if (entry.isIntersecting) {
                entry.target.classList.add('fade-in-u');
            }
        });
    }, { threshold: 0.1 });
    document.querySelectorAll('.glass-card').forEach(el => observer.observe(el));

    // ─── Panel Preview Modal (Explore Free) ───────────────
    const previewModal = document.getElementById('panel-preview-modal');
    const freeBtn = document.getElementById('plan-free-btn');
    const modalCloseBtn = document.getElementById('modal-close-btn');

    if (freeBtn) {
        freeBtn.addEventListener('click', () => {
            previewModal.style.display = 'flex';
        });
    }
    if (modalCloseBtn) {
        modalCloseBtn.addEventListener('click', () => {
            previewModal.style.display = 'none';
        });
    }
    if (previewModal) {
        previewModal.addEventListener('click', (e) => {
            if (e.target === previewModal) previewModal.style.display = 'none';
        });
    }

    // ─── Payment Confirmation Modal ───────────────────────
    const paymentModal = document.getElementById('payment-modal');
    const paymentCloseBtn = document.getElementById('payment-close-btn');
    const paymentCancelBtn = document.getElementById('payment-cancel-btn');
    const paymentConfirmBtn = document.getElementById('payment-confirm-btn');

    let selectedPlan = null;

    function openPaymentModal(plan) {
        selectedPlan = plan;
        const isMonthly = plan === 'monthly';
        document.getElementById('payment-plan-title').textContent = isMonthly ? 'Monthly Pro Plan' : 'Yearly Elite Plan';
        document.getElementById('payment-plan-desc').textContent = isMonthly
            ? 'Full access to all platform features, billed monthly.'
            : 'Best value — save 2 months with yearly billing.';
        document.getElementById('payment-plan-name').textContent = isMonthly ? 'Monthly Pro' : 'Yearly Elite';
        document.getElementById('payment-plan-amount').textContent = isMonthly ? '$9.99' : '$99.99';
        document.getElementById('payment-plan-billing').textContent = isMonthly ? 'Billed Monthly' : 'Billed Annually';
        paymentModal.style.display = 'flex';
    }

    const monthlyBtn = document.getElementById('plan-monthly-btn');
    const yearlyBtn = document.getElementById('plan-yearly-btn');

    if (monthlyBtn) monthlyBtn.addEventListener('click', () => openPaymentModal('monthly'));
    if (yearlyBtn) yearlyBtn.addEventListener('click', () => openPaymentModal('yearly'));

    if (paymentCloseBtn) paymentCloseBtn.addEventListener('click', () => { paymentModal.style.display = 'none'; });
    if (paymentCancelBtn) paymentCancelBtn.addEventListener('click', () => { paymentModal.style.display = 'none'; });
    if (paymentModal) {
        paymentModal.addEventListener('click', (e) => {
            if (e.target === paymentModal) paymentModal.style.display = 'none';
        });
    }

    if (paymentConfirmBtn) {
        paymentConfirmBtn.addEventListener('click', async () => {
            const ogText = paymentConfirmBtn.textContent;
            paymentConfirmBtn.textContent = 'Processing...';
            paymentConfirmBtn.disabled = true;
            
            try {
                if (isAuthenticated()) {
                    // User is already logged in, upgrade their actual account
                    const data = await apiRequest('/api/subscriptions/start', {
                        method: 'POST',
                        body: JSON.stringify({ plan: selectedPlan }),
                    });
                    
                    let user = await getCurrentUser();
                    if (user) {
                        user = { ...user, ...data };
                        setCachedUser(user);
                    }
                    
                    paymentModal.style.display = 'none';
                    alert(`Successfully upgraded to ${selectedPlan} plan!`);
                    window.location.hash = '#/dashboard';
                } else {
                    // User is NOT logged in: Lock the IP with a subscription
                    await submitIpSubscription(selectedPlan);
                    localStorage.setItem('gsl_pending_plan', selectedPlan);
                    paymentModal.style.display = 'none';
                    window.location.hash = '#/register';
                }
            } catch (err) {
                alert('Transaction Failed: ' + err.message);
            } finally {
                paymentConfirmBtn.textContent = ogText;
                paymentConfirmBtn.disabled = false;
            }
        });
    }
}
