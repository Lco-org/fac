import api from './client'

export default {
    // List all users
    list() {
        return api.get('/users')
    },

    // Get single user
    get(id) {
        return api.get(`/users/${id}`)
    },

    // Create new user
    create(data) {
        return api.post('/users', data)
    },

    // Update user
    update(id, data) {
        return api.put(`/users/${id}`, data)
    },

    // Delete user
    delete(id) {
        return api.delete(`/users/${id}`)
    },

    // Update user role
    updateRole(id, role) {
        return api.patch(`/users/${id}/role`, { role })
    },

    // Get current user
    getMe() {
        return api.get('/users/me')
    },
}
