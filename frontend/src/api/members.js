import api from './client'

export default {
    // List all members
    list() {
        return api.get('/members')
    },

    // Get single member
    get(id) {
        return api.get(`/members/${id}`)
    },

    // Create new member
    create(data) {
        return api.post('/members', data)
    },

    // Update member
    update(id, data) {
        return api.put(`/members/${id}`, data)
    },

    // Delete member
    delete(id) {
        return api.delete(`/members/${id}`)
    },

    // Get expiring members
    getExpiring() {
        return api.get('/members/expiring')
    },

    // Renew membership
    renew(id, data) {
        return api.post(`/members/${id}/renew`, data)
    },

    // Get member subscriptions
    getSubscriptions(id) {
        return api.get(`/members/${id}/subscriptions`)
    },
}
