import api from './client'

export default {
    // Upload an attachment
    upload(entityType, entityID, file, onUploadProgress) {
        const formData = new FormData()
        formData.append('file', file)
        formData.append('entity_type', entityType)
        formData.append('entity_id', entityID)

        return api.post('/attachments/upload', formData, {
            headers: {
                'Content-Type': 'multipart/form-data',
            },
            onUploadProgress,
        })
    },

    // List attachments
    list(entityType, entityID) {
        return api.get('/attachments', {
            params: {
                entity_type: entityType,
                entity_id: entityID,
            },
        })
    },

    // Delete attachment
    delete(id) {
        return api.delete(`/attachments/${id}`)
    },

    // Get download URL
    getDownloadURL(id) {
        return `/api/v1/attachments/${id}/download`
    }
}
