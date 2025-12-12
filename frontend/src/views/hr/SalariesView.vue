<script setup>
import { ref, onMounted, watch, computed } from 'vue'
import api from '@/api/client'

const loading = ref(false)
const summaries = ref([])
const month = ref(new Date().toISOString().slice(0, 7)) // YYYY-MM
const error = ref(null)

const fetchSalaries = async () => {
  loading.value = true
  error.value = null
  try {
    const res = await api.get(`/hr/salaries?month=${month.value}`)
    summaries.value = res.data || []
  } catch (err) {
    console.error('Failed to load salaries', err)
    error.value = 'فشل في تحميل بيانات الرواتب'
  } finally {
    loading.value = false
  }
}

const groupedSummaries = computed(() => {
  const groups = {}
  summaries.value.forEach(s => {
    const currency = s.currency || 'LYD'
    if (!groups[currency]) {
      groups[currency] = []
    }
    groups[currency].push(s)
  })
  return groups
})

watch(month, () => {
  fetchSalaries()
})

onMounted(() => {
  fetchSalaries()
})
</script>

<template>
  <div class="container mx-auto px-4 py-8">
    <div class="flex flex-col md:flex-row justify-between items-center mb-6 gap-4">
      <h1 class="text-2xl font-bold text-gray-800">كشف الرواتب الشهري</h1>
      
      <div class="flex items-center gap-2">
        <label class="text-gray-700 font-medium">الشهر:</label>
        <input 
          type="month" 
          v-model="month"
          class="form-input border-gray-300 rounded-lg shadow-sm focus:border-green-500 focus:ring-green-500"
        >
      </div>
    </div>

    <!-- Error Alert -->
    <div v-if="error" class="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4">
      {{ error }}
    </div>

    <div v-if="loading" class="text-center py-8 text-gray-500">
      جاري التحميل...
    </div>

    <div v-else-if="summaries.length === 0" class="text-center py-8 text-gray-500 bg-white rounded shadow">
      لا توجد بيانات لهذا الشهر
    </div>

    <!-- Grouped Tables -->
    <div v-else v-for="(group, currency) in groupedSummaries" :key="currency" class="mb-8">
      <h2 class="text-xl font-bold text-gray-700 mb-3 px-2 border-r-4 border-blue-500">
        عملة: <span class="text-blue-600">{{ currency }}</span>
      </h2>
      
      <div class="bg-white rounded-lg shadow overflow-hidden">
        <div class="overflow-x-auto">
          <table class="min-w-full divide-y divide-gray-200">
            <thead class="bg-gray-50">
              <tr>
                <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">رقم الموظف</th>
                <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">الموظف</th>
                <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">الراتب الأساسي</th>
                <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider text-green-600">(+) المكافآت</th>
                <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider text-red-600">(-) الخصومات</th>
                <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider font-bold">صافي الراتب</th>
              </tr>
            </thead>
            <tbody class="bg-white divide-y divide-gray-200">
              <tr v-for="s in group" :key="s.employee_id" class="hover:bg-gray-50">
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                  {{ s.employee_number }}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                  {{ s.employee_name }}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                  {{ s.base_salary.toLocaleString() }}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-green-600 font-medium">
                  {{ s.total_bonuses.toLocaleString() }}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-red-600 font-medium">
                  {{ s.total_deductions.toLocaleString() }}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm font-bold text-gray-900">
                  {{ s.net_salary.toLocaleString() }}
                </td>
              </tr>
            </tbody>
            <tfoot class="bg-gray-100 font-bold">
              <tr>
                <td colspan="2" class="px-6 py-3 text-center">الإجمالي ({{ currency }})</td>
                <td class="px-6 py-3">{{ group.reduce((sum, s) => sum + s.base_salary, 0).toLocaleString() }}</td>
                <td class="px-6 py-3 text-green-700">{{ group.reduce((sum, s) => sum + s.total_bonuses, 0).toLocaleString() }}</td>
                <td class="px-6 py-3 text-red-700">{{ group.reduce((sum, s) => sum + s.total_deductions, 0).toLocaleString() }}</td>
                <td class="px-6 py-3 text-blue-800">{{ group.reduce((sum, s) => sum + s.net_salary, 0).toLocaleString() }}</td>
              </tr>
            </tfoot>
          </table>
        </div>
      </div>
    </div>
  </div>
</template>
