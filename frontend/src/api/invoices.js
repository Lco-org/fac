import api from './client'

export default {
    // List invoices
    list(limit = 20, offset = 0) {
        return api.get('/invoices', {
            params: { limit, offset }
        })
    },

    // Get single invoice
    get(id) {
        return api.get(`/invoices/${id}`)
    },

    // Generate from transaction
    generateFromTransaction(transactionID) {
        return api.post(`/invoices/from-transaction/${transactionID}`)
    },

    // Get print URL
    getPrintURL(id) {
        return `/api/v1/invoices/${id}/print`
    }
}
