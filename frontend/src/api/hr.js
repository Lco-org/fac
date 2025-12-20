import api from './client'

export default {
    // Get monthly salaries
    getMonthlySalaries(month) {
        return api.get('/hr/salaries', { params: { month } })
    },

    // Create salary deduction
    createDeduction(data) {
        return api.post('/hr/deductions', data)
    },

    // List deductions
    listDeductions(params = {}) {
        return api.get('/hr/deductions', { params })
    },

    // Create salary bonus
    createBonus(data) {
        return api.post('/hr/bonuses', data)
    },

    // List bonuses
    listBonuses(params = {}) {
        return api.get('/hr/bonuses', { params })
    },

    // Get employee financial profile
    getEmployeeProfile(employeeId) {
        return api.get('/hr/employees/profile', { params: { employee_id: employeeId } })
    },
}
