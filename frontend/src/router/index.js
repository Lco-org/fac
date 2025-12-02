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
        path: '/register',
        name: 'Register',
        component: () => import('@/views/RegisterView.vue'),
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
                name: 'Salaries',
                component: () => import('@/views/disbursements/SalariesView.vue'),
                meta: { requiresRole: ['admin', 'finance'] },
            },
            {
                path: 'disbursements/operational',
                name: 'Operational',
                component: () => import('@/views/disbursements/OperationalView.vue'),
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
                path: 'reports/detailed-statement',
                name: 'DetailedStatement',
                component: () => import('@/views/reports/DetailedStatementView.vue'),
            },
        ],
    },
]

const router = createRouter({
    history: createWebHistory(),
    routes,
})

// Navigation guard
router.beforeEach(async (to, from, next) => {
    const authStore = useAuthStore()

    if (to.meta.requiresAuth) {
        if (!authStore.isAuthenticated) {
            // Try to fetch current user
            const success = await authStore.fetchCurrentUser()
            if (!success) {
                return next('/login')
            }
        }

        // Check role-based access
        if (to.meta.requiresRole) {
            const hasRequiredRole = to.meta.requiresRole.includes(authStore.userRole)
            if (!hasRequiredRole) {
                return next('/')
            }
        }
    }

    if (to.path === '/login' && authStore.isAuthenticated) {
        return next('/')
    }

    next()
})

export default router
