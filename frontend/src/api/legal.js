import api from './client'

export default {
    getPrivacyPolicy() {
        return api.get('/legal/privacy-policy')
    },

    getTermsOfService() {
        return api.get('/legal/terms-of-service')
    },
}
