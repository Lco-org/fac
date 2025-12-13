<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'

const router = useRouter()
const route = useRoute()

// State
const orgId = ref(route.params.id)
const email = ref(route.query.email || '')
const code = ref('')
const loading = ref(false)
const resending = ref(false)
const error = ref(null)
const success = ref(false)
const organization = ref(null)
const remainingTime = ref(300) // 5 minutes
const timer = ref(null)

// API Base URL
const API_BASE = '/api/v1/ssa'

// Format remaining time
const formattedTime = computed(() => {
  const mins = Math.floor(remainingTime.value / 60)
  const secs = remainingTime.value % 60
  return `${mins}:${secs.toString().padStart(2, '0')}`
})

// Start countdown timer
const startTimer = () => {
  timer.value = setInterval(() => {
    if (remainingTime.value > 0) {
      remainingTime.value--
    } else {
      clearInterval(timer.value)
    }
  }, 1000)
}

// Fetch organization details
const fetchOrganization = async () => {
  try {
    const token = localStorage.getItem('ssa_token')
    const response = await fetch(`${API_BASE}/organizations/${orgId.value}`, { 
      credentials: 'include',
      headers: { 'Authorization': `Bearer ${token}` }
    })
    if (response.ok) {
      organization.value = await response.json()
    }
  } catch (err) {
    console.error('Failed to fetch organization:', err)
  }
}

