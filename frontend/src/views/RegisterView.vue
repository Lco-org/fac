<script setup>
import { ref } from 'vue'
import api from '@/api/client'
import { useRouter } from 'vue-router'

const router = useRouter()
const loading = ref(false)
const success = ref(false)
const error = ref('')

const form = ref({
  username: '',
  email: '',
  password: '',
  full_name: '',
  full_name_ar: '',
  role: 'admin'
})

const handleSubmit = async () => {
  error.value = ''
  success.value = false
  loading.value = true

  try {
    const response = await api.post('/auth/register', form.value)
    success.value = true
    
    setTimeout(() => {
      router.push('/login')
    }, 2000)
  } catch (err) {
    error.value = err.response?.data?.error || 'فشل في إنشاء الحساب'
  } finally {
    loading.value = false
  }
}

const goToLogin = () => {
  router.push('/login')
}
</script>

<template>
  <div class="min-h-screen bg-gradient-to-br from-blue-600 to-blue-800 flex items-center justify-center p-4">
    <div class="card max-w-md w-full">
      <div class="text-center mb-6">
        <h1 class="text-3xl font-bold text-gray-800 mb-2">إنشاء حساب جديد</h1>
        <p class="text-gray-600">نظام ERP المالي - منظمات المجتمع المدني</p>
      </div>

      <div v-if="success" class="bg-surplus-light border border-surplus text-surplus-dark px-4 py-3 rounded-lg mb-4">
        ✅ تم إنشاء الحساب بنجاح! جاري التحويل لصفحة تسجيل الدخول...
      </div>

      <div v-if="error" class="bg-deficit-light border border-deficit text-deficit-dark px-4 py-3 rounded-lg mb-4">
        ❌ {{ error }}
      </div>

      <form @submit.prevent="handleSubmit" class="space-y-4">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">
            اسم المستخدم <span class="text-deficit">*</span>
          </label>
          <input 
            v-model="form.username" 
            type="text" 
            required 
            minlength="3"
            class="input-field" 
            placeholder="admin" 
          />
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">
            البريد الإلكتروني <span class="text-deficit">*</span>
          </label>
          <input 
            v-model="form.email" 
            type="email" 
            required 
            class="input-field" 
            placeholder="admin@example.com" 
          />
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">
            كلمة المرور <span class="text-deficit">*</span>
          </label>
          <input 
            v-model="form.password" 
            type="password" 
            required 
            minlength="6"
            class="input-field" 
            placeholder="••••••" 
          />
          <p class="text-xs text-gray-500 mt-1">يجب أن تكون 6 أحرف على الأقل</p>
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">
            الاسم الكامل <span class="text-deficit">*</span>
          </label>
          <input 
            v-model="form.full_name" 
            type="text" 
            required 
            class="input-field" 
            placeholder="Administrator" 
          />
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">
            الاسم بالعربية
          </label>
          <input 
            v-model="form.full_name_ar" 
            type="text" 
            class="input-field" 
            placeholder="المسؤول" 
          />
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">
            الصلاحية <span class="text-deficit">*</span>
          </label>
          <select v-model="form.role" required class="input-field">
            <option value="admin">مدير (Admin)</option>
            <option value="finance">مالي (Finance)</option>
            <option value="auditor">مراجع (Auditor)</option>
          </select>
        </div>

        <div class="flex gap-3">
          <button 
            type="submit" 
            :disabled="loading" 
            class="btn-primary flex-1 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <span v-if="loading">جاري الإنشاء...</span>
            <span v-else>إنشاء الحساب</span>
          </button>
          <button 
            type="button" 
            @click="goToLogin" 
            class="btn-secondary"
          >
            تسجيل الدخول
          </button>
        </div>
      </form>

      <div class="mt-6 p-4 bg-blue-50 rounded-lg">
        <p class="text-sm text-blue-800 font-medium mb-2">💡 ملاحظة:</p>
        <p class="text-xs text-blue-700">
          هذه صفحة مؤقتة لإنشاء حساب جديد. بعد إنشاء الحساب، يمكنك تسجيل الدخول مباشرة.
        </p>
      </div>
    </div>
  </div>
</template>
