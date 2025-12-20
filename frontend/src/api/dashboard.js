import api from './client'

export default {
    // Get widget configurations
    getWidgetConfigs() {
        return api.get('/admin/dashboard-widgets')
    },

    // Update widget config
    updateWidgetConfig(id, data) {
        return api.put(`/admin/dashboard-widgets/${id}`, data)
    },

    // Bulk update widget order
    bulkUpdateOrder(updates) {
        return api.post('/admin/dashboard-widgets/bulk-order', updates)
    },
}
