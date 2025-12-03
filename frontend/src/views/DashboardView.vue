<script setup>
import { ref, onMounted, computed } from 'vue'
import api from '@/api/client'
import { format } from 'date-fns'
import { ar } from 'date-fns/locale'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const loading = ref(true)
const error = ref(null)
const dashboardData = ref(null)
const widgetConfig = ref([])

// Transaction Types Mapping (Arabic)
const transactionTypes = {
  'income_grant': 'منحة',
  'income_donation': 'تبرع',
  'disbursement_custody': 'عهدة',
  'disbursement_salary': 'مرتبات',
  'disbursement_operational': 'تشغيلية',
  'disbursement_donation': 'تبرع صادر',
  'disbursement_bonus': 'مكافأة',
  'procurement': 'مشتريات',
  'transfer': 'تحويل'
}

// Widget key mapping
const widgetKeys = {
  FINANCIAL_SUMMARY: 'financial_summary',
  BANK_ACCOUNTS: 'banks',
  CASH_ACCOUNTS: 'cash_accounts',
  PROJECTS: 'projects',
  RECENT_INCOME: 'recent_income',
  RECENT_EXPENSES: 'recent_expenses'
}

onMounted(async () => {
  try {
    loading.value = true
    
    // Fetch both dashboard data and widget configuration
    const [dashboardResponse, widgetResponse] = await Promise.all([
      api.get('/reports/dashboard'),
      api.get('/admin/dashboard-widgets').catch(() => ({ data: [] })) // Gracefully handle if not admin
    ])
    
    dashboardData.value = dashboardResponse.data
    widgetConfig.value = widgetResponse.data.sort((a, b) => a.display_order - b.display_order)
  } catch (err) {
    console.error('Failed to load dashboard data:', err)
    error.value = 'فشل تحميل بيانات لوحة التحكم. يرجى المحاولة مرة أخرى.'
  } finally {
    loading.value = false
  }
})

// Check if a widget should be displayed
const isWidgetVisible = (widgetKey) => {
  if (!widgetConfig.value || widgetConfig.value.length === 0) {
    return true // Show all widgets if config not loaded
  }
  
  const widget = widgetConfig.value.find(w => w.widget_key === widgetKey)
  if (!widget) return true // Show if not configured
  
  const userRole = authStore.user?.role
  return widget.is_enabled && widget.allowed_roles.includes(userRole)
}

// Get ordered visible widgets
const orderedWidgets = computed(() => {
  if (!widgetConfig.value || widgetConfig.value.length === 0) {
    // Default order if no config
    return [
      widgetKeys.FINANCIAL_SUMMARY,
      widgetKeys.BANK_ACCOUNTS,
      widgetKeys.CASH_ACCOUNTS,
      widgetKeys.PROJECTS,
      widgetKeys.RECENT_TRANSACTIONS
    ]
  }
  
  const userRole = authStore.user?.role
  return widgetConfig.value
    .filter(w => w.is_enabled && w.allowed_roles.includes(userRole))
    .sort((a, b) => a.display_order - b.display_order)
    .map(w => w.widget_key)
})

const formatCurrency = (amount, currency) => {
  console.log(`Formatting: ${amount} | Currency: '${currency}'`)
  const code = (currency || 'LYD').trim().toUpperCase()
  let locale = 'ar-LY'
  
  // Use specific locales for foreign currencies to ensure correct symbols
  if (code === 'USD') locale = 'en-US'
  if (code === 'EUR') locale = 'en-IE'
  if (code === 'GBP') locale = 'en-GB'
  
  try {
    return new Intl.NumberFormat(locale, {
      style: 'currency',
      currency: code
    }).format(amount)
  } catch (e) {
    console.error('Currency formatting error:', e)
    return `${amount} ${code}`
  }
}

const formatDate = (dateString) => {
  if (!dateString) return '-'
  return format(new Date(dateString), 'dd MMMM yyyy', { locale: ar })
}

