// ============================================
// Navigation Configuration — Role-based menu
// ============================================

export const navSections = [
    {
        title: null,
        items: [
            { path: '/dashboard', icon: '📊', label: 'Dashboard', roles: ['user', 'admin'] },
        ],
    },
    {
        title: 'Profile Management',
        items: [
            { path: '/profile', icon: '👤', label: 'My Profile', roles: ['user', 'admin'] },
            { path: '/admin/users', icon: '👥', label: 'User Management', roles: ['admin'] },
        ],
    },
    {
        title: 'Operations',
        items: [
            { path: '/scores', icon: '📈', label: 'My Scores', roles: ['user'] },
            { path: '/subscription', icon: '💳', label: 'Subscription', roles: ['user'] },
            { path: '/charity', icon: '❤️', label: 'Charity Selection', roles: ['user'] },
            { path: '/admin/scores', icon: '✏️', label: 'Score Control', roles: ['admin'] },
            { path: '/admin/subscriptions', icon: '💰', label: 'Subscription Tracking', roles: ['admin'] },
            { path: '/admin/verify-winners', icon: '🏆', label: 'Verify Winners', roles: ['admin'] },
            { path: '/admin/draws', icon: '🔀', label: 'Draw Management', roles: ['admin'] },
            { path: '/admin/charities', icon: '🏛️', label: 'Charity Management', roles: ['admin'] },
        ],
    },
    {
        title: 'Analytics & Reports',
        items: [
            { path: '/my-stats', icon: '📊', label: 'My Statistics', roles: ['user'] },
            { path: '/admin/activity-logs', icon: '📄', label: 'Activity Logs', roles: ['admin'] },
            { path: '/admin/reports', icon: '📉', label: 'System Reports', roles: ['admin'] },
        ],
    },
];

/**
 * Filter nav sections based on user role.
 * Returns only sections that have at least one visible item for the role.
 */
export function getFilteredNav(role) {
    return navSections
        .map(section => ({
            ...section,
            items: section.items.filter(item => item.roles.includes(role)),
        }))
        .filter(section => section.items.length > 0);
}
