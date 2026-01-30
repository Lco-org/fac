<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()

// State
const loading = ref(false)
const checking = ref(true)
const registrationAvailable = ref(false)
const success = ref(false)
const error = ref(null)

const form = reactive({
  username: '',
  email: '',
  password: '',
  confirmPassword: '',
  fullName: ''
})

// API Base
const API_BASE = '/api/v1/system-auth'

// Check if registration is available
onMounted(async () => {
  try {
    const response = await fetch(`${API_BASE}/registration-status`, {
      credentials: 'include'
    })
    const data = await response.json()
    registrationAvailable.value = data.registration_available
  } catch (err) {
    console.error('Failed to check registration status:', err)
    registrationAvailable.value = false
  } finally {
    checking.value = false
  }
})

// Password strength
const passwordStrength = computed(() => {
  const pwd = form.password
  if (!pwd) return { level: 0, text: '', color: 'gray' }
  
  let score = 0
  if (pwd.length >= 8) score++
  if (pwd.length >= 12) score++
  if (/[A-Z]/.test(pwd)) score++
  if (/[a-z]/.test(pwd)) score++
  if (/[0-9]/.test(pwd)) score++
  if (/[^A-Za-z0-9]/.test(pwd)) score++

  if (score <= 2) return { level: 1, text: 'ضعيفة', color: 'red' }
  if (score <= 4) return { level: 2, text: 'متوسطة', color: 'yellow' }
  return { level: 3, text: 'قوية', color: 'green' }
})

// Password match
const passwordsMatch = computed(() => {
  return form.password === form.confirmPassword
})

// Form valid
const formValid = computed(() => {
  return form.username.length >= 3 &&
    form.email.includes('@') &&
    form.password.length >= 8 &&
    passwordsMatch.value &&
    form.fullName.length > 0
})

// Register
const register = async () => {
  if (!formValid.value) {
    error.value = 'يرجى ملء جميع الحقول بشكل صحيح'
    return
  }

  loading.value = true
  error.value = null

  try {
    const response = await fetch(`${API_BASE}/register`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify({
        username: form.username,
        email: form.email,
        password: form.password,
        full_name: form.fullName
      })
    })

    const data = await response.json()

    if (!response.ok) {
      throw new Error(data.error || 'فشل إنشاء الحساب')
    }

    success.value = true
    
    // Redirect to login after 3 seconds
    setTimeout(() => {
      router.push({ name: 'SSALogin' })
    }, 3000)

  } catch (err) {
    error.value = err.message
  } finally {
    loading.value = false
  }
}

// Go to login
const goToLogin = () => {
  router.push({ name: 'SSALogin' })
}
</script>