// Verify code
const verifyCode = async () => {
  if (code.value.length !== 6) {
    error.value = 'رمز التحقق يجب أن يكون 6 أحرف'
    return
  }
  
  loading.value = true
  error.value = null
  
  try {
    const token = localStorage.getItem('ssa_token')
    const response = await fetch(`${API_BASE}/organizations/${orgId.value}/verify`, {
      method: 'POST',
      headers: { 
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      credentials: 'include',
      body: JSON.stringify({ code: code.value.toUpperCase() })
    })
    
    if (!response.ok) {
      const data = await response.json()
      throw new Error(data.error || 'فشل التحقق')
    }
    
    success.value = true
    clearInterval(timer.value)
  } catch (err) {
    error.value = err.message
  } finally {
    loading.value = false
  }
}

// Resend verification code
const resendCode = async () => {
  resending.value = true
  error.value = null
  
  try {
    const token = localStorage.getItem('ssa_token')
    const response = await fetch(`${API_BASE}/organizations/${orgId.value}/resend-verification`, {
      method: 'POST',
      headers: { 
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      credentials: 'include',
      body: JSON.stringify({ email: email.value })
    })
    
    if (!response.ok) {
      const data = await response.json()
      throw new Error(data.error || 'فشل إعادة الإرسال')
    }
    
    remainingTime.value = 300
    startTimer()
  } catch (err) {
    error.value = err.message
  } finally {
    resending.value = false
  }
}

// Go to database creation
const goToCreateDatabase = () => {
  router.push({ 
    name: 'SSACreateDatabase', 
    params: { id: orgId.value }
  })
}

// Go back
const goBack = () => {
  router.push({ name: 'SSADashboard' })
}

// Handle code input (auto-uppercase and auto-submit)
const handleCodeInput = (event) => {
  code.value = event.target.value.toUpperCase().slice(0, 6)
  if (code.value.length === 6) {
    // Auto-verify when 6 chars entered
    verifyCode()
  }
}

onMounted(() => {
  fetchOrganization()
  startTimer()
})
</script>

<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-900 p-6" dir="rtl">
    <div class="max-w-xl mx-auto">
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
          <span class="text-4xl">📧</span>
          التحقق من البريد الإلكتروني
        </h1>
        <p class="mt-2 text-gray-500 dark:text-gray-400">
          أدخل رمز التحقق المرسل إلى بريدك الإلكتروني
        </p>
      </div>

      <!-- Success State -->
      <div v-if="success" class="bg-white dark:bg-gray-800 rounded-2xl shadow-lg p-8 border border-gray-100 dark:border-gray-700 text-center">
        <div class="w-24 h-24 mx-auto mb-6 bg-green-100 dark:bg-green-900/50 rounded-full flex items-center justify-center animate-bounce">
          <span class="text-5xl">✅</span>
        </div>
        
        <h2 class="text-2xl font-bold text-gray-900 dark:text-white mb-4">
          تم التحقق بنجاح!
        </h2>
        
        <p class="text-gray-600 dark:text-gray-400 mb-6">
          تم تفعيل المنظمة. الآن يمكنك إنشاء قاعدة البيانات وإعداد المدير الأول.
        </p>
        
        <button 
          @click="goToCreateDatabase"
          class="w-full py-4 bg-gradient-to-r from-emerald-500 to-green-600 text-white rounded-xl font-bold text-lg shadow-lg hover:shadow-xl transition-all flex items-center justify-center gap-3"
        >
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4"/>
          </svg>
          متابعة إلى إنشاء قاعدة البيانات
        </button>
      </div>

      <!-- Verification Form -->
      <div v-else class="bg-white dark:bg-gray-800 rounded-2xl shadow-lg p-8 border border-gray-100 dark:border-gray-700">
        <!-- Organization Info -->
        <div v-if="organization" class="mb-6 p-4 bg-blue-50 dark:bg-blue-900/30 rounded-xl">
          <p class="text-blue-700 dark:text-blue-300 text-sm">
            <strong>المنظمة:</strong> {{ organization.name }}
            <br/>
            <strong>الدومين:</strong> {{ organization.primary_domain }}
          </p>
        </div>

        <!-- Email Display -->
        <div class="text-center mb-8">
          <p class="text-gray-600 dark:text-gray-400 mb-2">تم إرسال الرمز إلى:</p>
          <code class="text-lg font-mono text-blue-600 dark:text-blue-400 bg-blue-50 dark:bg-blue-900/30 px-4 py-2 rounded-lg">
            {{ email }}
          </code>
        </div>

        <!-- Timer -->
        <div class="text-center mb-6">
          <div class="inline-flex items-center gap-2 px-4 py-2 bg-gray-100 dark:bg-gray-700 rounded-full">
            <svg class="w-5 h-5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"/>
            </svg>
            <span :class="['font-mono font-bold text-lg', remainingTime < 60 ? 'text-red-500' : 'text-gray-700 dark:text-gray-300']">
              {{ formattedTime }}
            </span>
          </div>
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

        <!-- Code Input -->
        <div class="mb-6">
          <label class="block text-sm font-semibold text-gray-700 dark:text-gray-300 mb-3 text-center">
            أدخل رمز التحقق (6 أحرف)
          </label>
          <input 
            :value="code"
            @input="handleCodeInput"
            type="text" 
            maxlength="6"
            :disabled="loading"
            placeholder="XXXXXX"
            class="w-full px-6 py-5 text-center text-3xl font-mono font-bold tracking-[0.5em] bg-gray-50 dark:bg-gray-700 border-2 border-gray-200 dark:border-gray-600 rounded-xl focus:ring-4 focus:ring-blue-500/30 focus:border-blue-500 transition-all uppercase"
            dir="ltr"
          />
        </div>

        <!-- Verify Button -->
        <button 
          @click="verifyCode"
          :disabled="loading || code.length !== 6"
          class="w-full py-4 bg-gradient-to-r from-blue-600 to-indigo-600 text-white rounded-xl font-bold text-lg shadow-lg hover:shadow-xl transition-all disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-3 mb-4"
        >
          <span v-if="loading" class="w-6 h-6 border-2 border-white border-t-transparent rounded-full animate-spin"></span>
          <span v-else>
            <svg class="w-6 h-6 inline-block ml-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
            </svg>
            تحقق
          </span>
        </button>

        <!-- Resend Link -->
        <div class="text-center">
          <button 
            @click="resendCode"
            :disabled="resending || remainingTime > 0"
            class="text-blue-600 dark:text-blue-400 hover:underline disabled:opacity-50 disabled:cursor-not-allowed disabled:no-underline flex items-center justify-center gap-2 mx-auto"
          >
            <span v-if="resending" class="w-4 h-4 border-2 border-blue-600 border-t-transparent rounded-full animate-spin"></span>
            <span v-else>
              {{ remainingTime > 0 ? `إعادة الإرسال بعد ${formattedTime}` : 'إعادة إرسال الرمز' }}
            </span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Bounce animation for success icon */
@keyframes bounce {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-10px); }
}
.animate-bounce {
  animation: bounce 0.6s ease-in-out;
}
</style>
