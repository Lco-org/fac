<template>
  <div class="invoices-management">
    <!-- Header -->
    <div class="page-header">
      <div class="header-content">
        <h1>
          <i class="fas fa-file-invoice"></i>
          إدارة الفواتير
        </h1>
        <p class="subtitle">عرض وإنشاء وطباعة الفواتير المالية</p>
      </div>
      <div class="header-stats">
        <div class="stat-card">
          <span class="stat-value">{{ invoices.length }}</span>
          <span class="stat-label">إجمالي الفواتير</span>
        </div>
      </div>
    </div>

    <!-- Filters -->
    <div class="filters-section">
      <div class="search-box">
        <i class="fas fa-search"></i>
        <input 
          v-model="searchQuery" 
          type="text" 
          placeholder="بحث برقم الفاتورة أو الوصف..."
          @input="handleSearch"
        />
      </div>
      <div class="filter-actions">
        <button class="btn btn-secondary" @click="refreshInvoices">
          <i class="fas fa-sync-alt"></i>
          تحديث
        </button>
      </div>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="loading-state">
      <div class="spinner"></div>
      <p>جاري تحميل الفواتير...</p>
    </div>

    <!-- Error State -->
    <div v-else-if="error" class="error-state">
      <i class="fas fa-exclamation-triangle"></i>
      <p>{{ error }}</p>
      <button class="btn btn-primary" @click="fetchInvoices">إعادة المحاولة</button>
    </div>

    <!-- Empty State -->
    <div v-else-if="filteredInvoices.length === 0" class="empty-state">
      <i class="fas fa-file-invoice"></i>
      <h3>لا توجد فواتير</h3>
      <p>لم يتم إنشاء أي فواتير بعد</p>
    </div>

    <!-- Invoices Table -->
    <div v-else class="invoices-table-container">
      <table class="invoices-table">
        <thead>
          <tr>
            <th>رقم الفاتورة</th>
            <th>التاريخ</th>
            <th>النوع</th>
            <th>المبلغ</th>
            <th>العملة</th>
            <th>الوصف</th>
            <th>الإجراءات</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="invoice in filteredInvoices" :key="invoice.id">
            <td class="invoice-number">
              <span class="badge badge-primary">{{ invoice.invoice_number }}</span>
            </td>
            <td>{{ formatDate(invoice.created_at) }}</td>
            <td>
              <span :class="['type-badge', getTypeBadgeClass(invoice.transaction_type)]">
                {{ getTypeLabel(invoice.transaction_type) }}
              </span>
            </td>
            <td class="amount">{{ formatAmount(invoice.amount) }}</td>
            <td>{{ invoice.currency }}</td>
            <td class="description">{{ invoice.description || '-' }}</td>
            <td class="actions">
              <button 
                class="btn-icon" 
                @click="viewInvoice(invoice)"
                title="عرض التفاصيل"
              >
                <i class="fas fa-eye"></i>
              </button>
              <button 
                class="btn-icon" 
                @click="printInvoice(invoice)"
                title="طباعة"
              >
                <i class="fas fa-print"></i>
              </button>
            </td>
          </tr>
        </tbody>
      </table>

      <!-- Pagination -->
      <div class="pagination">
        <button 
          class="btn btn-secondary" 
          :disabled="offset === 0"
          @click="previousPage"
        >
          <i class="fas fa-chevron-right"></i>
          السابق
        </button>
        <span class="page-info">صفحة {{ currentPage }}</span>
        <button 
          class="btn btn-secondary" 
          :disabled="invoices.length < limit"
          @click="nextPage"
        >
          التالي
          <i class="fas fa-chevron-left"></i>
        </button>
      </div>
    </div>

    <!-- Invoice Detail Modal -->
    <div v-if="selectedInvoice" class="modal-overlay" @click.self="closeModal">
      <div class="modal-content invoice-modal">
        <div class="modal-header">
          <h2>
            <i class="fas fa-file-invoice"></i>
            تفاصيل الفاتورة
          </h2>
          <button class="btn-close" @click="closeModal">
            <i class="fas fa-times"></i>
          </button>
        </div>
        <div class="modal-body">
          <div class="invoice-details">
            <div class="detail-row">
              <span class="label">رقم الفاتورة:</span>
              <span class="value">{{ selectedInvoice.invoice_number }}</span>
            </div>
            <div class="detail-row">
              <span class="label">التاريخ:</span>
              <span class="value">{{ formatDate(selectedInvoice.created_at) }}</span>
            </div>
            <div class="detail-row">
              <span class="label">المبلغ:</span>
              <span class="value amount-large">
                {{ formatAmount(selectedInvoice.amount) }} {{ selectedInvoice.currency }}
              </span>
            </div>
            <div class="detail-row">
              <span class="label">النوع:</span>
              <span class="value">{{ getTypeLabel(selectedInvoice.transaction_type) }}</span>
            </div>
            <div class="detail-row" v-if="selectedInvoice.payee_payer">
              <span class="label">المستفيد/الدافع:</span>
              <span class="value">{{ selectedInvoice.payee_payer }}</span>
            </div>
            <div class="detail-row" v-if="selectedInvoice.description">
              <span class="label">الوصف:</span>
              <span class="value">{{ selectedInvoice.description }}</span>
            </div>
            <div class="detail-row" v-if="selectedInvoice.project_name">
              <span class="label">المشروع:</span>
              <span class="value">{{ selectedInvoice.project_name }}</span>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="closeModal">إغلاق</button>
          <button class="btn btn-primary" @click="printInvoice(selectedInvoice)">
            <i class="fas fa-print"></i>
            طباعة
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { invoices as invoicesApi } from '@/api'

