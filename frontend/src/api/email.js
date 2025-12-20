import api from './client'

export default {
    // Send bulk email
    sendBulk(data) {
        return api.post('/email/bulk', data)
    },
}
