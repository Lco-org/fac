import api from './client'

export default {
    // Create backup
    create() {
        return api.post('/backups')
    },

    // List backups
    list() {
        return api.get('/backups')
    },
}