<template>
  <div class="min-h-screen bg-gradient-to-br from-slate-900 via-purple-900 to-slate-900 flex items-center justify-center p-4" dir="rtl">
    <!-- Background Effects -->
    <div class="absolute inset-0 overflow-hidden">
      <div class="absolute top-1/4 left-1/4 w-96 h-96 bg-purple-500/20 rounded-full blur-3xl"></div>
      <div class="absolute bottom-1/4 right-1/4 w-96 h-96 bg-blue-500/20 rounded-full blur-3xl"></div>
      <div class="absolute top-1/2 right-1/3 w-64 h-64 bg-emerald-500/10 rounded-full blur-2xl"></div>
    </div>

    <!-- Main Card -->
    <div class="relative w-full max-w-lg">
      <!-- Header -->
      <div class="text-center mb-8">
        <div class="inline-flex items-center justify-center w-20 h-20 bg-gradient-to-r from-emerald-600 to-teal-600 rounded-2xl shadow-2xl mb-4">
          <span class="text-4xl">🔐</span>
        </div>
        <h1 class="text-3xl font-bold text-white mb-2">إنشاء حساب المشرف الأعلى</h1>
        <p class="text-gray-400 text-sm">
          <span class="inline-block px-3 py-1 bg-emerald-500/20 rounded-full text-emerald-300 text-xs font-medium">
            SUPER SYSTEM ADMIN SETUP
          </span>
        </p>
      </div>

      <!-- Card -->
      <div class="bg-slate-800/80 backdrop-blur-xl rounded-3xl shadow-2xl border border-slate-700/50 overflow-hidden">
        <!-- Loading State -->
        <div v-if="checking" class="p-8 text-center">
          <div class="w-12 h-12 border-4 border-purple-500 border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
          <p class="text-gray-300">جاري التحقق من حالة النظام...</p>
        </div>

        <!-- Registration Not Available -->
        <div v-else-if="!registrationAvailable && !success" class="p-8 text-center">
          <div class="inline-flex items-center justify-center w-16 h-16 bg-amber-500/20 rounded-2xl mb-4">
            <span class="text-3xl">⚠️</span>
          </div>
          <h2 class="text-xl font-bold text-white mb-2">التسجيل غير متاح</h2>
          <p class="text-gray-400 mb-6">يوجد حساب مشرف أعلى بالفعل في النظام.</p>
          <button 
            @click="goToLogin"
            class="px-6 py-3 bg-gradient-to-r from-purple-600 to-indigo-600 text-white rounded-xl font-medium hover:shadow-purple-500/25 transition-all"
          >
            الانتقال لتسجيل الدخول
          </button>
        </div>

        <!-- Success State -->
        <div v-else-if="success" class="p-8 text-center">
          <div class="inline-flex items-center justify-center w-16 h-16 bg-emerald-500/20 rounded-2xl mb-4">
            <span class="text-3xl">✅</span>
          </div>
          <h2 class="text-xl font-bold text-white mb-2">تم إنشاء الحساب بنجاح!</h2>
          <p class="text-gray-400 mb-4">جاري توجيهك لصفحة تسجيل الدخول...</p>
          <div class="w-8 h-8 border-3 border-emerald-500 border-t-transparent rounded-full animate-spin mx-auto"></div>
        </div>

        <!-- Registration Form -->
        <div v-else class="p-8">
          <h2 class="text-xl font-bold text-white mb-6 text-center">إعداد النظام الأولي</h2>

          <!-- Error Alert -->
          <div v-if="error" class="mb-6 p-4 bg-red-500/20 border border-red-500/50 rounded-xl text-red-300 text-sm flex items-center gap-2">
            <span class="text-lg">⚠️</span>
            {{ error }}
          </div>

          <!-- Info Box -->
          <div class="mb-6 p-4 bg-blue-500/10 border border-blue-500/30 rounded-xl">
            <p class="text-blue-300 text-sm">
              ℹ️ هذه الصفحة متاحة للمرة الأولى فقط لإنشاء حساب المشرف الأعلى.
            </p>
          </div>

          <form @submit.prevent="register" class="space-y-5">
            <!-- Full Name -->
            <div>
              <label class="block text-sm font-medium text-gray-300 mb-2">الاسم الكامل</label>
              <div class="relative">
                <input 
                  v-model="form.fullName"
                  type="text"
                  required
                  class="w-full px-4 py-4 bg-slate-700/50 border border-slate-600 rounded-xl text-white placeholder-gray-500 focus:border-emerald-500 focus:ring-2 focus:ring-emerald-500/20 transition-all pr-12"
                  placeholder="مدير النظام الأعلى"
                />
                <span class="absolute left-4 top-1/2 -translate-y-1/2 text-gray-400">👤</span>
              </div>
            </div>

            <!-- Username -->
            <div>
              <label class="block text-sm font-medium text-gray-300 mb-2">اسم المستخدم</label>
              <div class="relative">
                <input 
                  v-model="form.username"
                  type="text"
                  required
                  autocomplete="username"
                  class="w-full px-4 py-4 bg-slate-700/50 border border-slate-600 rounded-xl text-white placeholder-gray-500 focus:border-emerald-500 focus:ring-2 focus:ring-emerald-500/20 transition-all pr-12"
                  placeholder="superadmin"
                  dir="ltr"
                />
                <span class="absolute left-4 top-1/2 -translate-y-1/2 text-gray-400">@</span>
              </div>
              <p v-if="form.username && form.username.length < 3" class="mt-1 text-xs text-red-400">
                يجب أن يكون 3 أحرف على الأقل
              </p>
            </div>

            <!-- Email -->
            <div>
              <label class="block text-sm font-medium text-gray-300 mb-2">البريد الإلكتروني</label>
              <div class="relative">
                <input 
                  v-model="form.email"
                  type="email"
                  required
                  autocomplete="email"
                  class="w-full px-4 py-4 bg-slate-700/50 border border-slate-600 rounded-xl text-white placeholder-gray-500 focus:border-emerald-500 focus:ring-2 focus:ring-emerald-500/20 transition-all pr-12"
                  placeholder="admin@example.com"
                  dir="ltr"
                />
                <span class="absolute left-4 top-1/2 -translate-y-1/2 text-gray-400">📧</span>
              </div>
            </div>

            <!-- Password -->
            <div>
              <label class="block text-sm font-medium text-gray-300 mb-2">كلمة المرور</label>
              <div class="relative">
                <input 
                  v-model="form.password"
                  type="password"
                  required
                  autocomplete="new-password"
                  class="w-full px-4 py-4 bg-slate-700/50 border border-slate-600 rounded-xl text-white placeholder-gray-500 focus:border-emerald-500 focus:ring-2 focus:ring-emerald-500/20 transition-all pr-12"
                  placeholder="••••••••••••"
                  dir="ltr"
                />
                <span class="absolute left-4 top-1/2 -translate-y-1/2 text-gray-400">🔒</span>
              </div>
              <!-- Password Strength -->
              <div v-if="form.password" class="mt-2">
                <div class="flex gap-1 mb-1">
                  <div 
                    v-for="i in 3" 
                    :key="i"
                    class="h-1 flex-1 rounded-full transition-colors"
                    :class="{
                      'bg-red-500': passwordStrength.level >= i && passwordStrength.color === 'red',
                      'bg-yellow-500': passwordStrength.level >= i && passwordStrength.color === 'yellow',
                      'bg-green-500': passwordStrength.level >= i && passwordStrength.color === 'green',
                      'bg-slate-600': passwordStrength.level < i
                    }"
                  ></div>
                </div>
                <p class="text-xs" :class="{
                  'text-red-400': passwordStrength.color === 'red',
                  'text-yellow-400': passwordStrength.color === 'yellow',
                  'text-green-400': passwordStrength.color === 'green'
                }">
                  قوة كلمة المرور: {{ passwordStrength.text }}
                </p>
              </div>
            </div>

            <!-- Confirm Password -->
            <div>
              <label class="block text-sm font-medium text-gray-300 mb-2">تأكيد كلمة المرور</label>
              <div class="relative">
                <input 
                  v-model="form.confirmPassword"
                  type="password"
                  required
                  autocomplete="new-password"
                  class="w-full px-4 py-4 bg-slate-700/50 border border-slate-600 rounded-xl text-white placeholder-gray-500 focus:border-emerald-500 focus:ring-2 focus:ring-emerald-500/20 transition-all pr-12"
                  placeholder="••••••••••••"
                  dir="ltr"
                />
                <span class="absolute left-4 top-1/2 -translate-y-1/2 text-gray-400">
                  {{ form.confirmPassword && passwordsMatch ? '✓' : '🔑' }}
                </span>
              </div>
              <p v-if="form.confirmPassword && !passwordsMatch" class="mt-1 text-xs text-red-400">
                كلمتا المرور غير متطابقتين
              </p>
            </div>

            <!-- Submit -->
            <button 
              type="submit"
              :disabled="loading || !formValid"
              class="w-full py-4 bg-gradient-to-r from-emerald-600 to-teal-600 text-white rounded-xl font-bold text-lg shadow-lg hover:shadow-emerald-500/25 transition-all disabled:opacity-50 flex items-center justify-center gap-2"
            >
              <span v-if="loading" class="w-5 h-5 border-2 border-white border-t-transparent rounded-full animate-spin"></span>
              <span v-else>
                🚀 إنشاء الحساب
              </span>
            </button>
          </form>

          <!-- Login Link -->
          <div class="mt-6 text-center">
            <button @click="goToLogin" class="text-purple-400 hover:text-purple-300 text-sm transition-colors">
              لديك حساب بالفعل؟ تسجيل الدخول
            </button>
          </div>

          <!-- Security Notice -->
          <div class="mt-6 p-4 bg-slate-700/30 rounded-xl border border-slate-600/50">
            <p class="text-xs text-gray-400 text-center">
              🔒 يتم تشفير كلمة المرور باستخدام Argon2id
              <br/>معيار التشفير العسكري الأكثر أمانًا
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
