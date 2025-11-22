<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const username = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)

const handleLogin = async () => {
  error.value = ''
  loading.value = true

  const result = await authStore.login(username.value, password.value)

  loading.value = false

  if (result.success) {
    router.push('/')
  } else {
    error.value = result.error
  }
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

        <!-- Default Credentials Info -->
        <div class="mt-6 p-4 bg-gray-50 rounded-lg text-sm text-gray-600">
          <p class="font-semibold mb-1">بيانات الدخول الافتراضية:</p>
          <p>المستخدم: <code class="bg-gray-200 px-2 py-1 rounded">admin</code></p>
          <p>كلمة المرور: <code class="bg-gray-200 px-2 py-1 rounded">Admin@123</code></p>
          <p class="text-deficit text-xs mt-2">⚠️ يجب تغيير كلمة المرور بعد أول تسجيل دخول</p>
        </div>
      </div>
    </div>
  </div>
</template>
