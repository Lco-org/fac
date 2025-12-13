<script setup>
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'

const router = useRouter()
const route = useRoute()

// State
const orgId = ref(route.params.id)
const loading = ref(true)
const error = ref(null)
const organization = ref(null)
const suspending = ref(false)
const activating = ref(false)

// API Base URL
const API_BASE = '/api/v1/ssa'

// Fetch organization details
const fetchOrganization = async () => {
  loading.value = true
  error.value = null
  try {
    const token = localStorage.getItem('ssa_token')
    const response = await fetch(`${API_BASE}/organizations/${orgId.value}`, { 
      credentials: 'include',
      headers: { 'Authorization': `Bearer ${token}` }
    })
    if (!response.ok) {
      throw new Error('فشل تحميل بيانات المنظمة')
    }
    organization.value = await response.json()
  } catch (err) {
    error.value = err.message
  } finally {
    loading.value = false
  }
}

// Suspend organization
const suspendOrg = async () => {
  const reason = prompt('أدخل سبب الإيقاف:')
  if (!reason) return
  
  suspending.value = true
  try {
    const token = localStorage.getItem('ssa_token')
    const response = await fetch(`${API_BASE}/organizations/${orgId.value}/suspend`, {
      method: 'POST',
      headers: { 
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      credentials: 'include',
      body: JSON.stringify({ reason })
    })
    if (!response.ok) throw new Error('فشل إيقاف المنظمة')
    await fetchOrganization()
  } catch (err) {
    alert(err.message)
  } finally {
    suspending.value = false
  }
}

// Activate organization
const activateOrg = async () => {
  activating.value = true
  try {
    const token = localStorage.getItem('ssa_token')
    const response = await fetch(`${API_BASE}/organizations/${orgId.value}/activate`, {
      method: 'POST',
      credentials: 'include',
      headers: { 'Authorization': `Bearer ${token}` }
    })
    if (!response.ok) throw new Error('فشل تفعيل المنظمة')
    await fetchOrganization()
  } catch (err) {
    alert(err.message)
  } finally {
    activating.value = false
  }
}

// Navigate functions
const goBack = () => router.push({ name: 'SSADashboard' })
const goToVerify = () => router.push({ name: 'SSAVerifyDomain', params: { id: orgId.value } })
const goToDatabase = () => router.push({ name: 'SSACreateDatabase', params: { id: orgId.value } })
const goToAdmin = () => router.push({ name: 'SSACreateAdmin', params: { id: orgId.value } })

// Status helpers
const getStatusColor = (status) => {
  const colors = {
    'active': 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200',
    'pending_verification': 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200',
    'suspended': 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200'
  }
  return colors[status] || 'bg-gray-100 text-gray-800'
}

const getStatusLabel = (status) => {
  const labels = {
    'active': 'نشط ✓',
    'pending_verification': 'بانتظار التحقق ⏳',
    'suspended': 'موقوف 🚫'
  }
  return labels[status] || status
}

const formatDate = (dateStr) => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleDateString('ar-LY', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

onMounted(() => {
  fetchOrganization()
})
</script>

<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-900 p-6" dir="rtl">
    <div class="max-w-4xl mx-auto">
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
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white">
          تفاصيل المنظمة
        </h1>
      </div>

      <!-- Loading -->
      <div v-if="loading" class="flex items-center justify-center p-12">
        <div class="w-12 h-12 border-4 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
      </div>

      <!-- Error -->
      <div v-else-if="error" class="bg-red-50 dark:bg-red-900/30 rounded-xl p-8 text-center">
        <p class="text-red-600 dark:text-red-400">{{ error }}</p>
        <button @click="fetchOrganization" class="mt-4 px-6 py-2 bg-red-600 text-white rounded-lg">إعادة المحاولة</button>
      </div>

      <!-- Content -->
      <div v-else-if="organization" class="space-y-6">
        <!-- Organization Header -->
        <div class="bg-white dark:bg-gray-800 rounded-2xl shadow-lg p-6 border border-gray-100 dark:border-gray-700">
          <div class="flex items-start justify-between">
            <div class="flex items-center gap-4">
              <div class="w-20 h-20 bg-gradient-to-br from-blue-500 to-purple-600 rounded-2xl flex items-center justify-center text-white font-bold text-3xl shadow-lg">
                {{ organization.name.charAt(0) }}
              </div>
              <div>
                <h2 class="text-2xl font-bold text-gray-900 dark:text-white">{{ organization.name }}</h2>
                <p v-if="organization.name_ar" class="text-gray-500 dark:text-gray-400">{{ organization.name_ar }}</p>
                <code class="mt-2 inline-block px-3 py-1 bg-gray-100 dark:bg-gray-700 rounded-lg text-sm font-mono text-gray-600 dark:text-gray-400">
                  {{ organization.primary_domain }}
                </code>
              </div>
            </div>
            <span :class="['px-4 py-2 rounded-full text-sm font-medium', getStatusColor(organization.status)]">
              {{ getStatusLabel(organization.status) }}
            </span>
          </div>
        </div>

        <!-- Quick Actions Based on Status -->
        <div v-if="organization.status === 'pending_verification'" class="bg-yellow-50 dark:bg-yellow-900/30 rounded-xl p-6 border border-yellow-200 dark:border-yellow-800">
          <h3 class="font-bold text-yellow-800 dark:text-yellow-200 mb-3">⏳ بانتظار التحقق من الدومين</h3>
          <p class="text-yellow-700 dark:text-yellow-300 text-sm mb-4">المنظمة بحاجة للتحقق من البريد الإلكتروني قبل المتابعة.</p>
          <button @click="goToVerify" class="px-6 py-3 bg-yellow-600 text-white rounded-lg font-medium hover:bg-yellow-700 transition-colors">
            متابعة التحقق
          </button>
        </div>

        <!-- Details Grid -->
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <!-- Contact Info -->
          <div class="bg-white dark:bg-gray-800 rounded-2xl shadow-lg p-6 border border-gray-100 dark:border-gray-700">
            <h3 class="text-lg font-bold text-gray-900 dark:text-white mb-4 flex items-center gap-2">
              <span>📧</span>
              معلومات التواصل
            </h3>
            <div class="space-y-4">
              <div class="flex justify-between items-center p-3 bg-gray-50 dark:bg-gray-700 rounded-lg">
                <span class="text-gray-600 dark:text-gray-400">البريد:</span>
                <span class="text-gray-900 dark:text-white font-mono">{{ organization.contact_email }}</span>
              </div>
              <div v-if="organization.contact_phone" class="flex justify-between items-center p-3 bg-gray-50 dark:bg-gray-700 rounded-lg">
                <span class="text-gray-600 dark:text-gray-400">الهاتف:</span>
                <span class="text-gray-900 dark:text-white">{{ organization.contact_phone }}</span>
              </div>
            </div>
          </div>

          <!-- Database Info -->
          <div class="bg-white dark:bg-gray-800 rounded-2xl shadow-lg p-6 border border-gray-100 dark:border-gray-700">
            <h3 class="text-lg font-bold text-gray-900 dark:text-white mb-4 flex items-center gap-2">
              <span>🗄️</span>
              قاعدة البيانات
            </h3>
            <div class="space-y-4">
              <div class="flex justify-between items-center p-3 bg-gray-50 dark:bg-gray-700 rounded-lg">
                <span class="text-gray-600 dark:text-gray-400">اسم قاعدة البيانات:</span>
                <code class="text-blue-600 dark:text-blue-400 font-mono">{{ organization.database_name }}</code>
              </div>
              <div class="flex justify-between items-center p-3 bg-gray-50 dark:bg-gray-700 rounded-lg">
                <span class="text-gray-600 dark:text-gray-400">Slug:</span>
                <code class="text-gray-600 dark:text-gray-400 font-mono">{{ organization.slug }}</code>
              </div>
            </div>
          </div>

          <!-- Timestamps -->
          <div class="bg-white dark:bg-gray-800 rounded-2xl shadow-lg p-6 border border-gray-100 dark:border-gray-700">
            <h3 class="text-lg font-bold text-gray-900 dark:text-white mb-4 flex items-center gap-2">
              <span>📅</span>
              التواريخ
            </h3>
            <div class="space-y-4">
              <div class="flex justify-between items-center p-3 bg-gray-50 dark:bg-gray-700 rounded-lg">
                <span class="text-gray-600 dark:text-gray-400">تاريخ الإنشاء:</span>
                <span class="text-gray-900 dark:text-white text-sm">{{ formatDate(organization.created_at) }}</span>
              </div>
              <div v-if="organization.activated_at" class="flex justify-between items-center p-3 bg-gray-50 dark:bg-gray-700 rounded-lg">
                <span class="text-gray-600 dark:text-gray-400">تاريخ التفعيل:</span>
                <span class="text-gray-900 dark:text-white text-sm">{{ formatDate(organization.activated_at) }}</span>
              </div>
              <div v-if="organization.suspended_at" class="flex justify-between items-center p-3 bg-red-50 dark:bg-red-900/30 rounded-lg border border-red-200 dark:border-red-800">
                <span class="text-red-600 dark:text-red-400">تاريخ الإيقاف:</span>
                <span class="text-red-700 dark:text-red-300 text-sm">{{ formatDate(organization.suspended_at) }}</span>
              </div>
            </div>
          </div>

          <!-- Actions -->
          <div class="bg-white dark:bg-gray-800 rounded-2xl shadow-lg p-6 border border-gray-100 dark:border-gray-700">
            <h3 class="text-lg font-bold text-gray-900 dark:text-white mb-4 flex items-center gap-2">
              <span>⚡</span>
              الإجراءات السريعة
            </h3>
            <div class="space-y-3">
              <button 
                v-if="organization.status === 'active'"
                @click="goToDatabase"
                class="w-full p-3 text-right bg-blue-50 dark:bg-blue-900/30 text-blue-700 dark:text-blue-300 rounded-lg hover:bg-blue-100 dark:hover:bg-blue-900/50 transition-colors flex items-center gap-3"
              >
                <span class="text-xl">🗄️</span>
                إدارة قاعدة البيانات
              </button>
              <button 
                v-if="organization.status === 'active'"
                @click="goToAdmin"
                class="w-full p-3 text-right bg-purple-50 dark:bg-purple-900/30 text-purple-700 dark:text-purple-300 rounded-lg hover:bg-purple-100 dark:hover:bg-purple-900/50 transition-colors flex items-center gap-3"
              >
                <span class="text-xl">👤</span>
                إنشاء مدير جديد
              </button>
              
              <!-- Suspend/Activate -->
              <button 
                v-if="organization.status === 'active'"
                @click="suspendOrg"
                :disabled="suspending"
                class="w-full p-3 text-right bg-red-50 dark:bg-red-900/30 text-red-700 dark:text-red-300 rounded-lg hover:bg-red-100 dark:hover:bg-red-900/50 transition-colors flex items-center gap-3 disabled:opacity-50"
              >
                <span class="text-xl">🚫</span>
                {{ suspending ? 'جاري الإيقاف...' : 'إيقاف المنظمة' }}
              </button>
              <button 
                v-if="organization.status === 'suspended'"
                @click="activateOrg"
                :disabled="activating"
                class="w-full p-3 text-right bg-green-50 dark:bg-green-900/30 text-green-700 dark:text-green-300 rounded-lg hover:bg-green-100 dark:hover:bg-green-900/50 transition-colors flex items-center gap-3 disabled:opacity-50"
              >
                <span class="text-xl">✅</span>
                {{ activating ? 'جاري التفعيل...' : 'إعادة تفعيل المنظمة' }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
