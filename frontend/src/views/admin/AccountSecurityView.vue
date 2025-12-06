<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()

// State
const loading = ref(false)
const error = ref('')
const success = ref('')

// Password change form
const showPasswordForm = ref(false)
const oldPassword = ref('')
const newPassword = ref('')
const confirmPassword = ref('')

// 2FA state
const twoFactorEnabled = ref(false)

onMounted(() => {
  if (authStore.user) {
    twoFactorEnabled.value = authStore.user.two_factor_enabled
  }
})

// Toggle 2FA
const handleToggle2FA = async () => {
  error.value = ''
  success.value = ''
  loading.value = true

  const result = await authStore.toggle2FA(!twoFactorEnabled.value)
  
  loading.value = false

  if (result.success) {
    twoFactorEnabled.value = !twoFactorEnabled.value
    success.value = result.message
  } else {
    error.value = result.error
  }
}

// Change password
const handleChangePassword = async () => {
  error.value = ''
  success.value = ''

  if (newPassword.value !== confirmPassword.value) {
    error.value = 'كلمة المرور الجديدة غير متطابقة'
    return
  }

  if (newPassword.value.length < 8) {
    error.value = 'كلمة المرور يجب أن تكون 8 أحرف على الأقل'
    return
  }

  loading.value = true

  const result = await authStore.changePassword(oldPassword.value, newPassword.value)
  
  loading.value = false

  if (result.success) {
    success.value = result.message
    setTimeout(() => {
      router.push('/login')
    }, 2000)
  } else {
    error.value = result.error
  }
}

// Revoke all sessions
const handleRevokeAllSessions = async () => {
  if (!confirm('هل أنت متأكد من إلغاء جميع الجلسات؟ سيتم تسجيل خروجك من جميع الأجهزة.')) {
    return
  }

  error.value = ''
  success.value = ''
  loading.value = true

  const result = await authStore.revokeAllSessions()
  
  loading.value = false

  if (result.success) {
    router.push('/login')
  } else {
    error.value = result.error
  }
}
</script>

<template>
  <div class="max-w-2xl mx-auto">
    <h1 class="text-2xl font-bold mb-6">إعدادات الأمان</h1>

    <!-- Messages -->
    <div v-if="error" class="mb-4 bg-deficit-light border border-deficit text-deficit-dark px-4 py-3 rounded-lg">
      {{ error }}
    </div>
    <div v-if="success" class="mb-4 bg-green-50 border border-green-200 text-green-700 px-4 py-3 rounded-lg">
      ✅ {{ success }}
    </div>

    <!-- Two-Factor Authentication -->
    <div class="card mb-6">
      <h2 class="text-lg font-semibold mb-4 flex items-center gap-2">
        <span class="text-2xl">🔐</span>
        التحقق الثنائي (2FA)
      </h2>
      <p class="text-gray-600 mb-4">
        عند تفعيل التحقق الثنائي، سيتم إرسال رمز تحقق إلى بريدك الإلكتروني في كل مرة تقوم فيها بتسجيل الدخول.
      </p>
      
      <div class="flex items-center justify-between p-4 bg-gray-50 rounded-lg">
        <div>
          <span class="font-medium">حالة التحقق الثنائي:</span>
          <span :class="twoFactorEnabled ? 'text-green-600' : 'text-gray-500'" class="mr-2">
            {{ twoFactorEnabled ? '🟢 مفعّل' : '⚪ معطّل' }}
          </span>
        </div>
        <button
          @click="handleToggle2FA"
          :disabled="loading"
          :class="twoFactorEnabled ? 'bg-gray-500 hover:bg-gray-600' : 'bg-green-500 hover:bg-green-600'"
          class="text-white font-medium py-2 px-4 rounded-lg transition-colors disabled:opacity-50"
        >
          {{ twoFactorEnabled ? 'تعطيل' : 'تفعيل' }}
        </button>
      </div>
    </div>

    <!-- Change Password -->
    <div class="card mb-6">
      <h2 class="text-lg font-semibold mb-4 flex items-center gap-2">
        <span class="text-2xl">🔑</span>
        تغيير كلمة المرور
      </h2>
      
      <button 
        v-if="!showPasswordForm"
        @click="showPasswordForm = true"
        class="btn-secondary"
      >
        تغيير كلمة المرور
      </button>

      <form v-else @submit.prevent="handleChangePassword" class="space-y-4">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">
            كلمة المرور الحالية
          </label>
          <input
            v-model="oldPassword"
            type="password"
            required
            class="input-field"
            placeholder="••••••••"
          />
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">
            كلمة المرور الجديدة
          </label>
          <input
            v-model="newPassword"
            type="password"
            required
            minlength="8"
            class="input-field"
            placeholder="8 أحرف على الأقل"
          />
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">
            تأكيد كلمة المرور الجديدة
          </label>
          <input
            v-model="confirmPassword"
            type="password"
            required
            class="input-field"
            placeholder="أعد كتابة كلمة المرور"
          />
        </div>

        <div class="flex gap-4">
          <button
            type="submit"
            :disabled="loading"
            class="btn-primary disabled:opacity-50"
          >
            {{ loading ? 'جاري الحفظ...' : 'حفظ' }}
          </button>
          <button
            type="button"
            @click="showPasswordForm = false"
            class="btn-secondary"
          >
            إلغاء
          </button>
        </div>
      </form>
    </div>

    <!-- Kill Switch -->
    <div class="card border-2 border-deficit">
      <h2 class="text-lg font-semibold mb-4 flex items-center gap-2 text-deficit-dark">
        <span class="text-2xl">⚠️</span>
        إلغاء جميع الجلسات
      </h2>
      <p class="text-gray-600 mb-4">
        سيتم تسجيل خروجك من جميع الأجهزة المتصلة بحسابك. استخدم هذا الخيار إذا كنت تشك في اختراق حسابك.
      </p>
      
      <button
        @click="handleRevokeAllSessions"
        :disabled="loading"
        class="bg-deficit hover:bg-deficit-dark text-white font-medium py-2 px-4 rounded-lg transition-colors disabled:opacity-50"
      >
        {{ loading ? 'جاري الإلغاء...' : 'إلغاء جميع الجلسات' }}
      </button>
    </div>
  </div>
</template>
