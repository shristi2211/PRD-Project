// ============================================
// Sidebar Component — Role-based navigation
// ============================================

import { getFilteredNav } from '/js/config/navigation.js';
import { getUser } from '/js/auth.js';

/**
 * Render the sidebar HTML.
 * Filters navigation items based on the current user's role.
 */
export function renderSidebar() {
    const user = getUser();
    const role = user?.role || 'user';
    const sections = getFilteredNav(role);
    const currentHash = window.location.hash.slice(1) || '/dashboard';
    const initial = user?.name ? user.name.charAt(0).toUpperCase() : '?';
    const displayName = user?.name || 'User';
    const displayEmail = user?.email || '';
    const roleLabel = role === 'admin' ? 'Admin Panel' : 'Player Dashboard';

    let navHTML = '';

    sections.forEach(section => {
        // Section header
        if (section.title) {
            navHTML += `<div class="sidebar-section-header">${section.title}</div>`;
        }

        // Menu items
        section.items.forEach(item => {
            const isActive = currentHash === item.path;
            navHTML += `
                <a href="#${item.path}"
                   class="sidebar-nav-item${isActive ? ' active' : ''}"
                   data-path="${item.path}"
                   id="nav-${item.path.replace(/\//g, '-').slice(1)}">
                    <span class="sidebar-nav-icon">${item.icon}</span>
                    <span class="sidebar-nav-label">${item.label}</span>
                </a>
            `;
        });
    });

    return `
        <aside class="sidebar" id="sidebar">
            <!-- Logo -->
            <div class="sidebar-logo">
                <div class="sidebar-logo-icon">⛳</div>
                <div class="sidebar-logo-text">
                    <h1>Golf Score Lottery</h1>
                    <p>${roleLabel}</p>
                </div>
            </div>

            <!-- Navigation -->
            <nav class="sidebar-nav">
                ${navHTML}
            </nav>

            <!-- User Info + Logout -->
            <div class="sidebar-footer">
                <div class="sidebar-user-info">
                    <div class="sidebar-avatar">${initial}</div>
                    <div class="sidebar-user-details">
                        <span class="sidebar-user-name">${displayName}</span>
                        <span class="sidebar-user-email">${displayEmail}</span>
                    </div>
                </div>
                <button class="sidebar-logout-btn" id="sidebar-logout-btn" title="Sign Out">
                    🚪
                </button>
            </div>
        </aside>

        <!-- Mobile overlay -->
        <div class="sidebar-overlay" id="sidebar-overlay"></div>
    `;
}

/**
 * Initialize sidebar interactivity.
 * Call this after the sidebar HTML is in the DOM.
 */
export function initSidebar() {
    const sidebar = document.getElementById('sidebar');
    const overlay = document.getElementById('sidebar-overlay');
    const toggleBtn = document.getElementById('sidebar-toggle');

    // Mobile toggle
    if (toggleBtn) {
        toggleBtn.addEventListener('click', () => {
            sidebar.classList.toggle('open');
            overlay.classList.toggle('visible');
        });
    }

    // Close sidebar on overlay click (mobile)
    if (overlay) {
        overlay.addEventListener('click', () => {
            sidebar.classList.remove('open');
            overlay.classList.remove('visible');
        });
    }

    // Close sidebar on nav item click (mobile)
    document.querySelectorAll('.sidebar-nav-item').forEach(item => {
        item.addEventListener('click', () => {
            if (window.innerWidth <= 768) {
                sidebar.classList.remove('open');
                overlay.classList.remove('visible');
            }
        });
    });

    // Update active state on hash change  
    updateActiveNavItem();
}

/**
 * Update the active state of sidebar nav items
 * based on the current hash.
 */
export function updateActiveNavItem() {
    const currentHash = window.location.hash.slice(1) || '/dashboard';
    document.querySelectorAll('.sidebar-nav-item').forEach(item => {
        const path = item.dataset.path;
        if (path === currentHash) {
            item.classList.add('active');
        } else {
            item.classList.remove('active');
        }
    });
}