// State
const invoices = ref([])
const loading = ref(true)
const error = ref(null)
const searchQuery = ref('')
const selectedInvoice = ref(null)
const limit = ref(20)
const offset = ref(0)

// Computed
const currentPage = computed(() => Math.floor(offset.value / limit.value) + 1)

const filteredInvoices = computed(() => {
  if (!searchQuery.value) return invoices.value
  const query = searchQuery.value.toLowerCase()
  return invoices.value.filter(inv => 
    inv.invoice_number?.toLowerCase().includes(query) ||
    inv.description?.toLowerCase().includes(query) ||
    inv.payee_payer?.toLowerCase().includes(query)
  )
})

// Methods
async function fetchInvoices() {
  loading.value = true
  error.value = null
  try {
    const response = await invoicesApi.list(limit.value, offset.value)
    invoices.value = response.data.invoices || []
  } catch (err) {
    error.value = 'فشل في تحميل الفواتير. يرجى المحاولة مرة أخرى.'
    console.error('Error fetching invoices:', err)
  } finally {
    loading.value = false
  }
}

function refreshInvoices() {
  offset.value = 0
  fetchInvoices()
}

function handleSearch() {
  // Client-side filtering via computed property
}

function viewInvoice(invoice) {
  selectedInvoice.value = invoice
}

function closeModal() {
  selectedInvoice.value = null
}

function printInvoice(invoice) {
  const printUrl = invoicesApi.getPrintURL(invoice.id)
  window.open(printUrl, '_blank')
}

function previousPage() {
  if (offset.value >= limit.value) {
    offset.value -= limit.value
    fetchInvoices()
  }
}

function nextPage() {
  offset.value += limit.value
  fetchInvoices()
}

function formatDate(dateStr) {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleDateString('ar-LY', {
    year: 'numeric',
    month: 'short',
    day: 'numeric'
  })
}

function formatAmount(amount) {
  if (amount == null) return '0.00'
  return new Intl.NumberFormat('ar-LY', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2
  }).format(amount)
}

function getTypeLabel(type) {
  const labels = {
    'income_grant': 'منحة',
    'income_donation': 'تبرع',
    'disbursement_custody': 'عهدة',
    'disbursement_salary': 'راتب',
    'disbursement_operational': 'تشغيلي',
    'disbursement_bonus': 'مكافأة',
    'transfer': 'تحويل'
  }
  return labels[type] || type
}

function getTypeBadgeClass(type) {
  if (type?.startsWith('income')) return 'type-income'
  if (type?.startsWith('disbursement')) return 'type-disbursement'
  return 'type-default'
}

// Lifecycle
onMounted(() => {
  fetchInvoices()
})
</script>

