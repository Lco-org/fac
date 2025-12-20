import api from './client'

export default {
    // Create new transaction
    create(data) {
        return api.post('/transactions', data)
    },

    // Get single transaction
    get(id) {
        return api.get(`/transactions/${id}`)
    },

    // Reverse transaction
    reverse(id, reason) {
        return api.post(`/transactions/${id}/reverse`, { reason })
    },
}
