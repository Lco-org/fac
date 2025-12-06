<script setup>
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()

// Form state
const username = ref('')
const password = ref('')
const otpCode = ref('')
const error = ref('')
const successMessage = ref('')
const loading = ref(false)

// 2FA state from store
const requires2FA = computed(() => authStore.requires2FA)

// Handle initial login
const handleLogin = async () => {
  error.value = ''
  successMessage.value = ''
  loading.value = true

  const result = await authStore.login(username.value, password.value)

  loading.value = false

  if (result.success) {
    if (result.requires2FA) {
      successMessage.value = result.message
    } else {
      router.push('/')
    }
  } else {
    error.value = result.error
  }
}

// Handle 2FA verification
const handleVerify2FA = async () => {
  error.value = ''
  loading.value = true

  const result = await authStore.verify2FA(otpCode.value)

  loading.value = false

  if (result.success) {
    router.push('/')
  } else {
    error.value = result.error
  }
}

// Resend OTP code
const handleResendOTP = async () => {
  error.value = ''
  loading.value = true

  const result = await authStore.resendOTP()

  loading.value = false

  if (result.success) {
    successMessage.value = result.message
  } else {
    error.value = result.error
  }
}

// Cancel 2FA and go back to login
const handleCancel2FA = () => {
  authStore.cancel2FA()
  otpCode.value = ''
  error.value = ''
  successMessage.value = ''
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-gradient-to-br from-primary-600 to-primary-900">
    <div class="max-w-md w-full mx-4">
      <!-- Logo/Header -->
      <div class="text-center mb-8">
        <h1 class="text-4xl font-bold text-white mb-2">نظام ERP المالي</h1>
        <p class="text-primary-100">منظمات المجتمع المدني</p>
      </div>

      <!-- Login Card -->
      <div class="card">
        <!-- Step 1: Login Form -->
        <template v-if="!requires2FA">
          <h2 class="text-2xl font-bold text-center mb-6">تسجيل الدخول</h2>

          <form @submit.prevent="handleLogin" class="space-y-4">
            <!-- Error Message -->
            <div v-if="error" class="bg-deficit-light border border-deficit text-deficit-dark px-4 py-3 rounded-lg">
              {{ error }}
            </div>

            <!-- Username -->
            <div>
              <label for="username" class="block text-sm font-medium text-gray-700 mb-2">
                اسم المستخدم أو البريد الإلكتروني
              </label>
              <input
                id="username"
                v-model="username"
                type="text"
                required
                class="input-field"
                placeholder="admin"
                autocomplete="username"
              />
            </div>

            <!-- Password -->
            <div>
              <label for="password" class="block text-sm font-medium text-gray-700 mb-2">
                كلمة المرور
              </label>
              <input
                id="password"
                v-model="password"
                type="password"
                required
                class="input-field"
                placeholder="••••••••"
                autocomplete="current-password"
              />
            </div>

            <!-- Submit Button -->
            <button
              type="submit"
              :disabled="loading"
              class="w-full btn-primary disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <span v-if="loading">جاري تسجيل الدخول...</span>
              <span v-else>دخول</span>
            </button>
          </form>

          <!-- Security Notice -->
          <div class="mt-6 p-4 bg-blue-50 rounded-lg text-sm text-blue-700 border border-blue-200">
            <p class="font-semibold mb-1">🔐 نظام محمي بالتحقق الثنائي</p>
            <p>سيتم إرسال رمز تحقق إلى بريدك الإلكتروني عند تسجيل الدخول.</p>
          </div>
        </template>

        <!-- Step 2: 2FA Verification -->
        <template v-else>
          <h2 class="text-2xl font-bold text-center mb-2">التحقق الثنائي</h2>
          <p class="text-gray-500 text-center mb-6">أدخل رمز التحقق المرسل إلى بريدك الإلكتروني</p>

          <form @submit.prevent="handleVerify2FA" class="space-y-4">
            <!-- Success Message -->
            <div v-if="successMessage" class="bg-green-50 border border-green-200 text-green-700 px-4 py-3 rounded-lg">
              ✅ {{ successMessage }}
            </div>

            <!-- Error Message -->
            <div v-if="error" class="bg-deficit-light border border-deficit text-deficit-dark px-4 py-3 rounded-lg">
              {{ error }}
            </div>

            <!-- OTP Input -->
            <div>
              <label for="otp" class="block text-sm font-medium text-gray-700 mb-2">
                رمز التحقق (6 أرقام)
              </label>
              <input
                id="otp"
                v-model="otpCode"
                type="text"
                inputmode="numeric"
                pattern="[0-9]{6}"
                maxlength="6"
                required
                class="input-field text-center text-2xl tracking-widest font-mono"
                placeholder="000000"
                autocomplete="one-time-code"
              />
            </div>

            <!-- Submit Button -->
            <button
              type="submit"
              :disabled="loading || otpCode.length !== 6"
              class="w-full btn-primary disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <span v-if="loading">جاري التحقق...</span>
              <span v-else>تأكيد</span>
            </button>

            <!-- Resend & Cancel Buttons -->
            <div class="flex gap-4">
              <button
                type="button"
                @click="handleResendOTP"
                :disabled="loading"
                class="flex-1 btn-secondary text-sm"
              >
                إعادة إرسال الرمز
              </button>
              <button
                type="button"
                @click="handleCancel2FA"
                class="flex-1 bg-gray-100 hover:bg-gray-200 text-gray-700 font-medium py-2 px-4 rounded-lg transition-colors text-sm"
              >
                إلغاء
              </button>
            </div>
          </form>

          <!-- Timer Notice -->
          <div class="mt-4 p-3 bg-amber-50 rounded-lg text-sm text-amber-700 border border-amber-200 text-center">
            ⏱️ صلاحية الرمز: 10 دقائق
          </div>
        </template>
      </div>
    </div>
  </div>
</template>
