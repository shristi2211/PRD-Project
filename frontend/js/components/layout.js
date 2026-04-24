// ============================================
// Layout Component — Sidebar + Content wrapper
// ============================================

import { renderSidebar, initSidebar } from '/js/components/sidebar.js';
import { logout } from '/js/auth.js';

/**
 * Render the full dashboard layout with sidebar + content area.
 * @param {string} contentHTML - The page content to render inside the main area
 * @param {string} pageTitle - Title shown in the topbar
 */
export function renderDashboardLayout(contentHTML, pageTitle = 'Dashboard') {
    return `
        <div class="dashboard-layout">
            ${renderSidebar()}

            <main class="dashboard-main">
                <!-- Top bar (mobile toggle + page title) -->
                <header class="dashboard-topbar">
                    <button class="sidebar-toggle" id="sidebar-toggle" aria-label="Toggle menu">
                        <span class="hamburger-line"></span>
                        <span class="hamburger-line"></span>
                        <span class="hamburger-line"></span>
                    </button>
                    <h2 class="topbar-title">${pageTitle}</h2>
                </header>

                <!-- Page content -->
                <div class="dashboard-content">
                    ${contentHTML}
                </div>
            </main>
        </div>
    `;
}

/**
 * Initialize layout-level interactivity (sidebar + logout).
 * Call this after renderDashboardLayout is in the DOM.
 */
export function initDashboardLayout() {
    initSidebar();

    // Logout handler
    const logoutBtn = document.getElementById('sidebar-logout-btn');
    if (logoutBtn) {
        logoutBtn.addEventListener('click', async () => {
            logoutBtn.disabled = true;
            logoutBtn.textContent = '⏳';
            await logout();
            window.location.hash = '#/login';
        });
    }
}
