<script setup>
import { ref, reactive, computed, watch } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()

// State
const step = ref(1) // 1: Login, 2: 2FA
const loading = ref(false)
const error = ref(null)
const pendingSessionId = ref(null)
const maskedEmail = ref('')

const loginForm = reactive({
  username: '',
  password: ''
})

const otpCode = ref('')
const otpDigits = ref(['', '', '', '', '', ''])

// API Base
const API_BASE = '/api/v1/system-auth'

// Watch OTP digits for auto-submit
watch(otpDigits, (newVal) => {
  if (newVal.every(d => d !== '')) {
    otpCode.value = newVal.join('')
    verify2FA()
  }
}, { deep: true })

// Handle OTP input
const handleOTPInput = (index, event) => {
  const value = event.target.value.replace(/\D/g, '')
  if (value) {
    otpDigits.value[index] = value[0]
    // Move to next input
    if (index < 5) {
      const nextInput = event.target.parentElement.querySelector(`input:nth-child(${index + 2})`)
      if (nextInput) nextInput.focus()
    }
  }
}

const handleOTPKeydown = (index, event) => {
  if (event.key === 'Backspace' && !otpDigits.value[index] && index > 0) {
    const prevInput = event.target.parentElement.querySelector(`input:nth-child(${index})`)
    if (prevInput) {
      prevInput.focus()
      otpDigits.value[index - 1] = ''
    }
  }
}

