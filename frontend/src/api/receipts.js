import api from './client'

/**
 * Receipts API Service
 * Connects to backend receipt_handler.go endpoints
 */
export const receipts = {
    /**
     * Get receipt data for a transaction
     * @param {string} transactionId - UUID of the transaction
     * @returns {Promise} Receipt data including transaction details, organization info
     */
    get(transactionId) {
        return api.get(`/receipts/${transactionId}`)
    },

    /**
     * Download receipt as PDF
     * @param {string} transactionId - UUID of the transaction
     * @returns {Promise} PDF blob
     */
    downloadPdf(transactionId) {
        return api.get(`/receipts/${transactionId}/pdf`, {
            responseType: 'blob'
        })
    },
}

export default receipts
