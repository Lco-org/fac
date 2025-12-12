<script setup>
import { ref, onMounted, computed } from 'vue'
import api from '@/api/client'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const deductions = ref([])
const loading = ref(false)
const showModal = ref(false)
const employees = ref([])
const error = ref(null)

const form = ref({
  employee_id: '',
  month_year: new Date().toISOString().slice(0, 7), // YYYY-MM
  amount: '',
  currency: 'LYD',
  reason: '',
  reason_ar: ''
})

const fetchEmployees = async () => {
  try {
    const res = await api.get('/employees')
    employees.value = res.data
  } catch (err) {
    console.error('Failed to fetch employees', err)
  }
}

const fetchDeductions = async () => {
  loading.value = true
  try {
    const res = await api.get('/hr/deductions')
    deductions.value = res.data || []
  } catch (err) {
    error.value = 'Failed to load deductions'
    console.error(err)
  } finally {
    loading.value = false
  }
}

const groupedDeductions = computed(() => {
  const groups = {}
  deductions.value.forEach(d => {
    const currency = d.currency || 'LYD'
    if (!groups[currency]) {
      groups[currency] = []
    }
    groups[currency].push(d)
  })
  return groups
})

const handleSubmit = async () => {
  if (!form.value.employee_id || !form.value.amount || !form.value.reason) {
    alert('Please fill all required fields')
    return
  }

  try {
    const payload = {
      ...form.value,
      amount: parseFloat(form.value.amount),
      month_year: new Date(form.value.month_year + '-01').toISOString()
    }

    await api.post('/hr/deductions', payload)
    showModal.value = false
    await fetchDeductions()
    resetForm()
  } catch (err) {
    console.error('Failed to create deduction', err)
    const serverError = err.response?.data?.error || 'Failed to create deduction'
    alert(serverError)
  }
}

const resetForm = () => {
  form.value = {
    employee_id: '',
    month_year: new Date().toISOString().slice(0, 7),
    amount: '',
    currency: 'LYD',
    reason: '',
    reason_ar: ''
  }
}

const getEmployeeName = (id) => {
  const emp = employees.value.find(e => e.id === id)
  return emp ? emp.full_name : 'Unknown'
}

onMounted(() => {
  fetchEmployees()
  fetchDeductions()
})
</script>

<template>
  <div class="container mx-auto px-4 py-8">
    <div class="flex justify-between items-center mb-6">
      <h1 class="text-2xl font-bold text-gray-800">إدارة الخصومات والجزاءات</h1>
      <button 
        @click="showModal = true"
        class="bg-red-600 text-white px-4 py-2 rounded-lg hover:bg-red-700 transition-colors flex items-center gap-2"
      >
        <span>➕</span>
        <span>إضافة خصم جديد</span>
      </button>
    </div>

    <!-- Error Alert -->
    <div v-if="error" class="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4">
      {{ error }}
    </div>

    <div v-if="loading" class="text-center py-8 text-gray-500">
      جاري التحميل...
    </div>

    <div v-else-if="deductions.length === 0" class="text-center py-8 text-gray-500 bg-white rounded shadow">
      لا توجد خصومات مسجلة
    </div>

    <!-- Grouped Deductions Tables -->
    <div v-else v-for="(group, currency) in groupedDeductions" :key="currency" class="mb-8">
      <h2 class="text-xl font-bold text-gray-700 mb-3 px-2 border-r-4 border-red-500">
        عملة: <span class="text-red-600">{{ currency }}</span>
      </h2>

      <div class="bg-white rounded-lg shadow overflow-hidden">
        <table class="min-w-full divide-y divide-gray-200">
          <thead class="bg-gray-50">
            <tr>
              <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">الموظف</th>
              <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">الشهر</th>
              <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">المبلغ</th>
              <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">السبب</th>
              <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">تاريخ الإضافة</th>
            </tr>
          </thead>
          <tbody class="bg-white divide-y divide-gray-200">
            <tr v-for="d in group" :key="d.id" class="hover:bg-gray-50">
              <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                {{ getEmployeeName(d.employee_id) }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                {{ new Date(d.month_year).toLocaleDateString('en-GB', { month: 'long', year: 'numeric' }) }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm font-bold text-red-600">
                {{ d.amount }} {{ d.currency }}
              </td>
              <td class="px-6 py-4 text-sm text-gray-500">
                {{ d.reason }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                {{ new Date(d.created_at).toLocaleDateString('ar-LY') }}
              </td>
            </tr>
          </tbody>
           <tfoot class="bg-gray-100 font-bold">
              <tr>
                <td colspan="2" class="px-6 py-3 text-center">الإجمالي ({{ currency }})</td>
                <td class="px-6 py-3 text-red-700">{{ group.reduce((sum, d) => sum + d.amount, 0).toLocaleString() }}</td>
                <td colspan="2"></td>
              </tr>
            </tfoot>
        </table>
      </div>
    </div>

    <!-- Add Deduction Modal -->
    <Teleport to="body">
    <div v-if="showModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-[9999]">
      <div class="bg-white rounded-lg shadow-xl w-full max-w-md">
        <div class="flex justify-between items-center p-4 border-b">
          <h3 class="text-lg font-bold">إضافة خصم جديد</h3>
          <button @click="showModal = false" class="text-gray-500 hover:text-gray-700">✕</button>
        </div>
        
        <form @submit.prevent="handleSubmit" class="p-6 space-y-4">
          <!-- Employee -->
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">الموظف</label>
            <select v-model="form.employee_id" required class="form-input w-full border-gray-300 rounded-md shadow-sm">
              <option value="" disabled>اختر الموظف</option>
              <option v-for="emp in employees" :key="emp.id" :value="emp.id">
                {{ emp.full_name }}
              </option>
            </select>
          </div>

          <!-- Month -->
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">الشهر</label>
            <input type="month" v-model="form.month_year" required class="form-input w-full border-gray-300 rounded-md shadow-sm">
          </div>

          <!-- Amount -->
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">المبلغ</label>
            <div class="flex gap-2">
              <input type="number" step="0.01" v-model="form.amount" required class="form-input flex-1 border-gray-300 rounded-md shadow-sm">
              <select v-model="form.currency" class="form-select w-24 border-gray-300 rounded-md shadow-sm">
                <option value="LYD">LYD</option>
                <option value="USD">USD</option>
                <option value="EUR">EUR</option>
              </select>
            </div>
          </div>

          <!-- Reason -->
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">السبب</label>
            <textarea v-model="form.reason" rows="3" required class="form-textarea w-full border-gray-300 rounded-md shadow-sm"></textarea>
          </div>

          <div class="flex justify-end gap-3 mt-6">
            <button type="button" @click="showModal = false" class="btn-secondary">إلغاء</button>
            <button type="submit" class="bg-red-600 text-white px-4 py-2 rounded hover:bg-red-700">حفظ الخصم</button>
          </div>
        </form>
      </div>
    </div>
    </Teleport>
  </div>
</template>
