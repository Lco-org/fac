<script setup>
import { ref, onMounted } from 'vue'
import api from '@/api/client'
import { format } from 'date-fns'
import { ar } from 'date-fns/locale'

const loading = ref(true)
const dashboardData = ref(null)

onMounted(async () => {
  try {
    const response = await api.get('/reports/dashboard')
    dashboardData.value = response.data
  } catch (error) {
    console.error('Failed to load dashboard data:', error)
  } finally {
    loading.value = false
  }
})

const getBalanceClass = (balance) => {
  if (balance < 0) return 'text-red-600'
  if (balance > 0) return 'text-green-600'
  return 'text-gray-600'
}

const formatCurrency = (amount, currency) => {
  return `${amount.toFixed(2)} ${currency}`
}

const formatDate = (date) => {
  if (!date) return '-'
  return format(new Date(date), 'dd MMMM yyyy', { locale: ar })
}

const getTransactionTypeLabel = (type) => {
  const types = {
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
  return types[type] || type
}
</script>

<template>
  <div>
    <h1 class="text-3xl font-bold mb-6">لوحة التحكم</h1>

    <div v-if="loading" class="text-center py-12">
      <p class="text-gray-500">جاري التحميل...</p>
    </div>

    <div v-else-if="dashboardData" class="space-y-8">
      
      <!-- Bank Accounts Section -->
      <section>
        <h2 class="text-xl font-bold mb-4 text-gray-800 flex items-center gap-2">
          <span>🏦</span>
          الموقف المصرفي (آخر الحركات)
        </h2>
        <div class="grid grid-cols-1 gap-6">
          <div v-for="bank in dashboardData.bank_accounts" :key="bank.container.id" class="card">
            <div class="flex justify-between items-start mb-4 border-b pb-4">
              <div>
                <h3 class="text-lg font-bold text-gray-900">{{ bank.container.bank_name_ar || bank.container.bank_name }}</h3>
                <p class="text-sm text-gray-500">{{ bank.container.account_number }}</p>
              </div>
              <div class="text-left">
                <p class="text-sm text-gray-500">الرصيد الحالي</p>
                <p :class="['text-xl font-bold', getBalanceClass(bank.container.balance)]">
                  {{ formatCurrency(bank.container.balance, bank.container.currency) }}
                </p>
              </div>
            </div>

            <!-- Mini Ledger Table -->
            <div class="overflow-x-auto">
              <table class="min-w-full divide-y divide-gray-200 text-sm">
                <thead class="bg-gray-50">
                  <tr>
                    <th class="px-4 py-2 text-right font-medium text-gray-500">التاريخ</th>
                    <th class="px-4 py-2 text-right font-medium text-gray-500">البيان</th>
                    <th class="px-4 py-2 text-right font-medium text-gray-500">النوع</th>
                    <th class="px-4 py-2 text-right font-medium text-gray-500">القيمة</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-gray-200">
                  <tr v-for="tx in bank.last_transactions" :key="tx.id" class="hover:bg-gray-50">
                    <td class="px-4 py-2 whitespace-nowrap text-gray-600">{{ formatDate(tx.transaction_date) }}</td>
                    <td class="px-4 py-2 text-gray-900">{{ tx.description_ar || tx.description }}</td>
                    <td class="px-4 py-2 whitespace-nowrap text-gray-500">{{ getTransactionTypeLabel(tx.type) }}</td>
                    <td class="px-4 py-2 whitespace-nowrap font-medium" :class="['income_grant', 'income_donation'].includes(tx.type) ? 'text-green-600' : 'text-red-600'">
                      {{ formatCurrency(tx.amount, tx.currency) }}
                    </td>
                  </tr>
                  <tr v-if="!bank.last_transactions || bank.last_transactions.length === 0">
                    <td colspan="4" class="px-4 py-4 text-center text-gray-500">لا توجد حركات حديثة</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </section>

      <!-- Projects Section -->
      <section>
        <h2 class="text-xl font-bold mb-4 text-gray-800 flex items-center gap-2">
          <span>📁</span>
          المشاريع (آخر الحركات)
        </h2>
        <div class="grid grid-cols-1 gap-6">
          <div v-for="project in dashboardData.projects" :key="project.project.id" class="card">
            <div class="flex justify-between items-start mb-4 border-b pb-4">
              <div>
                <h3 class="text-lg font-bold text-gray-900">{{ project.project.name_ar || project.project.name }}</h3>
                <p class="text-sm text-gray-500">{{ project.project.code }}</p>
                <p v-if="project.project.donor_name_ar || project.project.donor_name" class="text-xs text-gray-400">
                  الممول: {{ project.project.donor_name_ar || project.project.donor_name }}
                </p>
              </div>
              <div class="text-left">
                <p class="text-sm text-gray-500">الميزانية</p>
                <p class="text-lg font-bold text-gray-700">
                  {{ project.project.budget ? formatCurrency(project.project.budget, project.project.currency) : '-' }}
                </p>
              </div>
            </div>

            <!-- Mini Ledger Table -->
            <div class="overflow-x-auto">
              <table class="min-w-full divide-y divide-gray-200 text-sm">
                <thead class="bg-gray-50">
                  <tr>
                    <th class="px-4 py-2 text-right font-medium text-gray-500">التاريخ</th>
                    <th class="px-4 py-2 text-right font-medium text-gray-500">البيان</th>
                    <th class="px-4 py-2 text-right font-medium text-gray-500">النوع</th>
                    <th class="px-4 py-2 text-right font-medium text-gray-500">القيمة</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-gray-200">
                  <tr v-for="tx in project.last_transactions" :key="tx.id" class="hover:bg-gray-50">
                    <td class="px-4 py-2 whitespace-nowrap text-gray-600">{{ formatDate(tx.transaction_date) }}</td>
                    <td class="px-4 py-2 text-gray-900">{{ tx.description_ar || tx.description }}</td>
                    <td class="px-4 py-2 whitespace-nowrap text-gray-500">{{ getTransactionTypeLabel(tx.type) }}</td>
                    <td class="px-4 py-2 whitespace-nowrap font-medium" :class="['income_grant', 'income_donation'].includes(tx.type) ? 'text-green-600' : 'text-red-600'">
                      {{ formatCurrency(tx.amount, tx.currency) }}
                    </td>
                  </tr>
                  <tr v-if="!project.last_transactions || project.last_transactions.length === 0">
                    <td colspan="4" class="px-4 py-4 text-center text-gray-500">لا توجد حركات حديثة</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </section>

      <!-- Recent Activity Section -->
      <section class="grid grid-cols-1 lg:grid-cols-2 gap-6">
        
        <!-- Recent Income -->
        <div class="card">
          <h2 class="text-lg font-bold mb-4 text-green-700 flex items-center gap-2">
            <span>📥</span>
            آخر المقبوضات
          </h2>
          <div class="overflow-x-auto">
            <table class="min-w-full divide-y divide-gray-200 text-sm">
              <thead class="bg-gray-50">
                <tr>
                  <th class="px-3 py-2 text-right font-medium text-gray-500">التاريخ</th>
                  <th class="px-3 py-2 text-right font-medium text-gray-500">البيان</th>
                  <th class="px-3 py-2 text-right font-medium text-gray-500">القيمة</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-200">
                <tr v-for="tx in dashboardData.recent_income" :key="tx.id" class="hover:bg-gray-50">
                  <td class="px-3 py-2 whitespace-nowrap text-gray-600">{{ formatDate(tx.transaction_date) }}</td>
                  <td class="px-3 py-2 text-gray-900">
                    <div class="font-medium">{{ tx.description_ar || tx.description }}</div>
                    <div class="text-xs text-gray-500">{{ tx.container_name }}</div>
                  </td>
                  <td class="px-3 py-2 whitespace-nowrap font-bold text-green-600">
                    {{ formatCurrency(tx.amount, tx.currency) }}
                  </td>
                </tr>
                <tr v-if="!dashboardData.recent_income || dashboardData.recent_income.length === 0">
                  <td colspan="3" class="px-3 py-4 text-center text-gray-500">لا توجد مقبوضات حديثة</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <!-- Recent Expenses -->
        <div class="card">
          <h2 class="text-lg font-bold mb-4 text-red-700 flex items-center gap-2">
            <span>📤</span>
            آخر المصروفات
          </h2>
          <div class="overflow-x-auto">
            <table class="min-w-full divide-y divide-gray-200 text-sm">
              <thead class="bg-gray-50">
                <tr>
                  <th class="px-3 py-2 text-right font-medium text-gray-500">التاريخ</th>
                  <th class="px-3 py-2 text-right font-medium text-gray-500">البيان</th>
                  <th class="px-3 py-2 text-right font-medium text-gray-500">القيمة</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-200">
                <tr v-for="tx in dashboardData.recent_expenses" :key="tx.id" class="hover:bg-gray-50">
                  <td class="px-3 py-2 whitespace-nowrap text-gray-600">{{ formatDate(tx.transaction_date) }}</td>
                  <td class="px-3 py-2 text-gray-900">
                    <div class="font-medium">{{ tx.description_ar || tx.description }}</div>
                    <div class="text-xs text-gray-500">{{ tx.container_name }}</div>
                  </td>
                  <td class="px-3 py-2 whitespace-nowrap font-bold text-red-600">
                    {{ formatCurrency(tx.amount, tx.currency) }}
                  </td>
                </tr>
                <tr v-if="!dashboardData.recent_expenses || dashboardData.recent_expenses.length === 0">
                  <td colspan="3" class="px-3 py-4 text-center text-gray-500">لا توجد مصروفات حديثة</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

      </section>
    </div>
  </div>
</template>