// Login with password
const login = async () => {
  if (!loginForm.username || !loginForm.password) {
    error.value = 'يرجى إدخال اسم المستخدم وكلمة المرور'
    return
  }

  loading.value = true
  error.value = null

  try {
    const response = await fetch(`${API_BASE}/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify({
        username: loginForm.username,
        password: loginForm.password
      })
    })

    const data = await response.json()

    if (!response.ok) {
      throw new Error(data.error || 'فشل تسجيل الدخول')
    }

    if (data.requires_2fa) {
      pendingSessionId.value = data.pending_session_id
      maskedEmail.value = data.message
      step.value = 2
    } else if (data.token) {
      // Immediate login success
      localStorage.setItem('ssa_token', data.token)
      localStorage.setItem('ssa_admin', JSON.stringify(data.admin))
      router.push({ name: 'SSADashboard' })
    }
  } catch (err) {
    error.value = err.message
  } finally {
    loading.value = false
  }
}

// Verify 2FA code
const verify2FA = async () => {
  const code = otpDigits.value.join('')
  if (code.length !== 6) return

  loading.value = true
  error.value = null

  try {
    const response = await fetch(`${API_BASE}/verify-2fa`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify({
        session_id: pendingSessionId.value,
        code: code
      })
    })

    const data = await response.json()

    if (!response.ok) {
      throw new Error(data.error || 'فشل التحقق')
    }

    // Store token and redirect
    if (data.token) {
      localStorage.setItem('ssa_token', data.token)
      localStorage.setItem('ssa_admin', JSON.stringify(data.admin))
      router.push({ name: 'SSADashboard' })
    }
  } catch (err) {
    error.value = err.message
    // Clear OTP on error
    otpDigits.value = ['', '', '', '', '', '']
  } finally {
    loading.value = false
  }
}

// Clear and retry
const retry = () => {
  step.value = 1
  error.value = null
  pendingSessionId.value = null
  loginForm.username = ''
  loginForm.password = ''
  otpDigits.value = ['', '', '', '', '', '']
}
</script>

<template>
  <div class="min-h-screen bg-gradient-to-br from-slate-900 via-purple-900 to-slate-900 flex items-center justify-center p-4" dir="rtl">
    <!-- Background Effects -->
    <div class="absolute inset-0 overflow-hidden">
      <div class="absolute top-1/4 left-1/4 w-96 h-96 bg-purple-500/20 rounded-full blur-3xl"></div>
      <div class="absolute bottom-1/4 right-1/4 w-96 h-96 bg-blue-500/20 rounded-full blur-3xl"></div>
    </div>

    <!-- Login Card -->
    <div class="relative w-full max-w-md">
      <!-- Header -->
      <div class="text-center mb-8">
        <div class="inline-flex items-center justify-center w-20 h-20 bg-gradient-to-r from-purple-600 to-indigo-600 rounded-2xl shadow-2xl mb-4">
          <span class="text-4xl">🛡️</span>
        </div>
        <h1 class="text-3xl font-bold text-white mb-2">لوحة تحكم المشرف الأعلى</h1>
        <p class="text-gray-400 text-sm">
          <span class="inline-block px-3 py-1 bg-purple-500/20 rounded-full text-purple-300 text-xs font-medium">
            SUPER SYSTEM ADMIN
          </span>
        </p>
      </div>

      <!-- Card -->
      <div class="bg-slate-800/80 backdrop-blur-xl rounded-3xl shadow-2xl border border-slate-700/50 overflow-hidden">
        <!-- Step 1: Login Form -->
        <div v-if="step === 1" class="p-8">
          <h2 class="text-xl font-bold text-white mb-6 text-center">تسجيل الدخول</h2>

          <!-- Error Alert -->
          <div v-if="error" class="mb-6 p-4 bg-red-500/20 border border-red-500/50 rounded-xl text-red-300 text-sm flex items-center gap-2">
            <span class="text-lg">⚠️</span>
            {{ error }}
          </div>

          <form @submit.prevent="login" class="space-y-5">
            <!-- Username -->
            <div>
              <label class="block text-sm font-medium text-gray-300 mb-2">اسم المستخدم</label>
              <div class="relative">
                <input 
                  v-model="loginForm.username"
                  type="text"
                  required
                  autocomplete="username"
                  class="w-full px-4 py-4 bg-slate-700/50 border border-slate-600 rounded-xl text-white placeholder-gray-500 focus:border-purple-500 focus:ring-2 focus:ring-purple-500/20 transition-all pr-12"
                  placeholder="superadmin"
                  dir="ltr"
                />
                <span class="absolute left-4 top-1/2 -translate-y-1/2 text-gray-400">👤</span>
              </div>
            </div>

            <!-- Password -->
            <div>
              <label class="block text-sm font-medium text-gray-300 mb-2">كلمة المرور</label>
              <div class="relative">
                <input 
                  v-model="loginForm.password"
                  type="password"
                  required
                  autocomplete="current-password"
                  class="w-full px-4 py-4 bg-slate-700/50 border border-slate-600 rounded-xl text-white placeholder-gray-500 focus:border-purple-500 focus:ring-2 focus:ring-purple-500/20 transition-all pr-12"
                  placeholder="••••••••••••"
                  dir="ltr"
                />
                <span class="absolute left-4 top-1/2 -translate-y-1/2 text-gray-400">🔒</span>
              </div>
            </div>

            <!-- Submit -->
            <button 
              type="submit"
              :disabled="loading"
              class="w-full py-4 bg-gradient-to-r from-purple-600 to-indigo-600 text-white rounded-xl font-bold text-lg shadow-lg hover:shadow-purple-500/25 transition-all disabled:opacity-50 flex items-center justify-center gap-2"
            >
              <span v-if="loading" class="w-5 h-5 border-2 border-white border-t-transparent rounded-full animate-spin"></span>
              <span v-else>
                🔐 تسجيل الدخول
              </span>
            </button>
          </form>

          <!-- Security Notice -->
          <div class="mt-6 p-4 bg-slate-700/30 rounded-xl border border-slate-600/50">
            <p class="text-xs text-gray-400 text-center">
              🔒 هذه الصفحة محمية بتشفير عالي الأمان.
              <br/>جميع محاولات الدخول مسجلة ومراقبة.
            </p>
          </div>
        </div>

        <!-- Step 2: 2FA Verification -->
        <div v-else class="p-8">
          <div class="text-center mb-6">
            <div class="inline-flex items-center justify-center w-16 h-16 bg-purple-500/20 rounded-2xl mb-4">
              <span class="text-3xl">📧</span>
            </div>
            <h2 class="text-xl font-bold text-white mb-2">التحقق بخطوتين</h2>
            <p class="text-gray-400 text-sm">{{ maskedEmail }}</p>
          </div>

          <!-- Error -->
          <div v-if="error" class="mb-6 p-4 bg-red-500/20 border border-red-500/50 rounded-xl text-red-300 text-sm">
            {{ error }}
          </div>

          <!-- OTP Input -->
          <div class="flex justify-center gap-3 mb-6" dir="ltr">
            <input
              v-for="(digit, index) in otpDigits"
              :key="index"
              :value="digit"
              @input="handleOTPInput(index, $event)"
              @keydown="handleOTPKeydown(index, $event)"
              type="text"
              maxlength="1"
              class="w-12 h-14 text-center text-2xl font-bold bg-slate-700/50 border-2 border-slate-600 rounded-xl text-white focus:border-purple-500 focus:ring-2 focus:ring-purple-500/20 transition-all"
            />
          </div>

          <!-- Loading -->
          <div v-if="loading" class="flex justify-center mb-6">
            <div class="w-8 h-8 border-3 border-purple-500 border-t-transparent rounded-full animate-spin"></div>
          </div>

          <!-- Retry Button -->
          <div class="text-center">
            <button @click="retry" class="text-purple-400 hover:text-purple-300 text-sm transition-colors">
              ↩️ العودة لتسجيل الدخول
            </button>
          </div>

          <!-- Timer Info -->
          <div class="mt-6 p-4 bg-amber-500/10 border border-amber-500/30 rounded-xl">
            <p class="text-amber-300 text-xs text-center">
              ⏱️ صلاحية الرمز: 10 دقائق
              <br/>لم تصلك الرسالة؟ تحقق من مجلد البريد العشوائي (Spam)
            </p>
          </div>
        </div>
      </div>

      <!-- Footer -->
      <div class="mt-6 text-center">
        <p class="text-gray-500 text-xs">
          © {{ new Date().getFullYear() }} CSO ERP - منظمة LCO
        </p>
      </div>
    </div>
  </div>
</template>
