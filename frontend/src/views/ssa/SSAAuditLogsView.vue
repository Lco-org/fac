<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ssaAuditLogs } from '@/api/ssa'

const router = useRouter()

const logs = ref([])
const loading = ref(true)
const error = ref(null)
const total = ref(0)
const limit = ref(50)
const offset = ref(0)

const fetchLogs = async () => {
  loading.value = true
  error.value = null
  try {
    const response = await ssaAuditLogs.list(limit.value, offset.value)
    logs.value = response.data.logs || []
    total.value = response.data.total || 0
  } catch (err) {
    error.value = err.message
    console.error('Failed to fetch audit logs:', err)
  } finally {
    loading.value = false
  }
}

const goBack = () => {
  router.push({ name: 'SSADashboard' })
}

const prevPage = () => {
  if (offset.value > 0) {
    offset.value -= limit.value
    fetchLogs()
  }
}

const nextPage = () => {
  if (offset.value + limit.value < total.value) {
    offset.value += limit.value
    fetchLogs()
  }
}

const formatDate = (dateStr) => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('ar-LY')
}

const translateAction = (action) => {
  const actions = {
    'ssa_login_success': 'تسجيل دخول ناجح',
    'ssa_login_complete': 'اكتمال تسجيل الدخول',
    'ssa_login_password_verified': 'تحقق كلمة المرور',
    'ssa_login_failed': 'فشل تسجيل الدخول',
    'create_organization': 'إنشاء منظمة',
    'suspend_organization': 'إيقاف منظمة',
    'activate_organization': 'تفعيل منظمة',
    'create_database': 'إنشاء قاعدة بيانات',
    'create_tenant_admin': 'إنشاء مشرف منظمة',
    'verify_domain': 'تحقق من النطاق',
  }
  return actions[action] || action
}

const getActionClass = (action) => {
  if (action.includes('success') || action.includes('complete') || action.includes('create') || action.includes('activate')) {
    return 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200'
  }
  if (action.includes('failed') || action.includes('suspend')) {
    return 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200'
  }
  return 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200'
}

onMounted(() => {
  fetchLogs()
})
</script>

<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-900 p-6" dir="rtl">
    <div class="max-w-6xl mx-auto">
      <!-- Header -->
      <div class="mb-8">
        <button 
          @click="goBack"
          class="flex items-center gap-2 text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300 mb-4 transition-colors"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18"/>
          </svg>
          العودة للوحة التحكم
        </button>
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white flex items-center gap-3">
          <span class="text-4xl">📋</span>
          سجل العمليات
        </h1>
        <p class="mt-2 text-gray-500 dark:text-gray-400">
          جميع العمليات التي تمت في لوحة المشرف الأعلى
        </p>
      </div>

      <!-- Error Alert -->
      <div v-if="error" class="mb-6 p-4 bg-red-50 dark:bg-red-900/30 border border-red-200 dark:border-red-800 rounded-xl">
        <p class="text-red-600 dark:text-red-400 flex items-center gap-2">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
          </svg>
          {{ error }}
        </p>
      </div>

      <!-- Main Content -->
      <div class="bg-white dark:bg-gray-800 rounded-2xl shadow-lg border border-gray-100 dark:border-gray-700 overflow-hidden">
        <!-- Header -->
        <div class="p-6 border-b border-gray-100 dark:border-gray-700 flex justify-between items-center">
          <h2 class="text-xl font-bold text-gray-900 dark:text-white">
            سجل الأحداث ({{ total }} عملية)
          </h2>
          <button 
            @click="fetchLogs" 
            class="p-2 bg-gray-100 dark:bg-gray-700 rounded-lg hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors"
            :class="{ 'animate-spin': loading }"
          >
            <svg class="w-5 h-5 text-gray-600 dark:text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/>
            </svg>
          </button>
        </div>

        <!-- Loading -->
        <div v-if="loading" class="p-12 flex items-center justify-center">
          <div class="flex flex-col items-center gap-4">
            <div class="w-12 h-12 border-4 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
            <p class="text-gray-500 dark:text-gray-400">جاري التحميل...</p>
          </div>
        </div>

        <!-- Empty State -->
        <div v-else-if="logs.length === 0" class="p-12 text-center">
          <div class="w-24 h-24 mx-auto mb-4 bg-gray-100 dark:bg-gray-700 rounded-full flex items-center justify-center">
            <span class="text-4xl">📝</span>
          </div>
          <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-2">لا توجد عمليات</h3>
          <p class="text-gray-500 dark:text-gray-400">سيظهر هنا سجل جميع العمليات</p>
        </div>

        <!-- Table -->
        <div v-else class="overflow-x-auto">
          <table class="w-full">
            <thead class="bg-gray-50 dark:bg-gray-900">
              <tr>
                <th class="px-6 py-4 text-right text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase">التاريخ</th>
                <th class="px-6 py-4 text-right text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase">المشرف</th>
                <th class="px-6 py-4 text-right text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase">الإجراء</th>
                <th class="px-6 py-4 text-right text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase">النوع</th>
                <th class="px-6 py-4 text-right text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase">IP</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-gray-700">
              <tr 
                v-for="log in logs" 
                :key="log.id"
                class="hover:bg-gray-50 dark:hover:bg-gray-700/50 transition-colors"
              >
                <td class="px-6 py-4 text-sm text-gray-500 dark:text-gray-400 whitespace-nowrap">
                  {{ formatDate(log.created_at) }}
                </td>
                <td class="px-6 py-4">
                  <span class="text-sm font-medium text-gray-900 dark:text-white">
                    {{ log.admin_email }}
                  </span>
                </td>
                <td class="px-6 py-4">
                  <span :class="['px-3 py-1 rounded-full text-xs font-medium', getActionClass(log.action)]">
                    {{ translateAction(log.action) }}
                  </span>
                </td>
                <td class="px-6 py-4 text-sm text-gray-500 dark:text-gray-400">
                  {{ log.resource_type }}
                </td>
                <td class="px-6 py-4 text-sm font-mono text-gray-500 dark:text-gray-400">
                  {{ log.ip_address || '-' }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Pagination -->
        <div v-if="logs.length > 0" class="p-4 border-t border-gray-100 dark:border-gray-700 flex justify-between items-center">
          <div class="text-sm text-gray-500 dark:text-gray-400">
            عرض {{ offset + 1 }} - {{ Math.min(offset + limit, total) }} من {{ total }}
          </div>
          <div class="flex gap-2">
            <button 
              @click="prevPage" 
              :disabled="offset === 0"
              class="px-4 py-2 bg-gray-100 dark:bg-gray-700 rounded-lg disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors"
            >
              السابق
            </button>
            <button 
              @click="nextPage" 
              :disabled="offset + limit >= total"
              class="px-4 py-2 bg-gray-100 dark:bg-gray-700 rounded-lg disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors"
            >
              التالي
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
