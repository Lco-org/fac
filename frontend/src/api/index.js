// Unified API module - exports all services
import api from './client'
import activities from './activities'
import attachments from './attachments'
import backups from './backups'
import containers from './containers'
import dashboard from './dashboard'
import email from './email'
import employees from './employees'
import hr from './hr'
import invoices from './invoices'
import legal from './legal'
import members from './members'
import notifications from './notifications'
import organization from './organization'
import projects from './projects'
import receipts from './receipts'
import reports from './reports'
import ssa from './ssa'
import transactions from './transactions'
import transfers from './transfers'
import users from './users'

export {
    api as default,
    activities,
    attachments,
    backups,
    containers,
    dashboard,
    email,
    employees,
    hr,
    invoices,
    legal,
    members,
    notifications,
    organization,
    projects,
    receipts,
    reports,
    ssa,
    transactions,
    transfers,
    users,
}
