<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()

// State
const loading = ref(true)
const error = ref(null)
const stats = ref({
  total_organizations: 0,
  active_organizations: 0,
  pending_verification: 0,
  suspended_organizations: 0,
  total_users: 0
})
const organizations = ref([])
const statusFilter = ref('')
const searchQuery = ref('')

// API Base URL
const API_BASE = '/api/v1/ssa'

// Computed filtered organizations
const filteredOrganizations = computed(() => {
  let filtered = organizations.value
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    filtered = filtered.filter(org => 
      org.name.toLowerCase().includes(query) || 
      org.primary_domain.toLowerCase().includes(query)
    )
  }
  return filtered
})

// Fetch dashboard data
const fetchDashboard = async () => {
  loading.value = true
  error.value = null
  const token = localStorage.getItem('ssa_token')
  const headers = {
    'Authorization': `Bearer ${token}`
  }
  try {
    // Get stats
    const statsRes = await fetch(`${API_BASE}/stats`, { 
      credentials: 'include',
      headers 
    })
    if (!statsRes.ok) throw new Error('Failed to fetch stats')
    stats.value = await statsRes.json()

    // Get organizations
    const orgsRes = await fetch(`${API_BASE}/organizations?limit=100`, { 
      credentials: 'include',
      headers 
    })
    if (!orgsRes.ok) throw new Error('Failed to fetch organizations')
    const data = await orgsRes.json()
    organizations.value = data.organizations || []
  } catch (err) {
    error.value = err.message
    console.error('SSA Dashboard Error:', err)
  } finally {
    loading.value = false
  }
}

// Navigate to create organization
const goToCreate = () => {
  router.push({ name: 'SSACreateOrganization' })
}

// View organization details
const viewOrganization = (orgId) => {
  router.push({ name: 'SSAOrganizationDetail', params: { id: orgId } })
}

// Status badge color
const getStatusColor = (status) => {
  const colors = {
    'active': 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200',
    'pending_verification': 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200',
    'suspended': 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200',
    'deleted': 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300'
  }
  return colors[status] || colors.deleted
}

// Status label
const getStatusLabel = (status) => {
  const labels = {
    'active': 'نشط',
    'pending_verification': 'بانتظار التحقق',
    'suspended': 'موقوف',
    'deleted': 'محذوف'
  }
  return labels[status] || status
}

// Format date
const formatDate = (dateStr) => {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return date.toLocaleDateString('ar-LY', { 
    year: 'numeric', 
    month: 'short', 
    day: 'numeric' 
  })
}

onMounted(() => {
  fetchDashboard()
})
</script>

