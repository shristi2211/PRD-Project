// ============================================
// Placeholder Page Generator
// ============================================

/**
 * Creates a placeholder page for routes that are under construction.
 * Each page gets a unique icon, title, and description.
 */

const placeholderPages = {
    '/profile': {
        icon: '👤',
        title: 'My Profile',
        description: 'View and manage your personal information, change password, and account settings.',
        features: ['View personal info', 'Edit name & email', 'Change password', 'Delete account'],
    },
    '/scores': {
        icon: '📈',
        title: 'My Scores',
        description: 'Track your golf scores. You can submit up to 5 Stableford scores (1-45 points).',
        features: ['View all 5 scores', 'Add new score', 'Delete unused scores', 'Score history timeline'],
    },
    '/subscription': {
        icon: '💳',
        title: 'Subscription',
        description: 'Manage your subscription plan, view billing history, and upgrade options.',
        features: ['View current plan', 'Upgrade/Downgrade', 'Billing history', 'Cancel subscription'],
    },
    '/charity': {
        icon: '❤️',
        title: 'Charity Selection',
        description: 'Choose your preferred charity and set your contribution percentage.',
        features: ['Select charity', 'Set percentage (min 10%)', 'View impact dashboard'],
    },
    '/my-stats': {
        icon: '📊',
        title: 'My Statistics',
        description: 'View your score progression, draw participation, and winnings summary.',
        features: ['Score progression graph', 'Draw history', 'Winnings summary', 'Charity contribution total'],
    },
    '/admin/users': {
        icon: '👥',
        title: 'User Management',
        description: 'View, search, and manage all registered users in the system.',
        features: ['Paginated user table', 'Search by name/email', 'Filter by status', 'Edit user profiles', 'Activate/Deactivate'],
    },
    '/admin/scores': {
        icon: '✏️',
        title: 'Score Control',
        description: 'Review and manage all user scores. Edit incorrect entries or flag suspicious patterns.',
        features: ['View all user scores', 'Edit incorrect scores', 'Delete fraudulent entries', 'Flag suspicious patterns'],
    },
    '/admin/subscriptions': {
        icon: '💰',
        title: 'Subscription Tracking',
        description: 'Monitor all subscriptions, filter by plan type, and view revenue breakdowns.',
        features: ['View all subscriptions', 'Filter by plan type', 'Active/Expired/Cancelled', 'Revenue breakdown'],
    },
    '/admin/verify-winners': {
        icon: '🏆',
        title: 'Verify Winners',
        description: 'Review winner proof submissions, approve or reject claims, and release payouts.',
        features: ['View pending proofs', 'Review screenshots', 'Approve/Reject claims', 'Release payouts'],
    },
    '/admin/draws': {
        icon: '🔀',
        title: 'Draw Management',
        description: 'Run draw simulations, execute monthly draws, and publish results.',
        features: ['Run simulation', 'Execute monthly draw', 'Publish results', 'View draw history'],
    },
    '/admin/charities': {
        icon: '🏛️',
        title: 'Charity Management',
        description: 'Add, edit, and manage charities. View distribution statistics.',
        features: ['Add new charity', 'Edit charity details', 'Activate/Deactivate', 'Distribution stats'],
    },
    '/admin/activity-logs': {
        icon: '📄',
        title: 'Activity Logs',
        description: 'Monitor user activities with drill-down from user → year → month → events.',
        features: ['User-wise activity sections', 'Timeline view per user', 'Filter by action type', 'Export logs (CSV)'],
    },
    '/admin/reports': {
        icon: '📉',
        title: 'System Reports',
        description: 'Generate and view system-wide reports including revenue, users, and draws.',
        features: ['User count reports', 'Revenue reports', 'Draw performance', 'Export PDF/Excel'],
    },
};

export function renderPlaceholderPage(path) {
    const page = placeholderPages[path];

    if (!page) {
        return `
            <div class="placeholder-page">
                <div class="placeholder-icon">🚧</div>
                <h2>Page Not Found</h2>
                <p>This page doesn't exist yet.</p>
            </div>
        `;
    }

    const featureListHTML = page.features
        .map(f => `<li><span class="feature-check">✓</span> ${f}</li>`)
        .join('');

    return `
        <div class="placeholder-page">
            <div class="placeholder-header">
                <div class="placeholder-icon-large">${page.icon}</div>
                <h2>${page.title}</h2>
                <p>${page.description}</p>
            </div>

            <div class="placeholder-card">
                <h3>🚧 Under Construction</h3>
                <p>This page is being built. Here's what's coming:</p>
                <ul class="feature-list">
                    ${featureListHTML}
                </ul>
            </div>
        </div>
    `;
}

export function initPlaceholderPage() {
    // No interactivity needed for placeholder pages
}
