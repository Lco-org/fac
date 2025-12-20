<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import ssa from '@/api/ssa'

const router = useRouter()
const route = useRoute()

// State
const orgId = ref(route.params.id)
const loading = ref(true)
const error = ref(null)
const organization = ref(null)
const suspending = ref(false)
const activating = ref(false)
const plans = ref([])

const currentPlan = computed(() => {
  if (!organization.value || !organization.value.subscription_plan_id) return null
  return plans.value.find(p => String(p.id) === String(organization.value.subscription_plan_id))
})

// Deletion state
const showDeleteModal = ref(false)
const deleteStep = ref(1) // 1: enter password, 2: enter OTP
const deletionLoading = ref(false)
const deletionError = ref(null)
const otpCode = ref('')
const adminPassword = ref('')
const maskedEmail = ref('')

// Update Subscription State
const showSubscriptionModal = ref(false)
const subForm = ref({
  plan_id: '',
  start_date: '',
  end_date: '',
  status: 'active'
})
const subSaving = ref(false)
const subError = ref(null)

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

const fetchPlans = async () => {
  try {
    const response = await ssa.organizations.getPlans()
    plans.value = response.data.plans || []
  } catch (err) {
    console.error('Failed to fetch plans', err)
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

// Open Subscription Modal
const openSubscriptionModal = () => {
  subForm.value = {
    plan_id: organization.value.subscription_plan_id,
    start_date: organization.value.subscription_start_date ? organization.value.subscription_start_date.split('T')[0] : '',
    end_date: organization.value.subscription_expires_at ? organization.value.subscription_expires_at.split('T')[0] : '',
    status: organization.value.subscription_status
  }
  showSubscriptionModal.value = true
}

const closeSubscriptionModal = () => {
  showSubscriptionModal.value = false
  subError.value = null
}

const updateSubscription = async () => {
  subSaving.value = true
  subError.value = null
  try {
    const payload = {
      subscription_plan_id: subForm.value.plan_id,
      subscription_status: subForm.value.status
    }

    if (subForm.value.start_date) {
      payload.subscription_start = new Date(subForm.value.start_date).toISOString()
    }
    if (subForm.value.end_date) {
      payload.subscription_end = new Date(subForm.value.end_date).toISOString()
    }

    await ssa.organizations.updateSubscription(orgId.value, payload)
    await fetchOrganization()
    closeSubscriptionModal()
  } catch (err) {
    subError.value = err.response?.data?.error || err.message || 'فشل تحديث الاشتراك'
  } finally {
    subSaving.value = false
  }
}

// ============================================================================
// DELETION FLOW
// ============================================================================

// Open delete modal
const openDeleteModal = () => {
  showDeleteModal.value = true
  deleteStep.value = 1
  deletionError.value = null
  otpCode.value = ''
  adminPassword.value = ''
  maskedEmail.value = ''
}

// Close delete modal
const closeDeleteModal = () => {
  showDeleteModal.value = false
  deleteStep.value = 1
  deletionError.value = null
  otpCode.value = ''
  adminPassword.value = ''
}

// Request deletion (send OTP) - requires password verification
const requestDeletion = async () => {
  if (!adminPassword.value) {
    deletionError.value = 'كلمة المرور مطلوبة'
    return
  }
  
  deletionLoading.value = true
  deletionError.value = null
  
  try {
    const token = localStorage.getItem('ssa_token')
    const response = await fetch(`${API_BASE}/organizations/${orgId.value}/request-deletion`, {
      method: 'POST',
      headers: { 
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      credentials: 'include',
      body: JSON.stringify({ password: adminPassword.value })
    })
    
    const data = await response.json()
    if (!response.ok) {
      throw new Error(data.error || 'فشل التحقق من كلمة المرور')
    }
    
    // Save masked email for display
    maskedEmail.value = data.masked_email || '***@***'
    
    // Move to OTP step
    deleteStep.value = 2
  } catch (err) {
    deletionError.value = err.message
  } finally {
    deletionLoading.value = false
  }
}

// Confirm deletion with OTP
const confirmDeletion = async () => {
  if (otpCode.value.length !== 6) {
    deletionError.value = 'رمز التحقق يجب أن يكون 6 أحرف'
    return
  }
  
  deletionLoading.value = true
  deletionError.value = null
  
  try {
    const token = localStorage.getItem('ssa_token')
    const response = await fetch(`${API_BASE}/organizations/${orgId.value}/confirm-deletion`, {
      method: 'POST',
      headers: { 
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      credentials: 'include',
      body: JSON.stringify({ code: otpCode.value.toUpperCase() })
    })
    
    const data = await response.json()
    if (!response.ok) {
      throw new Error(data.error || 'فشل حذف المنظمة')
    }
    
    // Success - redirect to dashboard
    alert('✅ تم حذف المنظمة بنجاح')
    router.push({ name: 'SSADashboard' })
  } catch (err) {
    deletionError.value = err.message
  } finally {
    deletionLoading.value = false
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
  fetchPlans()
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

          <!-- Subscription Details -->
          <div class="bg-white dark:bg-gray-800 rounded-2xl shadow-lg p-6 border border-gray-100 dark:border-gray-700">
            <h3 class="text-lg font-bold text-gray-900 dark:text-white mb-4 flex items-center gap-2">
              <span>💳</span>
              تفاصيل الاشتراك
            </h3>

            <div class="space-y-4">
              <div class="flex justify-between items-center p-3 bg-gray-50 dark:bg-gray-700 rounded-lg">
                <span class="text-gray-600 dark:text-gray-400">الخطة الحالية:</span>
                <span class="text-gray-900 dark:text-white font-bold">
                  {{ currentPlan ? (currentPlan.name_ar || currentPlan.name) : 'غير محدد' }}
                </span>
              </div>
              <div class="flex justify-between items-center p-3 bg-gray-50 dark:bg-gray-700 rounded-lg">
                <span class="text-gray-600 dark:text-gray-400">حالة الاشتراك:</span>
                <span :class="['px-2 py-0.5 rounded text-sm', organization.subscription_status === 'active' ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-800']">
                  {{ organization.subscription_status === 'active' ? 'نشط' : (organization.subscription_status || 'غير معروف') }}
                </span>
              </div>
              <div v-if="organization.subscription_start_date" class="flex justify-between items-center p-3 bg-gray-50 dark:bg-gray-700 rounded-lg">
                <span class="text-gray-600 dark:text-gray-400">تاريخ البدء:</span>
                <span class="text-gray-900 dark:text-white text-sm">{{ formatDate(organization.subscription_start_date) }}</span>
              </div>
              <div v-if="organization.subscription_expires_at" class="flex justify-between items-center p-3 bg-gray-50 dark:bg-gray-700 rounded-lg">
                <span class="text-gray-600 dark:text-gray-400">تاريخ الانتهاء:</span>
                <span class="text-gray-900 dark:text-white text-sm">{{ formatDate(organization.subscription_expires_at) }}</span>
              </div>
              
              <!-- Features List -->
              <div v-if="currentPlan && currentPlan.features && currentPlan.features.length > 0" class="pt-4 border-t border-gray-100 dark:border-gray-700">
                <span class="block text-sm text-gray-600 dark:text-gray-400 mb-2">المميزات المفعلة:</span>
                <div class="flex flex-wrap gap-2">
                  <span 
                    v-for="feature in currentPlan.features" 
                    :key="feature" 
                    class="px-2 py-1 bg-blue-50 dark:bg-blue-900/30 text-blue-700 dark:text-blue-300 text-xs rounded-md border border-blue-100 dark:border-blue-800"
                  >
                    {{ feature }}
                  </span>
                </div>
              </div>
            </div>
          </div>
          
          <!-- Edit Subscription Modal -->
          <div v-if="showSubscriptionModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
            <div class="absolute inset-0 bg-black/60 backdrop-blur-sm" @click="closeSubscriptionModal"></div>
            <div class="relative bg-white dark:bg-gray-800 rounded-2xl shadow-2xl max-w-lg w-full p-6">
              <div class="flex justify-between items-center mb-6">
                <h3 class="text-xl font-bold text-gray-900 dark:text-white">تعديل الاشتراك</h3>
                <button @click="closeSubscriptionModal" class="text-gray-400 hover:text-gray-600">✕</button>
              </div>

              <div class="space-y-4">
                <!-- Plan -->
                <div>
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">خطة الاشتراك</label>
                  <select v-model="subForm.plan_id" class="w-full p-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-white">
                    <option v-for="plan in plans" :key="plan.id" :value="plan.id">
                      {{ plan.name_ar || plan.name }} ({{ plan.price_monthly }} {{ plan.currency }})
                    </option>
                  </select>
                </div>

                <!-- Status -->
                <div>
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">حالة الاشتراك</label>
                  <select v-model="subForm.status" class="w-full p-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-white">
                    <option value="active">نشط</option>
                    <option value="suspended">موقوف</option>
                    <option value="cancelled">ملغى</option>
                    <option value="expired">منتهي</option>
                  </select>
                </div>

                <!-- Dates -->
                <div class="grid grid-cols-2 gap-4">
                  <div>
                    <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">تاريخ البدء</label>
                    <input v-model="subForm.start_date" type="date" class="w-full p-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-white">
                  </div>
                  <div>
                    <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">تاريخ الانتهاء</label>
                    <input v-model="subForm.end_date" type="date" class="w-full p-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-white">
                  </div>
                </div>

                <div v-if="subError" class="p-3 bg-red-50 text-red-600 rounded-lg text-sm">
                  {{ subError }}
                </div>

                <div class="flex gap-3 pt-4">
                  <button @click="closeSubscriptionModal" class="flex-1 py-2 text-gray-600 dark:text-gray-400">إلغاء</button>
                  <button 
                    @click="updateSubscription" 
                    :disabled="subSaving"
                    class="flex-1 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50"
                  >
                    {{ subSaving ? 'جاري الحفظ...' : 'حفظ التغييرات' }}
                  </button>
                </div>
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
              <button 
                v-if="organization.status === 'active'"
                @click="openSubscriptionModal"
                class="w-full p-3 text-right bg-indigo-50 dark:bg-indigo-900/30 text-indigo-700 dark:text-indigo-300 rounded-lg hover:bg-indigo-100 dark:hover:bg-indigo-900/50 transition-colors flex items-center gap-3"
              >
                <span class="text-xl">💳</span>
                تعديل الاشتراك
              </button>
              
              <!-- Suspend/Activate -->
              <button 
                v-if="organization.status === 'active'"
                @click="suspendOrg"
                :disabled="suspending"
                class="w-full p-3 text-right bg-orange-50 dark:bg-orange-900/30 text-orange-700 dark:text-orange-300 rounded-lg hover:bg-orange-100 dark:hover:bg-orange-900/50 transition-colors flex items-center gap-3 disabled:opacity-50"
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

              <!-- Delete Button (Danger Zone) -->
              <div class="mt-6 pt-4 border-t border-gray-200 dark:border-gray-600">
                <p class="text-xs text-gray-500 dark:text-gray-400 mb-3">⚠️ منطقة الخطر</p>
                <button 
                  @click="openDeleteModal"
                  class="w-full p-3 text-right bg-red-600 hover:bg-red-700 text-white rounded-lg transition-colors flex items-center gap-3 font-medium"
                >
                  <span class="text-xl">🗑️</span>
                  حذف المنظمة نهائياً
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Delete Confirmation Modal -->
      <div v-if="showDeleteModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <!-- Backdrop -->
        <div class="absolute inset-0 bg-black/60 backdrop-blur-sm" @click="closeDeleteModal"></div>
        
        <!-- Modal -->
        <div class="relative bg-white dark:bg-gray-800 rounded-2xl shadow-2xl max-w-md w-full p-6 border border-red-200 dark:border-red-800">
          <!-- Step 1: Confirm Request -->
          <div v-if="deleteStep === 1">
            <div class="text-center mb-6">
              <div class="w-16 h-16 bg-red-100 dark:bg-red-900/30 rounded-full flex items-center justify-center mx-auto mb-4">
                <span class="text-4xl">⚠️</span>
              </div>
              <h3 class="text-xl font-bold text-gray-900 dark:text-white">تأكيد حذف المنظمة</h3>
              <p class="text-gray-600 dark:text-gray-400 mt-2">هل أنت متأكد من حذف <strong class="text-red-600">{{ organization.name }}</strong>؟</p>
            </div>
            
            <div class="bg-red-50 dark:bg-red-900/30 rounded-lg p-4 mb-4 text-sm text-red-700 dark:text-red-300">
              <strong>تحذير:</strong> هذه العملية لا يمكن التراجع عنها. سيتم حذف:
              <ul class="mt-2 list-disc list-inside">
                <li>قاعدة بيانات المنظمة</li>
                <li>جميع المستخدمين والبيانات</li>
                <li>جميع الملفات والمرفقات</li>
              </ul>
            </div>

            <div class="mb-4">
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">أدخل كلمة المرور للتأكيد</label>
              <input 
                v-model="adminPassword" 
                type="password" 
                class="w-full p-3 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-red-500"
                placeholder="••••••••••••"
                @keyup.enter="requestDeletion"
              />
              <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">سيتم إرسال رمز التحقق إلى بريدك المسجل فقط</p>
            </div>

            <div v-if="deletionError" class="mb-4 p-3 bg-red-100 dark:bg-red-900/50 text-red-700 dark:text-red-300 rounded-lg text-sm">
              {{ deletionError }}
            </div>

            <div class="flex gap-3">
              <button 
                @click="closeDeleteModal" 
                class="flex-1 p-3 bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded-lg hover:bg-gray-300 dark:hover:bg-gray-600 transition-colors"
              >
                إلغاء
              </button>
              <button 
                @click="requestDeletion" 
                :disabled="deletionLoading || !adminPassword"
                class="flex-1 p-3 bg-red-600 text-white rounded-lg hover:bg-red-700 transition-colors disabled:opacity-50"
              >
                {{ deletionLoading ? 'جاري التحقق...' : 'تأكيد وإرسال الرمز' }}
              </button>
            </div>
          </div>

          <!-- Step 2: Enter OTP -->
          <div v-if="deleteStep === 2">
            <div class="text-center mb-6">
              <div class="w-16 h-16 bg-yellow-100 dark:bg-yellow-900/30 rounded-full flex items-center justify-center mx-auto mb-4">
                <span class="text-4xl">🔐</span>
              </div>
              <h3 class="text-xl font-bold text-gray-900 dark:text-white">أدخل رمز التحقق</h3>
              <p class="text-gray-600 dark:text-gray-400 mt-2">تم إرسال رمز مكون من 6 أحرف إلى:</p>
              <code class="text-lg font-mono text-blue-600 dark:text-blue-400">{{ maskedEmail }}</code>
            </div>

            <div class="mb-4">
              <input 
                v-model="otpCode" 
                type="text" 
                maxlength="6"
                class="w-full p-4 text-center text-2xl tracking-widest font-mono rounded-lg border-2 border-red-300 dark:border-red-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-red-500 uppercase"
                placeholder="______"
              />
              <p class="text-xs text-gray-500 mt-2 text-center">⏱️ الرمز صالح لمدة 10 دقائق</p>
            </div>

            <div v-if="deletionError" class="mb-4 p-3 bg-red-100 dark:bg-red-900/50 text-red-700 dark:text-red-300 rounded-lg text-sm">
              {{ deletionError }}
            </div>

            <div class="flex gap-3">
              <button 
                @click="deleteStep = 1" 
                class="flex-1 p-3 bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded-lg hover:bg-gray-300 dark:hover:bg-gray-600 transition-colors"
              >
                رجوع
              </button>
              <button 
                @click="confirmDeletion" 
                :disabled="deletionLoading || otpCode.length !== 6"
                class="flex-1 p-3 bg-red-600 text-white rounded-lg hover:bg-red-700 transition-colors disabled:opacity-50"
              >
                {{ deletionLoading ? 'جاري الحذف...' : '🗑️ تأكيد الحذف النهائي' }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