<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-900 p-6" dir="rtl">
    <!-- Header -->
    <div class="max-w-7xl mx-auto">
      <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-8">
        <div>
          <h1 class="text-3xl font-bold text-gray-900 dark:text-white flex items-center gap-3">
            <span class="text-4xl">🔐</span>
            لوحة تحكم النظام الأعلى
          </h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
            إدارة المنظمات والأنظمة المتعددة
          </p>
        </div>
        <button 
          @click="goToCreate"
          class="inline-flex items-center gap-2 px-6 py-3 bg-gradient-to-r from-emerald-500 to-green-600 text-white rounded-xl font-bold shadow-lg hover:shadow-xl transition-all transform hover:-translate-y-0.5"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"/>
          </svg>
          إنشاء منظمة جديدة
        </button>
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

      <!-- Stats Cards -->
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        <!-- Total Organizations -->
        <div class="bg-white dark:bg-gray-800 rounded-2xl shadow-lg p-6 border border-gray-100 dark:border-gray-700">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-sm font-medium text-gray-500 dark:text-gray-400">إجمالي المنظمات</p>
              <p class="text-3xl font-bold text-gray-900 dark:text-white mt-1">{{ stats.total_organizations }}</p>
            </div>
            <div class="w-14 h-14 bg-blue-100 dark:bg-blue-900/50 rounded-xl flex items-center justify-center">
              <span class="text-2xl">🏢</span>
            </div>
          </div>
        </div>

        <!-- Active Organizations -->
        <div class="bg-white dark:bg-gray-800 rounded-2xl shadow-lg p-6 border border-gray-100 dark:border-gray-700">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-sm font-medium text-gray-500 dark:text-gray-400">المنظمات النشطة</p>
              <p class="text-3xl font-bold text-green-600 dark:text-green-400 mt-1">{{ stats.active_organizations }}</p>
            </div>
            <div class="w-14 h-14 bg-green-100 dark:bg-green-900/50 rounded-xl flex items-center justify-center">
              <span class="text-2xl">✅</span>
            </div>
          </div>
        </div>

        <!-- Pending Verification -->
        <div class="bg-white dark:bg-gray-800 rounded-2xl shadow-lg p-6 border border-gray-100 dark:border-gray-700">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-sm font-medium text-gray-500 dark:text-gray-400">بانتظار التحقق</p>
              <p class="text-3xl font-bold text-yellow-600 dark:text-yellow-400 mt-1">{{ stats.pending_verification }}</p>
            </div>
            <div class="w-14 h-14 bg-yellow-100 dark:bg-yellow-900/50 rounded-xl flex items-center justify-center">
              <span class="text-2xl">⏳</span>
            </div>
          </div>
        </div>

        <!-- Suspended -->
        <div class="bg-white dark:bg-gray-800 rounded-2xl shadow-lg p-6 border border-gray-100 dark:border-gray-700">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-sm font-medium text-gray-500 dark:text-gray-400">المنظمات الموقوفة</p>
              <p class="text-3xl font-bold text-red-600 dark:text-red-400 mt-1">{{ stats.suspended_organizations }}</p>
            </div>
            <div class="w-14 h-14 bg-red-100 dark:bg-red-900/50 rounded-xl flex items-center justify-center">
              <span class="text-2xl">🚫</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Organizations Table -->
      <div class="bg-white dark:bg-gray-800 rounded-2xl shadow-lg border border-gray-100 dark:border-gray-700 overflow-hidden">
        <div class="p-6 border-b border-gray-100 dark:border-gray-700">
          <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
            <h2 class="text-xl font-bold text-gray-900 dark:text-white flex items-center gap-2">
              <span>📋</span>
              قائمة المنظمات
            </h2>
            <div class="flex items-center gap-3">
              <!-- Search -->
              <div class="relative">
                <input 
                  v-model="searchQuery"
                  type="text" 
                  placeholder="بحث..."
                  class="w-64 px-4 py-2 pr-10 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                />
                <svg class="w-5 h-5 absolute left-3 top-2.5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"/>
                </svg>
              </div>
              <!-- Refresh -->
              <button 
                @click="fetchDashboard" 
                class="p-2 bg-gray-100 dark:bg-gray-700 rounded-lg hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors"
                :class="{ 'animate-spin': loading }"
              >
                <svg class="w-5 h-5 text-gray-600 dark:text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/>
                </svg>
              </button>
            </div>
          </div>
        </div>

        <!-- Loading State -->
        <div v-if="loading" class="p-12 flex items-center justify-center">
          <div class="flex flex-col items-center gap-4">
            <div class="w-12 h-12 border-4 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
            <p class="text-gray-500 dark:text-gray-400">جاري التحميل...</p>
          </div>
        </div>

        <!-- Empty State -->
        <div v-else-if="filteredOrganizations.length === 0" class="p-12 text-center">
          <div class="w-24 h-24 mx-auto mb-4 bg-gray-100 dark:bg-gray-700 rounded-full flex items-center justify-center">
            <span class="text-4xl">🏢</span>
          </div>
          <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-2">لا توجد منظمات</h3>
          <p class="text-gray-500 dark:text-gray-400 mb-6">ابدأ بإنشاء أول منظمة</p>
          <button 
            @click="goToCreate"
            class="px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
          >
            إنشاء منظمة جديدة
          </button>
        </div>

        <!-- Table -->
        <div v-else class="overflow-x-auto">
          <table class="w-full">
            <thead class="bg-gray-50 dark:bg-gray-900">
              <tr>
                <th class="px-6 py-4 text-right text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">المنظمة</th>
                <th class="px-6 py-4 text-right text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">الدومين</th>
                <th class="px-6 py-4 text-right text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">الحالة</th>
                <th class="px-6 py-4 text-right text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">تاريخ الإنشاء</th>
                <th class="px-6 py-4 text-right text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">الإجراءات</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-gray-700">
              <tr 
                v-for="org in filteredOrganizations" 
                :key="org.id"
                class="hover:bg-gray-50 dark:hover:bg-gray-700/50 transition-colors"
              >
                <td class="px-6 py-4">
                  <div class="flex items-center gap-3">
                    <div class="w-10 h-10 bg-gradient-to-br from-blue-500 to-purple-600 rounded-lg flex items-center justify-center text-white font-bold text-lg">
                      {{ org.name.charAt(0) }}
                    </div>
                    <div>
                      <p class="font-semibold text-gray-900 dark:text-white">{{ org.name }}</p>
                      <p v-if="org.name_ar" class="text-sm text-gray-500 dark:text-gray-400">{{ org.name_ar }}</p>
                    </div>
                  </div>
                </td>
                <td class="px-6 py-4">
                  <code class="px-2 py-1 bg-gray-100 dark:bg-gray-700 rounded text-sm font-mono text-gray-700 dark:text-gray-300">
                    {{ org.primary_domain }}
                  </code>
                </td>
                <td class="px-6 py-4">
                  <span :class="['px-3 py-1 rounded-full text-xs font-medium', getStatusColor(org.status)]">
                    {{ getStatusLabel(org.status) }}
                  </span>
                </td>
                <td class="px-6 py-4 text-sm text-gray-500 dark:text-gray-400">
                  {{ formatDate(org.created_at) }}
                </td>
                <td class="px-6 py-4">
                  <button 
                    @click="viewOrganization(org.id)"
                    class="px-4 py-2 bg-blue-50 dark:bg-blue-900/30 text-blue-600 dark:text-blue-400 rounded-lg hover:bg-blue-100 dark:hover:bg-blue-900/50 transition-colors text-sm font-medium"
                  >
                    عرض التفاصيل
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Security Notice -->
      <div class="mt-8 p-4 bg-amber-50 dark:bg-amber-900/20 border border-amber-200 dark:border-amber-800 rounded-xl">
        <p class="text-amber-700 dark:text-amber-400 text-sm flex items-center gap-2">
          <span class="text-lg">🔒</span>
          <span>
            <strong>تذكير أمني:</strong> 
            هذه اللوحة متاحة فقط من عناوين IP المصرح بها مع التحقق الثنائي (2FA).
          </span>
        </p>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Custom animations */
@keyframes slideIn {
  from {
    opacity: 0;
    transform: translateY(-10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.animate-slideIn {
  animation: slideIn 0.3s ease-out;
}
</style>

