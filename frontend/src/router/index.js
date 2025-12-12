import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const routes = [
    {
        path: '/login',
        name: 'Login',
        component: () => import('@/views/LoginView.vue'),
        meta: { requiresAuth: false },
    },
    {
        path: '/',
        component: () => import('@/layouts/DashboardLayout.vue'),
        meta: { requiresAuth: true },
        children: [
            {
                path: '',
                name: 'Dashboard',
                component: () => import('@/views/DashboardView.vue'),
            },
            {
                path: 'income/grants',
                name: 'Grants',
                component: () => import('@/views/income/GrantsView.vue'),
                meta: { requiresRole: ['admin', 'finance'] },
            },
            {
                path: 'income/donations',
                name: 'Donations',
                component: () => import('@/views/income/DonationsView.vue'),
                meta: { requiresRole: ['admin', 'finance'] },
            },
            {
                path: 'disbursements/custody',
                name: 'Custody',
                component: () => import('@/views/disbursements/CustodyView.vue'),
                meta: { requiresRole: ['admin', 'finance'] },
            },
            {
                path: 'disbursements/salaries',
                name: 'DisbursementsSalaries',
                component: () => import('@/views/disbursements/SalariesView.vue'),
                meta: { requiresRole: ['admin', 'finance'] },
            },
            {
                path: 'hr/salaries',
                name: 'Salaries',
                component: () => import('@/views/hr/SalariesView.vue'),
                meta: { requiresRole: ['admin', 'hr', 'finance'] },
            },
            {
                path: 'disbursements/operational',
                name: 'Operational',
                component: () => import('@/views/disbursements/OperationalView.vue'),
                meta: { requiresRole: ['admin', 'finance'] },
            },
            {
                path: 'transfers',
                name: 'Transfers',
                component: () => import('@/views/transfers/TransfersView.vue'),
                meta: { requiresRole: ['admin', 'finance'] },
            },
            {
                path: 'admin/users',
                name: 'Users',
                component: () => import('@/views/admin/UsersView.vue'),
                meta: { requiresRole: ['admin'] },
            },
            {
                path: 'admin/containers',
                name: 'Containers',
                component: () => import('@/views/admin/ContainersView.vue'),
                meta: { requiresRole: ['admin'] },
            },
            {
                path: 'admin/projects',
                name: 'Projects',
                component: () => import('@/views/admin/ProjectsView.vue'),
                meta: { requiresRole: ['admin'] },
            },
            {
                path: 'admin/employees',
                name: 'Employees',
                component: () => import('@/views/admin/EmployeesView.vue'),
                meta: { requiresRole: ['admin'] },
            },
            {
                path: 'admin/members',
                name: 'Members',
                component: () => import('@/views/admin/MembersView.vue'),
                meta: { requiresRole: ['admin'] },
            },
            {
                path: 'admin/dashboard-config',
                name: 'DashboardConfig',
                component: () => import('@/views/admin/DashboardConfigView.vue'),
                meta: { requiresRole: ['admin'] },
            },
            {
                path: 'admin/messaging',
                name: 'Messaging',
                component: () => import('@/views/admin/MessagingView.vue'),
                meta: { requiresRole: ['admin'] },
            },
            {
                path: 'account/security',
                name: 'AccountSecurity',
                component: () => import('@/views/admin/AccountSecurityView.vue'),
                meta: { requiresAuth: true },
            },
            {
                path: 'hr/employees',
                name: 'HREmployees',
                component: () => import('@/views/hr/HREmployeesView.vue'),
                meta: { requiresRole: ['admin', 'hr'] },
            },
            {
                path: 'hr/deductions',
                name: 'Deductions',
                component: () => import('@/views/hr/DeductionsView.vue'),
                meta: { requiresRole: ['admin', 'hr'] },
            },
            {
                path: 'hr/bonuses',
                name: 'Bonuses',
                component: () => import('@/views/hr/BonusesView.vue'),
                meta: { requiresRole: ['admin', 'hr'] },
            },
            {
                path: 'reports/detailed-statement',
                name: 'DetailedStatement',
                component: () => import('@/views/reports/DetailedStatementView.vue'),
                meta: { title: 'كشف حساب تفصيلي' }
            },
            {
                path: 'reports/daily-position',
                name: 'DailyPosition',
                component: () => import('@/views/reports/DailyPositionView.vue'),
                meta: { title: 'الموقف اليومي' }
            },
            {
                path: 'reports/invoices',
                name: 'InvoicesReport',
                component: () => import('@/views/reports/InvoicesReportView.vue'),
                meta: { title: 'تقرير الفواتير', requiresPermission: 'view_invoices' }
            },
        ],
    },
    {
        path: '/receipt/:id',
        name: 'Receipt',
        component: () => import('@/views/ReceiptView.vue'),
        meta: { requiresAuth: true },
    },
]

const router = createRouter({
    history: createWebHistory(),
    routes,
})

// Navigation guard
router.beforeEach(async (to, from, next) => {
    const authStore = useAuthStore()
    console.log(`Navigating to: ${to.path}, Role: ${authStore.userRole}, Required: ${to.meta.requiresRole}`)

    if (to.meta.requiresAuth || to.path.startsWith('/reports/')) { // Reports typically need auth
        if (!authStore.isAuthenticated) {
            // Try to fetch current user
            const success = await authStore.fetchCurrentUser()
            if (!success) {
                return next('/login')
            }
        }
    }

    // Role checks
    if (to.meta.requiresRole) {
        const hasRequiredRole = to.meta.requiresRole.includes(authStore.userRole)
        if (!hasRequiredRole) {
            return next('/')
        }
    }

    // Permission checks
    if (to.meta.requiresPermission) {
        if (!authStore.hasPermission(to.meta.requiresPermission)) {
            return next('/')
        }
    }

    if (to.path === '/login' && authStore.isAuthenticated) {
        return next('/')
    }

    next()
})

export default router
