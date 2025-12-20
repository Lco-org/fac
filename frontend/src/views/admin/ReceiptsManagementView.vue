<template>
  <div class="receipts-management">
    <!-- Header -->
    <div class="page-header">
      <div class="header-content">
        <h1>
          <i class="fas fa-receipt"></i>
          إدارة الإيصالات
        </h1>
        <p class="subtitle">عرض وطباعة إيصالات العمليات المالية</p>
      </div>
    </div>

    <!-- Filters -->
    <div class="filters-section">
      <div class="search-box">
        <i class="fas fa-search"></i>
        <input 
          v-model="searchQuery" 
          type="text" 
          placeholder="بحث برقم العملية أو الوصف..."
          @input="handleSearch"
        />
      </div>
      <div class="filter-actions">
        <button class="btn btn-secondary" @click="refreshReceipts">
          <i class="fas fa-sync-alt"></i>
          تحديث
        </button>
      </div>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="loading-state">
      <div class="spinner"></div>
      <p>جاري تحميل البيانات...</p>
    </div>

    <!-- Error State -->
    <div v-else-if="error" class="error-state">
      <i class="fas fa-exclamation-triangle"></i>
      <p>{{ error }}</p>
      <button class="btn btn-primary" @click="fetchTransactions">إعادة المحاولة</button>
    </div>

    <!-- Empty State -->
    <div v-else-if="filteredTransactions.length === 0" class="empty-state">
      <i class="fas fa-search"></i>
      <h3>لا توجد نتائج</h3>
      <p>لا توجد عمليات مطابقة للبحث</p>
    </div>

    <!-- Receipts Table -->
    <div v-else class="table-container">
      <table class="receipts-table">
        <thead>
          <tr>
            <th>رقم العملية</th>
            <th>التاريخ</th>
            <th>النوع</th>
            <th>المبلغ</th>
            <th>العملة</th>
            <th>الوصف</th>
            <th>الإجراءات</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="t in filteredTransactions" :key="t.id">
            <td class="receipt-number">
              <span class="badge badge-primary">{{ t.transaction_number }}</span>
            </td>
            <td>{{ formatDate(t.transaction_date) }}</td>
            <td>
              <span :class="['type-badge', getTypeBadgeClass(t.type)]">
                {{ getTypeLabel(t.type) }}
              </span>
            </td>
            <td class="amount">{{ formatAmount(t.amount) }}</td>
            <td>{{ t.currency }}</td>
            <td class="description">{{ t.description || '-' }}</td>
            <td class="actions">
              <a 
                :href="`/receipt/${t.id}`" 
                target="_blank"
                class="btn-icon" 
                title="عرض الإيصال"
              >
                <i class="fas fa-file-invoice"></i>
              </a>
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
          :disabled="transactions.length < limit"
          @click="nextPage"
        >
          التالي
          <i class="fas fa-chevron-left"></i>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { transactions as transactionsApi } from '@/api'

// State
const transactions = ref([])
const loading = ref(true)
const error = ref(null)
const searchQuery = ref('')
const limit = ref(20)
const offset = ref(0)
const total = ref(0)

// Computed
const currentPage = computed(() => Math.floor(offset.value / limit.value) + 1)

const filteredTransactions = computed(() => {
  if (!searchQuery.value) return transactions.value
  const query = searchQuery.value.toLowerCase()
  return transactions.value.filter(t => 
    t.transaction_number?.toLowerCase().includes(query) ||
    t.description?.toLowerCase().includes(query)
  )
})

// Methods
async function fetchTransactions() {
  loading.value = true
  error.value = null
  try {
    const response = await transactionsApi.list(limit.value, offset.value)
    transactions.value = response.data.transactions || []
  } catch (err) {
    error.value = 'فشل في تحميل العمليات'
    console.error('Error fetching transactions:', err)
  } finally {
    loading.value = false
  }
}

function refreshReceipts() {
  offset.value = 0
  fetchTransactions()
}

function handleSearch() {
  // Client side for now
}

function previousPage() {
  if (offset.value >= limit.value) {
    offset.value -= limit.value
    fetchTransactions()
  }
}

function nextPage() {
  offset.value += limit.value
  fetchTransactions()
}

function formatDate(dateStr) {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleDateString('ar-LY', {
    numeric: 'numeric',
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
    'transfer': 'تحويل',
    'procurement': 'شراء'
  }
  return labels[type] || type
}

function getTypeBadgeClass(type) {
  if (type?.startsWith('income')) return 'type-income'
  if (type?.startsWith('disbursement')) return 'type-disbursement'
  return 'type-default'
}

onMounted(() => {
  fetchTransactions()
})
</script>

<style scoped>
.receipts-management {
  padding: 1.5rem;
  direction: rtl;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 2rem;
  padding: 1.5rem;
  background: linear-gradient(135deg, #475569 0%, #1e293b 100%);
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

.filters-section {
  display: flex;
  gap: 1rem;
  margin-bottom: 1.5rem;
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

.btn-secondary {
  background: #f1f5f9;
  color: #475569;
}

.btn-secondary:hover {
  background: #e2e8f0;
}

.loading-state, .error-state, .empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 4rem 2rem;
  background: white;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.1);
  text-align: center;
}

.spinner {
  width: 48px;
  height: 48px;
  border: 4px solid #e2e8f0;
  border-top-color: #64748b;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin { to { transform: rotate(360deg); } }

.table-container {
  background: white;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.1);
  overflow: hidden;
}

.receipts-table {
  width: 100%;
  border-collapse: collapse;
}

.receipts-table th, .receipts-table td {
  padding: 1rem;
  text-align: right;
  border-bottom: 1px solid #f1f5f9;
}

.receipts-table th {
  background: #f8fafc;
  color: #475569;
  font-weight: 600;
  font-size: 0.85rem;
}

.receipt-number .badge {
  padding: 0.35rem 0.75rem;
  background: #e2e8f0;
  color: #1e293b;
  border-radius: 6px;
  font-weight: 600;
  font-size: 0.85rem;
}

.type-badge {
  padding: 0.25rem 0.75rem;
  border-radius: 20px;
  font-size: 0.8rem;
  font-weight: 500;
}

.type-income { background: #dcfce7; color: #166534; }
.type-disbursement { background: #fee2e2; color: #991b1b; }
.type-default { background: #f1f5f9; color: #475569; }

.amount {
  font-weight: 600;
  font-family: monospace;
}

.btn-icon {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f1f5f9;
  color: #475569;
  border-radius: 8px;
  text-decoration: none;
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
</style>