const getTransactionLabel = (type) => {
  return transactionTypes[type] || type
}
</script>

<template>
  <div class="p-6">
    <h1 class="text-3xl font-bold mb-8 text-gray-800">لوحة التحكم</h1>

    <!-- Loading State -->
    <div v-if="loading" class="flex justify-center items-center h-64">
      <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600"></div>
      <span class="mr-3 text-gray-600">جاري تحميل البيانات...</span>
    </div>

    <!-- Error State -->
    <div v-else-if="error" class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded relative" role="alert">
      <strong class="font-bold">خطأ!</strong>
      <span class="block sm:inline mr-2">{{ error }}</span>
    </div>

    <!-- Dashboard Content -->
    <div v-else-if="dashboardData" class="space-y-8">
      
      <!-- Render widgets in configured order -->
      <template v-for="widgetKey in orderedWidgets" :key="widgetKey">
        
        <!-- Financial Summary Section -->
        <section v-if="widgetKey === widgetKeys.FINANCIAL_SUMMARY && dashboardData.summary && (dashboardData.summary.total_balance.length > 0 || dashboardData.summary.total_income.length > 0 || dashboardData.summary.total_expenses.length > 0 || dashboardData.summary.total_custody.length > 0)" class="bg-gradient-to-r from-primary-50 to-primary-100 rounded-lg shadow-lg p-6 border border-primary-200">
          <h2 class="text-2xl font-bold mb-6 text-gray-800 flex items-center">
            <span class="ml-2">📊</span> الملخص المالي
          </h2>
          <div class="grid grid-cols-1 md:grid-cols-4 gap-6">
            <!-- Total Balance -->
            <div class="bg-white rounded-lg p-4 shadow">
              <h3 class="text-sm font-semibold text-gray-600 mb-3">إجمالي الرصيد</h3>
              <div v-if="dashboardData.summary.total_balance && dashboardData.summary.total_balance.length > 0" class="space-y-2">
                <div v-for="amount in dashboardData.summary.total_balance" :key="amount.currency" class="text-lg font-bold text-green-600">
                  {{ formatCurrency(amount.amount, amount.currency) }}
                </div>
              </div>
              <div v-else class="text-gray-400">لا توجد بيانات</div>
            </div>

            <!-- Total Income -->
            <div class="bg-white rounded-lg p-4 shadow">
              <h3 class="text-sm font-semibold text-gray-600 mb-3">إجمالي المقبوضات</h3>
              <div v-if="dashboardData.summary.total_income && dashboardData.summary.total_income.length > 0" class="space-y-2">
                <div v-for="amount in dashboardData.summary.total_income" :key="amount.currency" class="text-lg font-bold text-blue-600">
                  {{ formatCurrency(amount.amount, amount.currency) }}
                </div>
              </div>
              <div v-else class="text-gray-400">لا توجد بيانات</div>
            </div>

            <!-- Total Expenses -->
            <div class="bg-white rounded-lg p-4 shadow">
              <h3 class="text-sm font-semibold text-gray-600 mb-3">إجمالي المصروفات</h3>
              <div v-if="dashboardData.summary.total_expenses && dashboardData.summary.total_expenses.length > 0" class="space-y-2">
                <div v-for="amount in dashboardData.summary.total_expenses" :key="amount.currency" class="text-lg font-bold text-orange-600">
                  {{ formatCurrency(amount.amount, amount.currency) }}
                </div>
              </div>
              <div v-else class="text-gray-400">لا توجد بيانات</div>
            </div>

            <!-- Total Custody -->
            <div class="bg-white rounded-lg p-4 shadow">
              <h3 class="text-sm font-semibold text-gray-600 mb-3">إجمالي العهدة</h3>
              <div v-if="dashboardData.summary.total_custody && dashboardData.summary.total_custody.length > 0" class="space-y-2">
                <div v-for="amount in dashboardData.summary.total_custody" :key="amount.currency" class="text-lg font-bold text-purple-600">
                  {{ formatCurrency(amount.amount, amount.currency) }}
                </div>
              </div>
              <div v-else class="text-gray-400">لا توجد بيانات</div>
            </div>
          </div>
        </section>
        
        <!-- Bank Accounts Section -->
        <section v-if="widgetKey === widgetKeys.BANK_ACCOUNTS && dashboardData.bank_accounts && dashboardData.bank_accounts.length > 0">
          <h2 class="text-xl font-bold mb-4 text-gray-700 flex items-center">
            <span class="ml-2">🏦</span> الموقف المصرفي (آخر الحركات)
          </h2>
          <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div v-for="bank in dashboardData.bank_accounts" :key="bank.container.id" class="bg-white rounded-lg shadow-md p-6 border border-gray-100">
              <div class="flex justify-between items-start mb-4 border-b pb-4">
                <div>
                  <h3 class="text-lg font-bold text-gray-900">{{ bank.container.bank_name_ar || bank.container.bank_name }}</h3>
                  <p class="text-sm text-gray-500">{{ bank.container.account_number }}</p>
                </div>
                <div class="text-left">
                  <p class="text-sm text-gray-500">الرصيد الحالي</p>
                  <p class="text-xl font-bold" :class="bank.container.balance >= 0 ? 'text-green-600' : 'text-red-600'">
                    {{ formatCurrency(bank.container.balance, bank.container.currency) }}
                  </p>
                </div>
              </div>
              
              <!-- Bank Transactions -->
              <div class="overflow-x-auto">
                <table class="min-w-full text-sm">
                  <thead>
                    <tr class="bg-gray-50 text-gray-500">
                      <th class="px-3 py-2 text-right">التاريخ</th>
                      <th class="px-3 py-2 text-right">البيان</th>
                      <th class="px-3 py-2 text-right">القيمة</th>
                    </tr>
                  </thead>
                  <tbody class="divide-y divide-gray-100">
                    <tr v-for="tx in bank.last_transactions" :key="tx.id">
                      <td class="px-3 py-2 whitespace-nowrap">{{ formatDate(tx.transaction_date) }}</td>
                      <td class="px-3 py-2">{{ tx.description_ar || tx.description }}</td>
                      <td class="px-3 py-2 font-medium" :class="['income_grant', 'income_donation'].includes(tx.type) ? 'text-green-600' : 'text-red-600'">
                        {{ formatCurrency(tx.amount, tx.currency) }}
                      </td>
                    </tr>
                    <tr v-if="!bank.last_transactions || bank.last_transactions.length === 0">
                      <td colspan="3" class="px-3 py-4 text-center text-gray-400">لا توجد حركات حديثة</td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        </section>

        <!-- Cash Accounts Section -->
        <section v-if="widgetKey === widgetKeys.CASH_ACCOUNTS && dashboardData.cash_accounts && dashboardData.cash_accounts.length > 0">
          <h2 class="text-xl font-bold mb-4 text-gray-700 flex items-center">
            <span class="ml-2">💵</span> الموقف النقدي (آخر الحركات)
          </h2>
          <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div v-for="cash in dashboardData.cash_accounts" :key="cash.container.id" class="bg-white rounded-lg shadow-md p-6 border border-gray-100">
              <div class="flex justify-between items-start mb-4 border-b pb-4">
                <div>
                  <h3 class="text-lg font-bold text-gray-900">{{ cash.container.name_ar || cash.container.name }}</h3>
                  <p class="text-sm text-gray-500">{{ cash.container.location_ar || cash.container.location || 'صندوق نقدي' }}</p>
                </div>
                <div class="text-left">
                  <p class="text-sm text-gray-500">الرصيد الحالي</p>
                  <p class="text-xl font-bold" :class="cash.container.balance >= 0 ? 'text-green-600' : 'text-red-600'">
                    {{ formatCurrency(cash.container.balance, cash.container.currency) }}
                  </p>
                </div>
              </div>
              
              <!-- Cash Transactions -->
              <div class="overflow-x-auto">
                <table class="min-w-full text-sm">
                  <thead>
                    <tr class="bg-gray-50 text-gray-500">
                      <th class="px-3 py-2 text-right">التاريخ</th>
                      <th class="px-3 py-2 text-right">البيان</th>
                      <th class="px-3 py-2 text-right">القيمة</th>
                    </tr>
                  </thead>
                  <tbody class="divide-y divide-gray-100">
                    <tr v-for="tx in cash.last_transactions" :key="tx.id">
                      <td class="px-3 py-2 whitespace-nowrap">{{ formatDate(tx.transaction_date) }}</td>
                      <td class="px-3 py-2">{{ tx.description_ar || tx.description }}</td>
                      <td class="px-3 py-2 font-medium" :class="['income_grant', 'income_donation'].includes(tx.type) ? 'text-green-600' : 'text-red-600'">
                        {{ formatCurrency(tx.amount, tx.currency) }}
                      </td>
                    </tr>
                    <tr v-if="!cash.last_transactions || cash.last_transactions.length === 0">
                      <td colspan="3" class="px-3 py-4 text-center text-gray-400">لا توجد حركات حديثة</td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        </section>

        <!-- Projects Section -->
        <section v-if="widgetKey === widgetKeys.PROJECTS && dashboardData.projects && dashboardData.projects.length > 0">
          <h2 class="text-xl font-bold mb-4 text-gray-700 flex items-center">
            <span class="ml-2">📁</span> المشاريع (آخر الحركات)
          </h2>
          <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div v-for="project in dashboardData.projects" :key="project.project.id" class="bg-white rounded-lg shadow-md p-6 border border-gray-100">
              <div class="flex justify-between items-start mb-4 border-b pb-4">
                <div>
                  <h3 class="text-lg font-bold text-gray-900">{{ project.project.name_ar || project.project.name }}</h3>
                  <p class="text-sm text-gray-500">{{ project.project.code }}</p>
                </div>
                <div class="text-left">
                  <p class="text-sm text-gray-500">الميزانية</p>
                  <p class="text-lg font-bold text-gray-700">
                    {{ project.project.budget ? formatCurrency(project.project.budget, project.project.currency) : '-' }}
                  </p>
                </div>
              </div>

              <!-- Project Transactions -->
              <div class="overflow-x-auto">
                <table class="min-w-full text-sm">
                  <thead>
                    <tr class="bg-gray-50 text-gray-500">
                      <th class="px-3 py-2 text-right">التاريخ</th>
                      <th class="px-3 py-2 text-right">البيان</th>
                      <th class="px-3 py-2 text-right">القيمة</th>
                    </tr>
                  </thead>
                  <tbody class="divide-y divide-gray-100">
                    <tr v-for="tx in project.last_transactions" :key="tx.id">
                      <td class="px-3 py-2 whitespace-nowrap">{{ formatDate(tx.transaction_date) }}</td>
                      <td class="px-3 py-2">{{ tx.description_ar || tx.description }}</td>
                      <td class="px-3 py-2 font-medium" :class="['income_grant', 'income_donation'].includes(tx.type) ? 'text-green-600' : 'text-red-600'">
                        {{ formatCurrency(tx.amount, tx.currency) }}
                      </td>
                    </tr>
                    <tr v-if="!project.last_transactions || project.last_transactions.length === 0">
                      <td colspan="3" class="px-3 py-4 text-center text-gray-400">لا توجد حركات حديثة</td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        </section>

        <!-- Recent Income & Expenses (Combined Display) -->
        <!-- Only show once when we encounter either widget, and skip the second one -->
        <section 
          v-if="(widgetKey === widgetKeys.RECENT_INCOME && isWidgetVisible(widgetKeys.RECENT_INCOME)) || 
                (widgetKey === widgetKeys.RECENT_EXPENSES && isWidgetVisible(widgetKeys.RECENT_EXPENSES) && !isWidgetVisible(widgetKeys.RECENT_INCOME))" 
          class="grid grid-cols-1 lg:grid-cols-2 gap-6"
        >
          
          <!-- Recent Income -->
          <div v-if="isWidgetVisible(widgetKeys.RECENT_INCOME)" class="bg-white rounded-lg shadow-md p-6 border border-gray-100">
            <h2 class="text-lg font-bold mb-4 text-green-700 flex items-center">
              <span class="ml-2">📥</span> آخر المقبوضات
            </h2>
            <div class="overflow-x-auto">
              <table class="min-w-full text-sm">
                <thead>
                  <tr class="bg-gray-50 text-gray-500">
                    <th class="px-3 py-2 text-right">التاريخ</th>
                    <th class="px-3 py-2 text-right">البيان</th>
                    <th class="px-3 py-2 text-right">القيمة</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-gray-100">
                  <tr v-for="tx in dashboardData.recent_income" :key="tx.id" class="hover:bg-gray-50">
                    <td class="px-3 py-2 whitespace-nowrap text-gray-600">{{ formatDate(tx.transaction_date) }}</td>
                    <td class="px-3 py-2">
                      <div class="font-medium text-gray-900">{{ tx.description_ar || tx.description }}</div>
                      <div class="text-xs text-gray-500">{{ tx.container_name }}</div>
                    </td>
                    <td class="px-3 py-2 font-bold text-green-600">
                      {{ formatCurrency(tx.amount, tx.currency) }}
                    </td>
                  </tr>
                  <tr v-if="!dashboardData.recent_income || dashboardData.recent_income.length === 0">
                    <td colspan="3" class="px-3 py-8 text-center text-gray-400">لا توجد مقبوضات حديثة</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <!-- Recent Expenses -->
          <div v-if="isWidgetVisible(widgetKeys.RECENT_EXPENSES)" class="bg-white rounded-lg shadow-md p-6 border border-gray-100">
            <h2 class="text-lg font-bold mb-4 text-red-700 flex items-center">
              <span class="ml-2">📤</span> آخر المصروفات
            </h2>
            <div class="overflow-x-auto">
              <table class="min-w-full text-sm">
                <thead>
                  <tr class="bg-gray-50 text-gray-500">
                    <th class="px-3 py-2 text-right">التاريخ</th>
                    <th class="px-3 py-2 text-right">البيان</th>
                    <th class="px-3 py-2 text-right">القيمة</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-gray-100">
                  <tr v-for="tx in dashboardData.recent_expenses" :key="tx.id" class="hover:bg-gray-50">
                    <td class="px-3 py-2 whitespace-nowrap text-gray-600">{{ formatDate(tx.transaction_date) }}</td>
                    <td class="px-3 py-2">
                      <div class="font-medium text-gray-900">{{ tx.description_ar || tx.description }}</div>
                      <div class="text-xs text-gray-500">{{ tx.container_name }}</div>
                    </td>
                    <td class="px-3 py-2 font-bold text-red-600">
                      {{ formatCurrency(tx.amount, tx.currency) }}
                    </td>
                  </tr>
                  <tr v-if="!dashboardData.recent_expenses || dashboardData.recent_expenses.length === 0">
                    <td colspan="3" class="px-3 py-8 text-center text-gray-400">لا توجد مصروفات حديثة</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

        </section>

      </template>

    </div>

    <!-- Empty State (if data loaded but empty) -->
    <div v-else class="text-center py-12 bg-white rounded-lg shadow">
      <p class="text-gray-500 text-lg">لا توجد بيانات لعرضها حالياً</p>
    </div>
  </div>
</template>
