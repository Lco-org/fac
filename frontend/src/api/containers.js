import api from './client'

export default {
    // List all containers
    list() {
        return api.get('/containers')
    },

    // Get single container
    get(id) {
        return api.get(`/containers/${id}`)
    },

    // Create new container
    create(data) {
        return api.post('/containers', data)
    },

    // Update container
    update(id, data) {
        return api.put(`/containers/${id}`, data)
    },

    // Delete container
    delete(id) {
        return api.delete(`/containers/${id}`)
    },
}
