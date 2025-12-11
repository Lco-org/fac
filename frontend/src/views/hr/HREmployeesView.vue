<script setup>
import { ref, onMounted } from 'vue'
import api from '@/api/client'

const employees = ref([])
const loading = ref(false)
const error = ref(null)

const fetchEmployees = async () => {
  loading.value = true
  try {
    const res = await api.get('/employees')
    employees.value = res.data
  } catch (err) {
    error.value = 'Failed to load employees'
    console.error(err)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchEmployees()
})
</script>

<template>
  <div class="container mx-auto px-4 py-8">
    <h1 class="text-2xl font-bold text-gray-800 mb-6">سجلات الموظفين (الموارد البشرية)</h1>

    <div v-if="error" class="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4">
      {{ error }}
    </div>

    <div class="bg-white rounded-lg shadow overflow-hidden">
      <table class="min-w-full divide-y divide-gray-200">
        <thead class="bg-gray-50">
          <tr>
            <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">الإسم</th>
            <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">القسم</th>
            <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">المسمى الوظيفي</th>
            <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">الراتب الأساسي</th>
            <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">تاريخ التعيين</th>
            <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">الحالة</th>
          </tr>
        </thead>
        <tbody class="bg-white divide-y divide-gray-200">
          <tr v-if="loading">
            <td colspan="6" class="px-6 py-4 text-center text-gray-500">جاري التحميل...</td>
          </tr>
          <tr v-else-if="employees.length === 0">
            <td colspan="6" class="px-6 py-4 text-center text-gray-500">لا يوجد موظفين</td>
          </tr>
          <tr v-for="emp in employees" :key="emp.id" class="hover:bg-gray-50">
            <td class="px-6 py-4 whitespace-nowrap">
              <div class="text-sm font-medium text-gray-900">{{ emp.full_name }}</div>
              <div class="text-sm text-gray-500">{{ emp.phone }}</div>
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
              {{ emp.department || '-' }}
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
              {{ emp.position || '-' }}
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-sm font-bold text-gray-900">
              {{ emp.base_salary }} {{ emp.salary_currency }}
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
              {{ emp.join_date ? new Date(emp.join_date).toLocaleDateString('ar-LY') : '-' }}
            </td>
            <td class="px-6 py-4 whitespace-nowrap">
              <span 
                class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full"
                :class="emp.is_active ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'"
              >
                {{ emp.is_active ? 'نشط' : 'غير نشط' }}
              </span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
