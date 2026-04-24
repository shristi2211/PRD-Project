// ============================================
// SPA Router — Hash-based routing with sidebar
// ============================================

import { isAuthenticated, getUser } from '/js/auth.js';
import { renderLoginPage, initLoginPage } from '/js/pages/login.js';
import { renderRegisterPage, initRegisterPage } from '/js/pages/register.js';
import { renderDashboardPage, initDashboardPage } from '/js/pages/dashboard.js';
import { renderSetupPage, initSetupPage } from '/js/pages/setup.js';
import { renderPlaceholderPage, initPlaceholderPage } from '/js/pages/placeholder.js';
import { renderProfilePage, initProfilePage } from '/js/pages/profile.js';
import { renderAdminUsersPage, initAdminUsersPage } from '/js/pages/admin-users.js';
import { renderScores, initScores } from '/js/pages/scores.js';
import { renderCharity, initCharity } from '/js/pages/charity.js';
import { renderMyStats, initMyStats } from '/js/pages/my-stats.js';
import { renderAdminScores, initAdminScores } from '/js/pages/admin-scores.js';
import { renderAdminDraws, initAdminDraws } from '/js/pages/admin-draws.js';
import { renderAdminCharities, initAdminCharities } from '/js/pages/admin-charities.js';
import { renderAdminVerifyWinners, initAdminVerifyWinners } from '/js/pages/admin-verify-winners.js';
import { renderAdminActivityLogs, initAdminActivityLogs } from '/js/pages/admin-activity-logs.js';
import { renderAdminReports, initAdminReports } from '/js/pages/admin-reports.js';
import { renderDashboardLayout, initDashboardLayout } from '/js/components/layout.js';
import { renderSubscription, initSubscription } from '/js/pages/subscription.js';
import { renderLandingPage, initLandingPage } from '/js/pages/landing.js';
import { updateActiveNavItem } from '/js/components/sidebar.js';
import { renderAdminSubscriptions, initAdminSubscriptions } from '/js/pages/admin-subscriptions.js';

const app = document.getElementById('app');

// ─── Route Definitions ───────────────────────────
// Public routes (no layout)
const publicRoutes = {
    '/': {
        render: renderLandingPage,
        init: initLandingPage,
        title: 'Welcome',
    },
    '/login': {
        render: renderLoginPage,
        init: initLoginPage,
        title: 'Sign In',
    },
    '/register': {
        render: renderRegisterPage,
        init: initRegisterPage,
        title: 'Create Account',
    },
    '/setup': {
        render: renderSetupPage,
        init: initSetupPage,
        title: 'Admin Setup',
    },
};

// Protected routes (wrapped in sidebar layout)
const protectedRoutes = {
    '/dashboard': {
        render: renderDashboardPage,
        init: initDashboardPage,
        title: 'Dashboard',
        roles: ['user', 'admin'],
    },
    '/profile': {
        render: renderProfilePage,
        init: initProfilePage,
        title: 'My Profile',
        roles: ['user', 'admin'],
    },
    '/scores': {
        render: renderScores,
        init: initScores,
        title: 'My Scores',
        roles: ['user'],
    },
    '/subscription': {
        render: renderSubscription,
        init: initSubscription,
        title: 'Subscription',
        roles: ['user'],
    },
    '/charity': {
        render: renderCharity,
        init: initCharity,
        title: 'Charity Selection',
        roles: ['user'],
    },
    '/my-stats': {
        render: renderMyStats,
        init: initMyStats,
        title: 'My Statistics',
        roles: ['user'],
    },
    '/admin/users': {
        render: renderAdminUsersPage,
        init: initAdminUsersPage,
        title: 'User Management',
        roles: ['admin'],
    },
    '/admin/scores': {
        render: renderAdminScores,
        init: initAdminScores,
        title: 'Score Control',
        roles: ['admin'],
    },
    '/admin/subscriptions': {
        render: renderAdminSubscriptions,
        init: initAdminSubscriptions,
        title: 'Subscription Tracking',
        roles: ['admin'],
    },
    '/admin/verify-winners': {
        render: renderAdminVerifyWinners,
        init: initAdminVerifyWinners,
        title: 'Verify Winners',
        roles: ['admin'],
    },
    '/admin/draws': {
        render: renderAdminDraws,
        init: initAdminDraws,
        title: 'Draw Management',
        roles: ['admin'],
    },
    '/admin/charities': {
        render: renderAdminCharities,
        init: initAdminCharities,
        title: 'Charity Management',
        roles: ['admin'],
    },
    '/admin/activity-logs': {
        render: renderAdminActivityLogs,
        init: initAdminActivityLogs,
        title: 'Activity Logs',
        roles: ['admin'],
    },
    '/admin/reports': {
        render: renderAdminReports,
        init: initAdminReports,
        title: 'System Reports',
        roles: ['admin'],
    },
};

// Track if layout is already rendered (avoid full re-render on protected→protected navigation)
let layoutRendered = false;

/**
 * Navigate to a route based on the current hash.
 */
function navigate() {
    const hash = window.location.hash.slice(1) || '/';

    // ─── Check public routes ───────────────────
    const publicRoute = publicRoutes[hash];
    if (publicRoute) {
        // If authenticated and trying to hit core public pages → go to dashboard
        if (isAuthenticated() && (hash === '/' || hash === '/login' || hash === '/register')) {
            window.location.hash = '#/dashboard';
            return;
        }

        layoutRendered = false;
        app.innerHTML = publicRoute.render();
        if (publicRoute.init) publicRoute.init();
        document.title = `${publicRoute.title} — Golf Score Lottery`;
        return;
    }

    // ─── Check protected routes ────────────────
    const protectedRoute = protectedRoutes[hash];

    // Unknown route → redirect to landing
    if (!protectedRoute) {
        window.location.hash = '#/';
        return;
    }

    // Auth guard → redirect to landing
    if (!isAuthenticated()) {
        window.location.hash = '#/';
        return;
    }

    // Role guard
    const user = getUser();
    const userRole = user?.role || 'user';
    if (protectedRoute.roles && !protectedRoute.roles.includes(userRole)) {
        // User trying to access admin route → redirect to dashboard
        window.location.hash = '#/dashboard';
        return;
    }

    // ─── Render protected page inside layout ───
    const contentHTML = protectedRoute.render();

    if (layoutRendered) {
        // Layout already exists — only swap the content area
        const contentContainer = document.querySelector('.dashboard-content');
        const topbarTitle = document.querySelector('.topbar-title');
        if (contentContainer) {
            contentContainer.innerHTML = contentHTML;
        }
        if (topbarTitle) {
            topbarTitle.textContent = protectedRoute.title;
        }
        // Update sidebar active state
        updateActiveNavItem();
    } else {
        // First protected page load — render full layout
        app.innerHTML = renderDashboardLayout(contentHTML, protectedRoute.title);
        initDashboardLayout();
        layoutRendered = true;
    }

    // Init page-specific logic
    if (protectedRoute.init) {
        protectedRoute.init();
    }

    document.title = `${protectedRoute.title} — Golf Score Lottery`;
}

// Listen for hash changes
window.addEventListener('hashchange', navigate);

// Initial navigation
navigate();
