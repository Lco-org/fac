import api from './client'

// SSA Authentication
export const ssaAuth = {
    login(username, password) {
        return api.post('/system-auth/login', { username, password })
    },

    verify2FA(code, tempToken) {
        return api.post('/system-auth/verify-2fa', { code, temp_token: tempToken })
    },

    logout() {
        return api.post('/system-auth/logout')
    },

    getMe() {
        return api.get('/ssa/me')
    },
}

// SSA Organization Management
export const ssaOrganizations = {
    list(status = '', limit = 20, offset = 0) {
        return api.get('/ssa/organizations', {
            params: { status, limit, offset }
        })
    },

    get(id) {
        return api.get(`/ssa/organizations/${id}`)
    },

    create(data) {
        return api.post('/ssa/organizations', data)
    },

    getStats() {
        return api.get('/ssa/stats')
    },

    getPlans() {
        return api.get('/ssa/subscription-plans')
    },

    // Verification
    verify(id, code) {
        return api.post(`/ssa/organizations/${id}/verify`, { code })
    },

    resendVerification(id, email) {
        return api.post(`/ssa/organizations/${id}/resend-verification`, { email })
    },

    // Database
    createDatabase(id) {
        return api.post(`/ssa/organizations/${id}/database`)
    },

    // Admin
    createAdmin(id, adminData) {
        return api.post(`/ssa/organizations/${id}/admin`, adminData)
    },

    // Status
    suspend(id, reason) {
        return api.post(`/ssa/organizations/${id}/suspend`, { reason })
    },

    activate(id) {
        return api.post(`/ssa/organizations/${id}/activate`)
    },

    // Subscription
    updateSubscription(id, data) {
        return api.put(`/ssa/organizations/${id}/subscription`, data)
    },

    // Deletion - Request OTP
    requestDeletion(id, password) {
        return api.post(`/ssa/organizations/${id}/request-deletion`, { password })
    },

    // Deletion - Confirm with OTP
    confirmDeletion(id, code) {
        return api.post(`/ssa/organizations/${id}/confirm-deletion`, { code })
    },
}

// SSA Audit Logs
export const ssaAuditLogs = {
    list(limit = 50, offset = 0) {
        return api.get('/ssa/audit-logs', {
            params: { limit, offset }
        })
    },

}

// SSA Subscription Plans
export const ssaSubscriptionPlans = {
    list() {
        return api.get('/ssa/subscription-plans')
    },

    get(id) {
        return api.get(`/ssa/subscription-plans/${id}`)
    },

    create(data) {
        return api.post('/ssa/subscription-plans', data)
    },

    update(id, data) {
        return api.put(`/ssa/subscription-plans/${id}`, data)
    },

    delete(id) {
        return api.delete(`/ssa/subscription-plans/${id}`)
    },
}

export default {
    auth: ssaAuth,
    organizations: ssaOrganizations,
    auditLogs: ssaAuditLogs,
    plans: ssaSubscriptionPlans,
}
