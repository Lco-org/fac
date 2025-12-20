import api from './client'

export default {
    // Get dashboard summary
    getDashboard() {
        return api.get('/reports/dashboard')
    },

    // Get container statement
    getContainerStatement(id, params = {}) {
        return api.get(`/reports/containers/${id}/statement`, { params })
    },

    // Get project statement
    getProjectStatement(id, params = {}) {
        return api.get(`/reports/projects/${id}/statement`, { params })
    },

    // Get employee statement
    getEmployeeStatement(id, params = {}) {
        return api.get(`/reports/employees/${id}/statement`, { params })
    },

    // Get member statement
    getMemberStatement(id, params = {}) {
        return api.get(`/reports/members/${id}/statement`, { params })
    },

    // Get daily position
    getDailyPosition() {
        return api.get('/reports/daily-position')
    },

    // Get project summary
    getProjectSummary() {
        return api.get('/reports/project-summary')
    },

    // Get invoices report
    getInvoices(params = {}) {
        return api.get('/reports/invoices', { params })
    },
}
