import api from './client'

export default {
    // Get all notifications
    list() {
        return api.get('/notifications')
    },

    // Get unread count
    getUnreadCount() {
        return api.get('/notifications/unread-count')
    },

    // Mark as read
    markAsRead(id) {
        return api.post(`/notifications/${id}/read`)
    },

    // Mark all as read
    markAllAsRead() {
        return api.post('/notifications/read-all')
    },
}
