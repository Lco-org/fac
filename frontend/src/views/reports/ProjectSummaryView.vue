<script setup>
import { ref, onMounted } from 'vue'
import api from '@/api/client'

const loading = ref(true)
const projects = ref([])

onMounted(async () => {
  try {
    const response = await api.get('/reports/project-summary')
    projects.value = response.data
  } catch (error) {
    console.error('Failed to load projects:', error)
  } finally {
    loading.value = false
  }
})

const formatCurrency = (amount, currency) => {
  // Use Western numerals (123456) with comma separators
  const formatted = Number(amount).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
  return `${formatted} ${currency}`
}

const getBalanceClass = (balance) => {
  if (balance < 0) return 'balance-negative'
  if (balance > 0) return 'balance-positive'
  return 'balance-zero'
}
</script>

<template>
  <div>
    <h1 class="text-3xl font-bold mb-6">تقرير حركة المشاريع</h1>

    <div v-if="loading" class="text-center py-12">
      <p class="text-gray-500">جاري التحميل...</p>
    </div>

    <div v-else class="card">
      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200">
          <thead class="bg-gray-50">
            <tr>
              <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">الكود</th>
              <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">اسم المشروع</th>
              <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">المانح</th>
              <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">المقبوضات</th>
              <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">المصروفات</th>
              <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">صافي الموقف</th>
              <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">عدد العمليات</th>
            </tr>
          </thead>
          <tbody class="bg-white divide-y divide-gray-200">
            <tr v-for="project in projects" :key="project.id">
              <td class="px-6 py-4 whitespace-nowrap font-medium">{{ project.code }}</td>
              <td class="px-6 py-4 whitespace-nowrap">{{ project.name_ar || project.name }}</td>
              <td class="px-6 py-4 whitespace-nowrap">{{ project.donor_name || '-' }}</td>
              <td class="px-6 py-4 whitespace-nowrap text-surplus">
                {{ formatCurrency(project.total_income, project.currency) }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-deficit">
                {{ formatCurrency(project.total_expenses, project.currency) }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <span :class="getBalanceClass(project.net_position)">
                  {{ formatCurrency(project.net_position, project.currency) }}
                </span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-center">{{ project.transaction_count }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>
