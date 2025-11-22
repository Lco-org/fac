<script setup>
import { ref, onMounted } from 'vue'
import api from '@/api/client'

const loading = ref(true)
const deficitReport = ref(null)

onMounted(async () => {
  try {
    const response = await api.get('/reports/deficit')
    deficitReport.value = response.data
  } catch (error) {
    console.error('Failed to load deficit report:', error)
  } finally {
    loading.value = false
  }
})

const formatCurrency = (amount, currency) => {
  // Use Western numerals (123456) with comma separators
  const formatted = Number(amount).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
  return `${formatted} ${currency}`
}

const formatDate = (dateString) => {
  if (!dateString) return '-'
  const date = new Date(dateString)
  return date.toLocaleDateString('ar-LY')
}
</script>

<template>
  <div>
    <h1 class="text-3xl font-bold mb-6">تقرير الالتزامات والعجز</h1>

    <div v-if="loading" class="text-center py-12">
      <p class="text-gray-500">جاري التحميل...</p>
    </div>

    <div v-else-if="deficitReport" class="space-y-6">
      <!-- Summary Card -->
      <div class="card bg-deficit-light border-2 border-deficit">
        <div class="flex justify-between items-center">
          <div>
            <h3 class="text-sm font-medium text-deficit-dark mb-2">إجمالي العجز في النظام</h3>
            <p class="text-4xl font-bold text-deficit">
              {{ formatCurrency(deficitReport.total_deficit, 'LYD') }}
            </p>
          </div>
          <div class="text-center">
            <p class="text-sm text-deficit-dark">عدد الحسابات بالعجز</p>
            <p class="text-3xl font-bold text-deficit">{{ deficitReport.containers_count }}</p>
          </div>
        </div>
      </div>

      <!-- Warning Message -->
      <div class="card bg-yellow-50 border-2 border-yellow-400">
        <div class="flex items-start gap-3">
          <span class="text-3xl">⚠️</span>
          <div>
            <h3 class="font-bold text-yellow-800 mb-2">تنبيه هام</h3>
            <p class="text-sm text-yellow-700">
              هذه المبالغ تمثل التزامات (ديون داخلية) على المنظمة يجب سدادها فور توفر السيولة المالية.
              يرجى متابعة هذا التقرير بشكل دوري والعمل على تغطية العجز في أقرب وقت ممكن.
            </p>
          </div>
        </div>
      </div>

      <!-- Deficit Containers Table -->
      <div class="card">
        <h2 class="text-xl font-bold mb-4">تفاصيل الحسابات بالعجز</h2>
        
        <div v-if="deficitReport.containers.length === 0" class="text-center py-8">
          <p class="text-surplus text-lg">✅ لا يوجد عجز في النظام - جميع الحسابات بحالة جيدة</p>
        </div>

        <div v-else class="overflow-x-auto">
          <table class="min-w-full divide-y divide-gray-200">
            <thead class="bg-gray-50">
              <tr>
                <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">الاسم</th>
                <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">النوع</th>
                <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">العملة</th>
                <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">مبلغ العجز</th>
                <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">عدد العمليات</th>
                <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">أول عجز</th>
                <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">آخر عجز</th>
              </tr>
            </thead>
            <tbody class="bg-white divide-y divide-gray-200">
              <tr v-for="container in deficitReport.containers" :key="container.id" class="hover:bg-gray-50">
                <td class="px-6 py-4 whitespace-nowrap font-medium">
                  {{ container.name_ar || container.name }}
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                  <span :class="[
                    'px-2 py-1 text-xs rounded-full',
                    container.type === 'cash' ? 'bg-blue-100 text-blue-800' : 'bg-green-100 text-green-800'
                  ]">
                    {{ container.type === 'cash' ? 'نقدي' : 'بنك' }}
                  </span>
                </td>
                <td class="px-6 py-4 whitespace-nowrap">{{ container.currency }}</td>
                <td class="px-6 py-4 whitespace-nowrap">
                  <span class="text-deficit font-bold">
                    {{ formatCurrency(container.deficit_amount, container.currency) }}
                  </span>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-center">
                  {{ container.deficit_transactions_count }}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-600">
                  {{ formatDate(container.first_deficit_date) }}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-600">
                  {{ formatDate(container.last_deficit_date) }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </div>
</template>
