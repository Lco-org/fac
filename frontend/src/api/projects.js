import api from './client'

export default {
    // List all projects
    list() {
        return api.get('/projects')
    },

    // Get single project
    get(id) {
        return api.get(`/projects/${id}`)
    },

    // Create new project
    create(data) {
        return api.post('/projects', data)
    },

    // Update project
    update(id, data) {
        return api.put(`/projects/${id}`, data)
    },

    // Delete project
    delete(id) {
        return api.delete(`/projects/${id}`)
    },
}
