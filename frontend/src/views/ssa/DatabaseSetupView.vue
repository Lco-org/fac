<script setup>
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'

const router = useRouter()
const route = useRoute()

// State
const orgId = ref(route.params.id)
const loading = ref(false)
const creating = ref(false)
const error = ref(null)
const success = ref(false)
const organization = ref(null)

// API Base URL
const API_BASE = '/api/v1/ssa'

// Fetch organization details
const fetchOrganization = async () => {
  loading.value = true
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
  } finally {
    loading.value = false
  }
}

// Create database
const createDatabase = async () => {
  creating.value = true
  error.value = null
  
  try {
    const token = localStorage.getItem('ssa_token')
    const response = await fetch(`${API_BASE}/organizations/${orgId.value}/database`, {
      method: 'POST',
      credentials: 'include',
      headers: { 'Authorization': `Bearer ${token}` }
    })
    
    if (!response.ok) {
      const data = await response.json()
      throw new Error(data.error || 'فشل إنشاء قاعدة البيانات')
    }
    
    success.value = true
  } catch (err) {
    error.value = err.message
  } finally {
    creating.value = false
  }
}

// Navigate to create admin
const goToCreateAdmin = () => {
  router.push({ 
    name: 'SSACreateAdmin', 
    params: { id: orgId.value }
  })
}

// Go back
const goBack = () => {
  router.push({ name: 'SSADashboard' })
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
          <span class="text-4xl">🗄️</span>
          إنشاء قاعدة البيانات
        </h1>
        <p class="mt-2 text-gray-500 dark:text-gray-400">
          إنشاء قاعدة بيانات خاصة للمنظمة
        </p>
      </div>

      <!-- Loading State -->
      <div v-if="loading" class="flex items-center justify-center p-12">
        <div class="w-12 h-12 border-4 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
      </div>

      <!-- Success State -->
      <div v-else-if="success" class="bg-white dark:bg-gray-800 rounded-2xl shadow-lg p-8 border border-gray-100 dark:border-gray-700 text-center">
        <div class="w-24 h-24 mx-auto mb-6 bg-green-100 dark:bg-green-900/50 rounded-full flex items-center justify-center">
          <span class="text-5xl">✅</span>
        </div>
        
        <h2 class="text-2xl font-bold text-gray-900 dark:text-white mb-4">
          تم إنشاء قاعدة البيانات بنجاح!
        </h2>
        
        <div class="bg-gray-50 dark:bg-gray-700 rounded-xl p-4 mb-6">
          <p class="text-gray-600 dark:text-gray-400 text-sm">
            <strong>اسم قاعدة البيانات:</strong>
            <br/>
            <code class="text-lg font-mono text-blue-600 dark:text-blue-400">
              {{ organization?.database_name }}
            </code>
          </p>
        </div>
        
        <p class="text-gray-600 dark:text-gray-400 mb-6">
          الآن يمكنك إنشاء أول مدير نظام للمنظمة.
        </p>
        
        <button 
          @click="goToCreateAdmin"
          class="w-full py-4 bg-gradient-to-r from-purple-500 to-indigo-600 text-white rounded-xl font-bold text-lg shadow-lg hover:shadow-xl transition-all flex items-center justify-center gap-3"
        >
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M18 9v3m0 0v3m0-3h3m-3 0h-3m-2-5a4 4 0 11-8 0 4 4 0 018 0zM3 20a6 6 0 0112 0v1H3v-1z"/>
          </svg>
          متابعة إلى إنشاء المدير
        </button>
      </div>

      <!-- Main Content -->
      <div v-else class="bg-white dark:bg-gray-800 rounded-2xl shadow-lg p-8 border border-gray-100 dark:border-gray-700">
        <!-- Organization Info -->
        <div v-if="organization" class="mb-8">
          <div class="bg-gradient-to-r from-blue-50 to-indigo-50 dark:from-blue-900/30 dark:to-indigo-900/30 rounded-xl p-6 border border-blue-100 dark:border-blue-800">
            <div class="flex items-center gap-4 mb-4">
              <div class="w-14 h-14 bg-gradient-to-br from-blue-500 to-purple-600 rounded-xl flex items-center justify-center text-white font-bold text-2xl">
                {{ organization.name.charAt(0) }}
              </div>
              <div>
                <h3 class="text-xl font-bold text-gray-900 dark:text-white">{{ organization.name }}</h3>
                <p class="text-gray-500 dark:text-gray-400">{{ organization.primary_domain }}</p>
              </div>
            </div>
            <div class="grid grid-cols-2 gap-4 text-sm">
              <div>
                <span class="text-gray-500 dark:text-gray-400">البريد:</span>
                <span class="text-gray-900 dark:text-white mr-2">{{ organization.contact_email }}</span>
              </div>
              <div>
                <span class="text-gray-500 dark:text-gray-400">الحالة:</span>
                <span class="text-green-600 dark:text-green-400 mr-2 font-medium">نشط ✓</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Database Info -->
        <div class="mb-8">
          <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4 flex items-center gap-2">
            <span>📋</span>
            تفاصيل قاعدة البيانات
          </h3>
          <div class="space-y-4">
            <div class="flex justify-between items-center p-4 bg-gray-50 dark:bg-gray-700 rounded-xl">
              <span class="text-gray-600 dark:text-gray-400">اسم قاعدة البيانات:</span>
              <code class="font-mono text-blue-600 dark:text-blue-400 bg-blue-50 dark:bg-blue-900/30 px-3 py-1 rounded-lg">
                {{ organization?.database_name || 'جاري التحميل...' }}
              </code>
            </div>
            <div class="flex justify-between items-center p-4 bg-gray-50 dark:bg-gray-700 rounded-xl">
              <span class="text-gray-600 dark:text-gray-400">قاعدة البيانات الأساس:</span>
              <code class="font-mono text-gray-600 dark:text-gray-400">cso_erp (Template)</code>
            </div>
          </div>
        </div>

        <!-- Warning -->
        <div class="mb-8 p-4 bg-amber-50 dark:bg-amber-900/30 border border-amber-200 dark:border-amber-800 rounded-xl">
          <p class="text-amber-700 dark:text-amber-400 text-sm flex items-start gap-2">
            <span class="text-lg">⚠️</span>
            <span>
              <strong>ملاحظة:</strong> سيتم نسخ هيكل قاعدة البيانات الأساسية (cso_erp) بالكامل. هذه العملية قد تستغرق بضع ثوانٍ.
            </span>
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

        <!-- Create Button -->
        <button 
          @click="createDatabase"
          :disabled="creating"
          class="w-full py-4 bg-gradient-to-r from-emerald-500 to-green-600 text-white rounded-xl font-bold text-lg shadow-lg hover:shadow-xl transition-all disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-3"
        >
          <span v-if="creating" class="w-6 h-6 border-2 border-white border-t-transparent rounded-full animate-spin"></span>
          <span v-else>
            <svg class="w-6 h-6 inline-block ml-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4"/>
            </svg>
            إنشاء قاعدة البيانات
          </span>
        </button>
      </div>
    </div>
  </div>
</template>
