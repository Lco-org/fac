<script setup>
import { ref, onMounted } from 'vue'
import api from '@/api/client'

const loading = ref(true)
const dailyPosition = ref(null)

onMounted(async () => {
  try {
    const response = await api.get('/reports/daily-position')
    dailyPosition.value = response.data
  } catch (error) {
    console.error('Failed to load daily position:', error)
  } finally {
    loading.value = false
  }
})

const formatCurrency = (amount, currency) => {
  return `${amount.toFixed(2)} ${currency}`
}

const getBalanceClass = (balance) => {
  if (balance < 0) return 'balance-negative'
  if (balance > 0) return 'balance-positive'
  return 'balance-zero'
}
</script>

<template>
  <div>
    <h1 class="text-3xl font-bold mb-6">الجرد اليومي الشامل</h1>

    <div v-if="loading" class="text-center py-12">
      <p class="text-gray-500">جاري التحميل...</p>
    </div>

    <div v-else-if="dailyPosition" class="space-y-6">
      <!-- Summary -->
      <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
        <div class="card">
          <h3 class="text-sm font-medium text-gray-500 mb-2">إجمالي النقدية</h3>
          <p :class="['text-3xl font-bold', getBalanceClass(dailyPosition.total_cash)]">
            {{ formatCurrency(dailyPosition.total_cash, 'LYD') }}
          </p>
        </div>

        <div class="card">
          <h3 class="text-sm font-medium text-gray-500 mb-2">إجمالي البنوك</h3>
          <p :class="['text-3xl font-bold', getBalanceClass(dailyPosition.total_bank)]">
            {{ formatCurrency(dailyPosition.total_bank, 'LYD') }}
          </p>
        </div>

        <div class="card">
          <h3 class="text-sm font-medium text-gray-500 mb-2">إجمالي العجز</h3>
          <p :class="['text-3xl font-bold', getBalanceClass(dailyPosition.total_deficit)]">
            {{ formatCurrency(dailyPosition.total_deficit, 'LYD') }}
          </p>
        </div>
      </div>

      <!-- Cash Containers -->
      <div class="card">
        <h2 class="text-xl font-bold mb-4">الموقف النقدي (الخزائن)</h2>
        <div class="overflow-x-auto">
          <table class="min-w-full divide-y divide-gray-200">
            <thead class="bg-gray-50">
              <tr>
                <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">الاسم</th>
                <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">العملة</th>
                <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">الرصيد</th>
                <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">الحالة</th>
              </tr>
            </thead>
            <tbody class="bg-white divide-y divide-gray-200">
              <tr v-for="container in dailyPosition.cash_containers" :key="container.id">
                <td class="px-6 py-4">{{ container.name_ar || container.name }}</td>
                <td class="px-6 py-4">{{ container.currency }}</td>
                <td class="px-6 py-4">
                  <span :class="getBalanceClass(container.balance)">
                    {{ formatCurrency(container.balance, container.currency) }}
                  </span>
                </td>
                <td class="px-6 py-4">
                  <span :class="[
                    'px-2 py-1 text-xs rounded-full',
                    container.balance_status === 'deficit' ? 'bg-deficit-light text-deficit' :
                    container.balance_status === 'surplus' ? 'bg-surplus-light text-surplus' :
                    'bg-gray-100 text-gray-600'
                  ]">
                    {{ container.balance_status === 'deficit' ? 'عجز' : container.balance_status === 'surplus' ? 'فائض' : 'صفر' }}
                  </span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Bank Containers -->
      <div class="card">
        <h2 class="text-xl font-bold mb-4">الموقف المصرفي</h2>
        <div class="overflow-x-auto">
          <table class="min-w-full divide-y divide-gray-200">
            <thead class="bg-gray-50">
              <tr>
                <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">البنك</th>
                <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">رقم الحساب</th>
                <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">العملة</th>
                <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">الرصيد</th>
                <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">الحالة</th>
              </tr>
            </thead>
            <tbody class="bg-white divide-y divide-gray-200">
              <tr v-for="container in dailyPosition.bank_containers" :key="container.id">
                <td class="px-6 py-4">{{ container.bank_name_ar || container.bank_name }}</td>
                <td class="px-6 py-4">{{ container.account_number }}</td>
                <td class="px-6 py-4">{{ container.currency }}</td>
                <td class="px-6 py-4">
                  <span :class="getBalanceClass(container.balance)">
                    {{ formatCurrency(container.balance, container.currency) }}
                  </span>
                </td>
                <td class="px-6 py-4">
                  <span :class="[
                    'px-2 py-1 text-xs rounded-full',
                    container.balance_status === 'deficit' ? 'bg-deficit-light text-deficit' :
                    container.balance_status === 'surplus' ? 'bg-surplus-light text-surplus' :
                    'bg-gray-100 text-gray-600'
                  ]">
                    {{ container.balance_status === 'deficit' ? 'عجز' : container.balance_status === 'surplus' ? 'فائض' : 'صفر' }}
                  </span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </div>
</template>
