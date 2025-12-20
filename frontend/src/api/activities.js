import api from './client'

export default {
    // Get activities by entity
    getByEntity(entityType, entityId) {
        return api.get('/activities/entity', {
            params: { entity_type: entityType, entity_id: entityId }
        })
    },

    // Get activities by user
    getByUser(userId, limit = 50, offset = 0) {
        return api.get('/activities/user', {
            params: { user_id: userId, limit, offset }
        })
    },
}
