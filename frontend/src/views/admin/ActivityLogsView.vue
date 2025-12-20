<template>
  <div class="p-6">
    <div class="flex justify-between items-center mb-6">
      <h1 class="text-2xl font-bold text-gray-800">سجل النشاطات</h1>
      <div class="flex gap-4">
        <select v-model="filterType" @change="fetchActivities" class="border rounded px-3 py-2">
          <option value="">كل الأنواع</option>
          <option value="transaction">المعاملات</option>
          <option value="user">المستخدمين</option>
          <option value="project">المشاريع</option>
          <option value="member">الأعضاء</option>
        </select>
        <input 
          type="text" 
          v-model="searchQuery" 
          placeholder="بحث..." 
          class="border rounded px-3 py-2 w-64"
          @input="debouncedSearch"
        >
      </div>
    </div>

    <div v-if="loading" class="text-center py-10">
      <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mx-auto"></div>
    </div>

    <div v-else-if="activities.length === 0" class="text-center py-10 bg-white rounded shadow text-gray-500">
      لا توجد نشاطات
    </div>

    <div v-else class="bg-white rounded-lg shadow overflow-hidden">
      <table class="min-w-full divide-y divide-gray-200">
        <thead class="bg-gray-50">
          <tr>
            <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">التاريخ</th>
            <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">المستخدم</th>
            <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">الإجراء</th>
            <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">النوع</th>
            <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">التفاصيل</th>
            <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">IP</th>
          </tr>
        </thead>
        <tbody class="bg-white divide-y divide-gray-200">
          <tr v-for="activity in activities" :key="activity.id" class="hover:bg-gray-50">
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
              {{ formatDate(activity.created_at) }}
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
              {{ activity.username || 'غير معروف' }}
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-sm">
              <span 
                class="px-2 py-1 rounded-full text-xs font-medium"
                :class="getActionClass(activity.action)"
              >
                {{ translateAction(activity.action) }}
              </span>
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
              {{ translateType(activity.entity_type) }}
            </td>
            <td class="px-6 py-4 text-sm text-gray-500 max-w-xs truncate" :title="JSON.stringify(activity.details)">
              {{ formatDetails(activity.details) }}
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-400 font-mono">
              {{ activity.ip || '-' }}
            </td>
          </tr>
        </tbody>
      </table>

      <!-- Pagination -->
      <div class="bg-gray-50 px-6 py-3 flex justify-between items-center border-t">
        <div class="text-sm text-gray-500">
          عرض {{ activities.length }} من {{ total }} نتيجة
        </div>
        <div class="flex gap-2">
          <button 
            @click="prevPage" 
            :disabled="offset === 0"
            class="px-3 py-1 border rounded disabled:opacity-50"
          >
            السابق
          </button>
          <button 
            @click="nextPage" 
            :disabled="offset + limit >= total"
            class="px-3 py-1 border rounded disabled:opacity-50"
          >
            التالي
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import activitiesApi from '@/api/activities'

const activities = ref([])
const loading = ref(true)
const filterType = ref('')
const searchQuery = ref('')
const offset = ref(0)
const limit = ref(50)
const total = ref(0)

let debounceTimer = null

const fetchActivities = async () => {
  loading.value = true
  try {
    const params = {
      limit: limit.value,
      offset: offset.value,
    }
    if (filterType.value) {
      params.entity_type = filterType.value
    }
    
    const response = await activitiesApi.getByUser(null, limit.value, offset.value)
    activities.value = response.data.activities || []
    total.value = response.data.total || activities.value.length
  } catch (error) {
    console.error('Failed to load activities', error)
  } finally {
    loading.value = false
  }
}

const debouncedSearch = () => {
  clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    offset.value = 0
    fetchActivities()
  }, 300)
}

const prevPage = () => {
  if (offset.value > 0) {
    offset.value -= limit.value
    fetchActivities()
  }
}

const nextPage = () => {
  if (offset.value + limit.value < total.value) {
    offset.value += limit.value
    fetchActivities()
  }
}

const formatDate = (dateString) => {
  return new Date(dateString).toLocaleString('ar-LY')
}

const translateAction = (action) => {
  const actions = {
    'create': 'إنشاء',
    'update': 'تعديل',
    'delete': 'حذف',
    'create_transaction': 'إنشاء معاملة',
    'reverse_transaction': 'عكس معاملة',
    'create_user': 'إنشاء مستخدم',
    'update_user': 'تعديل مستخدم',
    'delete_user': 'حذف مستخدم',
  }
  return actions[action] || action
}

const translateType = (type) => {
  const types = {
    'transaction': 'معاملة',
    'user': 'مستخدم',
    'project': 'مشروع',
    'member': 'عضو',
    'employee': 'موظف',
    'container': 'حساب',
  }
  return types[type] || type
}

const getActionClass = (action) => {
  if (action.includes('create')) return 'bg-green-100 text-green-800'
  if (action.includes('delete')) return 'bg-red-100 text-red-800'
  if (action.includes('update')) return 'bg-blue-100 text-blue-800'
  return 'bg-gray-100 text-gray-800'
}

const formatDetails = (details) => {
  if (!details) return '-'
  if (typeof details === 'string') return details
  return JSON.stringify(details).substring(0, 50) + '...'
}

onMounted(() => {
  fetchActivities()
})
</script>
