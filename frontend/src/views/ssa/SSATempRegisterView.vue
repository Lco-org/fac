<script setup>
import { ref, reactive } from 'vue'

const loading = ref(false)
const result = ref(null)
const error = ref(null)

const form = reactive({
  username: 'superadmin',
  password: 'Admin@123456',
  email: 'admin@lco.org.ly',
  full_name: 'Super Admin'
})

const register = async () => {
  loading.value = true
  error.value = null
  result.value = null

  try {
    const response = await fetch('/api/v1/system-auth/temp-register', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(form)
    })

    const data = await response.json()

    if (!response.ok) {
      throw new Error(data.error || 'فشل التسجيل')
    }

    result.value = data
  } catch (err) {
    error.value = err.message
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen bg-gradient-to-br from-slate-900 via-purple-900 to-slate-900 flex items-center justify-center p-4" dir="rtl">
    <div class="bg-slate-800/80 backdrop-blur-xl rounded-3xl shadow-2xl border border-slate-700/50 p-8 w-full max-w-md">
      <div class="text-center mb-8">
        <div class="inline-flex items-center justify-center w-16 h-16 bg-gradient-to-r from-amber-500 to-orange-500 rounded-2xl mb-4">
          <span class="text-3xl">⚠️</span>
        </div>
        <h1 class="text-2xl font-bold text-white mb-2">تسجيل SSA مؤقت</h1>
        <p class="text-amber-300 text-sm">صفحة مؤقتة للاختبار فقط - احذفها في الإنتاج!</p>
      </div>

      <!-- Success -->
      <div v-if="result" class="mb-6 p-4 bg-green-500/20 border border-green-500/50 rounded-xl text-green-300">
        <p class="font-bold mb-2">✅ {{ result.message }}</p>
        <p class="text-sm">ID: {{ result.id }}</p>
        <p class="text-sm">Username: {{ result.username }}</p>
        <p class="text-sm mt-4">
          <a href="/b8fb17/ssa/login" class="text-green-400 underline">← اذهب لصفحة تسجيل الدخول</a>
        </p>
      </div>

      <!-- Error -->
      <div v-if="error" class="mb-6 p-4 bg-red-500/20 border border-red-500/50 rounded-xl text-red-300">
        ❌ {{ error }}
      </div>

      <form @submit.prevent="register" class="space-y-4">
        <div>
          <label class="block text-sm font-medium text-gray-300 mb-2">اسم المستخدم</label>
          <input 
            v-model="form.username"
            type="text"
            required
            class="w-full px-4 py-3 bg-slate-700/50 border border-slate-600 rounded-xl text-white"
            dir="ltr"
          />
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-300 mb-2">كلمة المرور</label>
          <input 
            v-model="form.password"
            type="text"
            required
            class="w-full px-4 py-3 bg-slate-700/50 border border-slate-600 rounded-xl text-white"
            dir="ltr"
          />
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-300 mb-2">البريد الإلكتروني</label>
          <input 
            v-model="form.email"
            type="email"
            required
            class="w-full px-4 py-3 bg-slate-700/50 border border-slate-600 rounded-xl text-white"
            dir="ltr"
          />
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-300 mb-2">الاسم الكامل</label>
          <input 
            v-model="form.full_name"
            type="text"
            class="w-full px-4 py-3 bg-slate-700/50 border border-slate-600 rounded-xl text-white"
          />
        </div>

        <button 
          type="submit"
          :disabled="loading"
          class="w-full py-4 bg-gradient-to-r from-amber-500 to-orange-500 text-white rounded-xl font-bold text-lg shadow-lg hover:shadow-orange-500/25 transition-all disabled:opacity-50"
        >
          <span v-if="loading">جاري التسجيل...</span>
          <span v-else>📝 إنشاء/تحديث الحساب</span>
        </button>
      </form>

      <div class="mt-6 text-center">
        <a href="/b8fb17/ssa/login" class="text-purple-400 text-sm hover:text-purple-300">
          ← العودة لتسجيل الدخول
        </a>
      </div>
    </div>
  </div>
</template>
