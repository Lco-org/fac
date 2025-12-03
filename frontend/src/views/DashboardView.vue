<script setup>
import { ref, onMounted } from 'vue'
import api from '@/api/client'
import { format } from 'date-fns'
import { ar } from 'date-fns/locale'

const loading = ref(true)
const error = ref(null)
const dashboardData = ref(null)

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

onMounted(async () => {
  try {
    loading.value = true
    const response = await api.get('/reports/dashboard')
    dashboardData.value = response.data
  } catch (err) {
    console.error('Failed to load dashboard data:', err)
    error.value = 'فشل تحميل بيانات لوحة التحكم. يرجى المحاولة مرة أخرى.'
  } finally {
    loading.value = false
  }
})

const formatCurrency = (amount, currency) => {
  return new Intl.NumberFormat('ar-LY', {
    style: 'currency',
    currency: currency || 'LYD'
  }).format(amount)
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
      
      <!-- 1. Bank Accounts Section -->
      <section v-if="dashboardData.bank_accounts && dashboardData.bank_accounts.length > 0">
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

      <!-- 2. Projects Section -->
      <section v-if="dashboardData.projects && dashboardData.projects.length > 0">
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

      <!-- 3. Recent Income & Expenses -->
      <section class="grid grid-cols-1 lg:grid-cols-2 gap-6">
        
        <!-- Recent Income -->
        <div class="bg-white rounded-lg shadow-md p-6 border border-gray-100">
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
        <div class="bg-white rounded-lg shadow-md p-6 border border-gray-100">
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

    </div>

    <!-- Empty State (if data loaded but empty) -->
    <div v-else class="text-center py-12 bg-white rounded-lg shadow">
      <p class="text-gray-500 text-lg">لا توجد بيانات لعرضها حالياً</p>
    </div>
  </div>
</template>
