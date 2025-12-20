import api from './client'

export default {
    // List all employees
    list() {
        return api.get('/employees')
    },

    // Get single employee
    get(id) {
        return api.get(`/employees/${id}`)
    },

    // Create new employee
    create(data) {
        return api.post('/employees', data)
    },

    // Update employee
    update(id, data) {
        return api.put(`/employees/${id}`, data)
    },

    // Delete employee
    delete(id) {
        return api.delete(`/employees/${id}`)
    },
}
