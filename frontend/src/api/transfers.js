import api from './client'

export default {
    // Create transfer between containers
    create(data) {
        return api.post('/transfers', data)
    },
}
