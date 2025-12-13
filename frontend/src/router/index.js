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
                path: 'account',
                name: 'AccountSettings',
                component: () => import('@/views/account/AccountSettingsView.vue'),
                meta: { requiresAuth: true },
            },
            {
                path: 'account/security',
                name: 'AccountSecurity',
                redirect: '/account',
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
            // ==========================================================================
            // ORGANIZATION ADMIN SETTINGS ROUTES
            // ==========================================================================
            {
                path: 'admin/email-templates',
                name: 'EmailTemplates',
                component: () => import('@/views/admin/EmailTemplateEditorView.vue'),
                meta: { requiresRole: ['admin'], title: 'قوالب البريد الإلكتروني' },
            },
            {
                path: 'admin/smtp-settings',
                name: 'SMTPSettings',
                component: () => import('@/views/admin/SMTPSettingsView.vue'),
                meta: { requiresRole: ['admin'], title: 'إعدادات SMTP' },
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
    // ==========================================================================
    // SUPER SYSTEM ADMIN (SSA) ROUTES - Hidden panel with secret URL
    // ==========================================================================
    // SSA Login - No auth required (it's for initial authentication)
    {
        path: '/b8fb17/ssa/login',
        name: 'SSALogin',
        component: () => import('@/views/ssa/SSALoginView.vue'),
        meta: { title: 'تسجيل دخول المشرف الأعلى' },
    },
    // TEMPORARY: SSA Registration (remove in production!)
    {
        path: '/b8fb17/ssa/temp-register',
        name: 'SSATempRegister',
        component: () => import('@/views/ssa/SSATempRegisterView.vue'),
        meta: { title: 'تسجيل SSA مؤقت' },
    },
    {
        path: '/b8fb17/ssa',
        component: () => import('@/views/ssa/SuperAdminDashboardView.vue'),
        meta: { requiresAuth: true, requiresRole: ['system_admin'] },
        children: [
            {
                path: '',
                name: 'SSADashboard',
                component: () => import('@/views/ssa/SuperAdminDashboardView.vue'),
            },
        ]
    },
    {
        path: '/b8fb17/ssa/create',
        name: 'SSACreateOrganization',
        component: () => import('@/views/ssa/OrganizationCreateView.vue'),
        meta: { requiresAuth: true, requiresRole: ['system_admin'] },
    },
    {
        path: '/b8fb17/ssa/verify/:id',
        name: 'SSAVerifyDomain',
        component: () => import('@/views/ssa/DomainVerificationView.vue'),
        meta: { requiresAuth: true, requiresRole: ['system_admin'] },
    },
    {
        path: '/b8fb17/ssa/database/:id',
        name: 'SSACreateDatabase',
        component: () => import('@/views/ssa/DatabaseSetupView.vue'),
        meta: { requiresAuth: true, requiresRole: ['system_admin'] },
    },
    {
        path: '/b8fb17/ssa/admin/:id',
        name: 'SSACreateAdmin',
        component: () => import('@/views/ssa/TenantAdminSetupView.vue'),
        meta: { requiresAuth: true, requiresRole: ['system_admin'] },
    },
    {
        path: '/b8fb17/ssa/organization/:id',
        name: 'SSAOrganizationDetail',
        component: () => import('@/views/ssa/OrganizationDetailView.vue'),
        meta: { requiresAuth: true, requiresRole: ['system_admin'] },
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
