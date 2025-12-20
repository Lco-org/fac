import api from './client'

// Email Templates API
export const emailTemplates = {
    // List all templates
    list() {
        return api.get('/organization/email-templates')
    },

    // Get single template
    get(id) {
        return api.get(`/organization/email-templates/${id}`)
    },

    // Update template
    update(id, data) {
        return api.put(`/organization/email-templates/${id}`, data)
    },

    // Preview template
    preview(data) {
        return api.post('/organization/email-templates/preview', data)
    },

    // Reset template to default
    reset(key) {
        return api.post(`/organization/email-templates/${key}/reset`)
    },
}

// SMTP Settings API
export const smtpSettings = {
    // List all SMTP settings
    list() {
        return api.get('/organization/smtp-settings')
    },

    // Get single SMTP setting
    get(id) {
        return api.get(`/organization/smtp-settings/${id}`)
    },

    // Create SMTP setting
    create(data) {
        return api.post('/organization/smtp-settings', data)
    },

    // Update SMTP setting
    update(id, data) {
        return api.put(`/organization/smtp-settings/${id}`, data)
    },

    // Delete SMTP setting
    delete(id) {
        return api.delete(`/organization/smtp-settings/${id}`)
    },

    // Test SMTP connection
    test(id, toEmail) {
        return api.post(`/organization/smtp-settings/${id}/test`, { to_email: toEmail })
    },
}

export default {
    emailTemplates,
    smtpSettings,
}