<style scoped>
.invoices-management {
  padding: 1.5rem;
  direction: rtl;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 2rem;
  padding: 1.5rem;
  background: linear-gradient(135deg, #1e3a5f 0%, #2d5a87 100%);
  border-radius: 12px;
  color: white;
}

.header-content h1 {
  margin: 0;
  font-size: 1.75rem;
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.header-content .subtitle {
  margin: 0.5rem 0 0;
  opacity: 0.8;
  font-size: 0.95rem;
}

.header-stats {
  display: flex;
  gap: 1rem;
}

.stat-card {
  background: rgba(255, 255, 255, 0.15);
  padding: 1rem 1.5rem;
  border-radius: 8px;
  text-align: center;
  backdrop-filter: blur(10px);
}

.stat-value {
  display: block;
  font-size: 1.75rem;
  font-weight: 700;
}

.stat-label {
  font-size: 0.85rem;
  opacity: 0.9;
}

.filters-section {
  display: flex;
  gap: 1rem;
  margin-bottom: 1.5rem;
  flex-wrap: wrap;
}

.search-box {
  flex: 1;
  min-width: 250px;
  position: relative;
}

.search-box i {
  position: absolute;
  right: 1rem;
  top: 50%;
  transform: translateY(-50%);
  color: #64748b;
}

.search-box input {
  width: 100%;
  padding: 0.75rem 2.5rem 0.75rem 1rem;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  font-size: 0.95rem;
  transition: all 0.2s;
}

.search-box input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.btn {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem 1.25rem;
  border: none;
  border-radius: 8px;
  font-size: 0.95rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-primary {
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
  color: white;
}

.btn-primary:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(37, 99, 235, 0.3);
}

.btn-secondary {
  background: #f1f5f9;
  color: #475569;
}

.btn-secondary:hover {
  background: #e2e8f0;
}

.btn-secondary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.loading-state,
.error-state,
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 4rem 2rem;
  text-align: center;
  background: white;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.spinner {
  width: 48px;
  height: 48px;
  border: 4px solid #e2e8f0;
  border-top-color: #3b82f6;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.error-state i,
.empty-state i {
  font-size: 3rem;
  color: #94a3b8;
  margin-bottom: 1rem;
}

.error-state i {
  color: #ef4444;
}

.invoices-table-container {
  background: white;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  overflow: hidden;
}

.invoices-table {
  width: 100%;
  border-collapse: collapse;
}

.invoices-table th,
.invoices-table td {
  padding: 1rem;
  text-align: right;
  border-bottom: 1px solid #f1f5f9;
}

.invoices-table th {
  background: #f8fafc;
  font-weight: 600;
  color: #475569;
  font-size: 0.85rem;
  text-transform: uppercase;
}

.invoices-table tbody tr:hover {
  background: #f8fafc;
}

.invoice-number .badge {
  display: inline-block;
  padding: 0.35rem 0.75rem;
  border-radius: 6px;
  font-size: 0.85rem;
  font-weight: 600;
}

.badge-primary {
  background: #dbeafe;
  color: #1e40af;
}

.type-badge {
  display: inline-block;
  padding: 0.25rem 0.75rem;
  border-radius: 20px;
  font-size: 0.8rem;
  font-weight: 500;
}

.type-income {
  background: #dcfce7;
  color: #166534;
}

.type-disbursement {
  background: #fee2e2;
  color: #991b1b;
}

.type-default {
  background: #f1f5f9;
  color: #475569;
}

.amount {
  font-weight: 600;
  font-family: 'Roboto Mono', monospace;
}

.description {
  max-width: 200px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.actions {
  display: flex;
  gap: 0.5rem;
}

.btn-icon {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  border-radius: 8px;
  background: #f1f5f9;
  color: #475569;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-icon:hover {
  background: #3b82f6;
  color: white;
}

.pagination {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 1rem;
  padding: 1rem;
  border-top: 1px solid #f1f5f9;
}

.page-info {
  color: #64748b;
  font-size: 0.9rem;
}

/* Modal */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  backdrop-filter: blur(4px);
}

.modal-content {
  background: white;
  border-radius: 16px;
  width: 90%;
  max-width: 500px;
  max-height: 90vh;
  overflow: hidden;
  animation: slideUp 0.3s ease-out;
}

@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.25rem 1.5rem;
  border-bottom: 1px solid #f1f5f9;
}

.modal-header h2 {
  margin: 0;
  font-size: 1.25rem;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.btn-close {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  background: #f1f5f9;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-close:hover {
  background: #e2e8f0;
}

.modal-body {
  padding: 1.5rem;
  max-height: 60vh;
  overflow-y: auto;
}

.invoice-details {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.detail-row {
  display: flex;
  justify-content: space-between;
  padding: 0.75rem;
  background: #f8fafc;
  border-radius: 8px;
}

.detail-row .label {
  font-weight: 500;
  color: #64748b;
}

.detail-row .value {
  font-weight: 600;
  color: #1e293b;
}

.amount-large {
  font-size: 1.25rem;
  color: #059669 !important;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  padding: 1rem 1.5rem;
  border-top: 1px solid #f1f5f9;
}

/* Dark mode support */
@media (prefers-color-scheme: dark) {
  .invoices-table-container,
  .modal-content {
    background: #1e293b;
  }

  .invoices-table th {
    background: #0f172a;
    color: #94a3b8;
  }

  .invoices-table td {
    color: #e2e8f0;
    border-bottom-color: #334155;
  }

  .search-box input {
    background: #1e293b;
    border-color: #334155;
    color: #e2e8f0;
  }
}
</style>
