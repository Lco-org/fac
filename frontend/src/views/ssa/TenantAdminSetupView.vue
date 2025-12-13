<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'

const router = useRouter()
const route = useRoute()

// State
const orgId = ref(route.params.id)
const loading = ref(false)
const submitting = ref(false)
const error = ref(null)
const success = ref(false)
const organization = ref(null)
const adminResponse = ref(null)

const form = reactive({
  username: '',
  email: '',
  full_name: '',
  full_name_ar: ''
})

// API Base URL
const API_BASE = '/api/v1/ssa'

// Fetch organization details
const fetchOrganization = async () => {
  loading.value = true
  try {
    const response = await fetch(`${API_BASE}/organizations/${orgId.value}`, { 
      credentials: 'include' 
    })
    if (response.ok) {
      organization.value = await response.json()
      // Pre-fill email domain
      if (organization.value?.primary_domain) {
        form.email = `admin@${organization.value.primary_domain}`
      }
    }
  } catch (err) {
    console.error('Failed to fetch organization:', err)
  } finally {
    loading.value = false
  }
}

// Validate form
const isFormValid = ref(false)
const validateForm = () => {
  isFormValid.value = 
    form.username.length >= 3 &&
    form.email.includes('@') &&
    form.full_name.length >= 2
}

// Create tenant admin
const createAdmin = async () => {
  if (!isFormValid.value) return
  
  submitting.value = true
  error.value = null
  
  try {
    const token = localStorage.getItem('ssa_token')
    const response = await fetch(`${API_BASE}/organizations/${orgId.value}/admin`, {
      method: 'POST',
      headers: { 
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      credentials: 'include',
      body: JSON.stringify({
        username: form.username,
        email: form.email,
        full_name: form.full_name,
        full_name_ar: form.full_name_ar || null
      })
    })
    
    if (!response.ok) {
      const data = await response.json()
      throw new Error(data.error || 'فشل إنشاء المدير')
    }
    
    const data = await response.json()
    adminResponse.value = data.admin
    success.value = true
  } catch (err) {
    error.value = err.message
  } finally {
    submitting.value = false
  }
}

// Go to dashboard
const goToDashboard = () => {
  router.push({ name: 'SSADashboard' })
}

// Go back
const goBack = () => {
  router.push({ name: 'SSADashboard' })
}

// Copy to clipboard
const copyToClipboard = (text) => {
  navigator.clipboard.writeText(text)
}

onMounted(() => {
  fetchOrganization()
})
</script>

<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-900 p-6" dir="rtl">
    <div class="max-w-2xl mx-auto">
      <!-- Header -->
      <div class="mb-8">
        <button 
          @click="goBack"
          class="flex items-center gap-2 text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300 mb-4 transition-colors"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18"/>
          </svg>
          العودة
        </button>
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white flex items-center gap-3">
          <span class="text-4xl">👤</span>
          إنشاء مدير المنظمة
        </h1>
        <p class="mt-2 text-gray-500 dark:text-gray-400">
          إنشاء أول حساب مدير للمنظمة
        </p>
      </div>

      <!-- Loading State -->
      <div v-if="loading" class="flex items-center justify-center p-12">
        <div class="w-12 h-12 border-4 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
      </div>

      <!-- Success State -->
      <div v-else-if="success" class="bg-white dark:bg-gray-800 rounded-2xl shadow-lg p-8 border border-gray-100 dark:border-gray-700">
        <div class="text-center mb-8">
          <div class="w-24 h-24 mx-auto mb-6 bg-green-100 dark:bg-green-900/50 rounded-full flex items-center justify-center">
            <span class="text-5xl">🎉</span>
          </div>
          
          <h2 class="text-2xl font-bold text-gray-900 dark:text-white mb-2">
            تم إنشاء المدير بنجاح!
          </h2>
          <p class="text-gray-500 dark:text-gray-400">
            تم إرسال بيانات الدخول إلى البريد الإلكتروني
          </p>
        </div>

        <!-- Credentials Box -->
        <div class="bg-gradient-to-r from-blue-50 to-indigo-50 dark:from-blue-900/30 dark:to-indigo-900/30 rounded-xl p-6 border border-blue-100 dark:border-blue-800 mb-6">
          <h3 class="font-bold text-gray-900 dark:text-white mb-4 flex items-center gap-2">
            <span>🔑</span>
            بيانات الدخول (احتفظ بها لديك)
          </h3>
          
          <div class="space-y-3">
            <div class="flex justify-between items-center p-3 bg-white dark:bg-gray-800 rounded-lg">
              <span class="text-gray-600 dark:text-gray-400">اسم المستخدم:</span>
              <div class="flex items-center gap-2">
                <code class="font-mono font-bold text-gray-900 dark:text-white">{{ adminResponse?.username }}</code>
                <button @click="copyToClipboard(adminResponse?.username)" class="p-1 hover:bg-gray-100 dark:hover:bg-gray-700 rounded">
                  <svg class="w-4 h-4 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"/>
                  </svg>
                </button>
              </div>
            </div>
            
            <div class="flex justify-between items-center p-3 bg-white dark:bg-gray-800 rounded-lg">
              <span class="text-gray-600 dark:text-gray-400">البريد:</span>
              <div class="flex items-center gap-2">
                <code class="font-mono text-gray-900 dark:text-white">{{ adminResponse?.email }}</code>
                <button @click="copyToClipboard(adminResponse?.email)" class="p-1 hover:bg-gray-100 dark:hover:bg-gray-700 rounded">
                  <svg class="w-4 h-4 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"/>
                  </svg>
                </button>
              </div>
            </div>
            
            <div class="flex justify-between items-center p-3 bg-yellow-50 dark:bg-yellow-900/30 rounded-lg border border-yellow-200 dark:border-yellow-800">
              <span class="text-yellow-700 dark:text-yellow-400 font-medium">كلمة المرور المؤقتة:</span>
              <div class="flex items-center gap-2">
                <code class="font-mono font-bold text-yellow-800 dark:text-yellow-300">{{ adminResponse?.temporary_password }}</code>
                <button @click="copyToClipboard(adminResponse?.temporary_password)" class="p-1 hover:bg-yellow-100 dark:hover:bg-yellow-900 rounded">
                  <svg class="w-4 h-4 text-yellow-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"/>
                  </svg>
                </button>
              </div>
            </div>
            
            <div class="flex justify-between items-center p-3 bg-white dark:bg-gray-800 rounded-lg">
              <span class="text-gray-600 dark:text-gray-400">رابط الدخول:</span>
              <a 
                :href="adminResponse?.login_url" 
                target="_blank"
                class="text-blue-600 dark:text-blue-400 hover:underline font-mono text-sm"
              >
                {{ adminResponse?.login_url }}
              </a>
            </div>
          </div>
        </div>

        <!-- Warning -->
        <div class="mb-6 p-4 bg-amber-50 dark:bg-amber-900/30 border border-amber-200 dark:border-amber-800 rounded-xl">
          <p class="text-amber-700 dark:text-amber-400 text-sm">
            ⚠️ <strong>هام:</strong> كلمة المرور هذه مؤقتة. يجب على المدير تغييرها عند أول تسجيل دخول.
          </p>
        </div>

        <!-- Complete Setup -->
        <div class="text-center">
          <p class="text-green-600 dark:text-green-400 font-medium mb-4 flex items-center justify-center gap-2">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
            </svg>
            اكتمل إعداد المنظمة بنجاح!
          </p>
          
          <button 
            @click="goToDashboard"
            class="w-full py-4 bg-gradient-to-r from-gray-700 to-gray-800 text-white rounded-xl font-bold text-lg shadow-lg hover:shadow-xl transition-all flex items-center justify-center gap-3"
          >
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6"/>
            </svg>
            العودة للوحة التحكم
          </button>
        </div>
      </div>

      <!-- Form -->
      <div v-else class="bg-white dark:bg-gray-800 rounded-2xl shadow-lg p-8 border border-gray-100 dark:border-gray-700">
        <!-- Organization Info -->
        <div v-if="organization" class="mb-8 p-4 bg-blue-50 dark:bg-blue-900/30 rounded-xl">
          <p class="text-blue-700 dark:text-blue-300 text-sm">
            <strong>المنظمة:</strong> {{ organization.name }} | 
            <strong>الدومين:</strong> {{ organization.primary_domain }}
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

        <form @submit.prevent="createAdmin" @input="validateForm">
          <!-- Username -->
          <div class="mb-6">
            <label class="block text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">
              اسم المستخدم <span class="text-red-500">*</span>
            </label>
            <input 
              v-model="form.username"
              type="text" 
              required
              minlength="3"
              placeholder="admin"
              class="w-full px-4 py-3 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all"
              dir="ltr"
            />
          </div>

          <!-- Email -->
          <div class="mb-6">
            <label class="block text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">
              البريد الإلكتروني <span class="text-red-500">*</span>
            </label>
            <input 
              v-model="form.email"
              type="email" 
              required
              placeholder="admin@company.com"
              class="w-full px-4 py-3 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all"
              dir="ltr"
            />
          </div>

          <!-- Full Name -->
          <div class="mb-6">
            <label class="block text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">
              الاسم الكامل (English) <span class="text-red-500">*</span>
            </label>
            <input 
              v-model="form.full_name"
              type="text" 
              required
              placeholder="John Doe"
              class="w-full px-4 py-3 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all"
              dir="ltr"
            />
          </div>

          <!-- Arabic Name -->
          <div class="mb-8">
            <label class="block text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">
              الاسم بالعربية (اختياري)
            </label>
            <input 
              v-model="form.full_name_ar"
              type="text" 
              placeholder="جون دو"
              class="w-full px-4 py-3 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all"
              dir="rtl"
            />
          </div>

          <!-- Info Note -->
          <div class="mb-6 p-4 bg-blue-50 dark:bg-blue-900/30 border border-blue-200 dark:border-blue-800 rounded-xl">
            <p class="text-blue-700 dark:text-blue-400 text-sm">
              💡 <strong>ملاحظة:</strong> سيتم إنشاء كلمة مرور مؤقتة (12 حرف) وإرسالها مع رسالة ترحيب إلى البريد المحدد.
            </p>
          </div>

          <!-- Submit Button -->
          <button 
            type="submit"
            :disabled="submitting || !isFormValid"
            class="w-full py-4 bg-gradient-to-r from-purple-500 to-indigo-600 text-white rounded-xl font-bold text-lg shadow-lg hover:shadow-xl transition-all disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-3"
          >
            <span v-if="submitting" class="w-6 h-6 border-2 border-white border-t-transparent rounded-full animate-spin"></span>
            <span v-else>
              <svg class="w-6 h-6 inline-block ml-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M18 9v3m0 0v3m0-3h3m-3 0h-3m-2-5a4 4 0 11-8 0 4 4 0 018 0zM3 20a6 6 0 0112 0v1H3v-1z"/>
              </svg>
              إنشاء حساب المدير
            </span>
          </button>
        </form>
      </div>
    </div>
  </div>
</template>
