<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import ssa from '@/api/ssa'

const router = useRouter()

// Form State
const form = reactive({
  name: '',
  name_ar: '',
  domain: '',
  contact_email: '',
  contact_phone: '',
  subscription_plan_code: '',
  subscription_start: new Date().toISOString().split('T')[0],
  subscription_end: ''
})

const plans = ref([])

const loading = ref(false)
const error = ref(null)
const step = ref(1)
const createdOrg = ref(null)

// API Base URL
const API_BASE = '/api/v1/ssa'

// Validate domain format - supports .com, .com.ly, .org.ly, etc.
const isValidDomain = (domain) => {
  const pattern = /^[a-zA-Z0-9][a-zA-Z0-9-]*(\.[a-zA-Z]{2,})+$/
  return pattern.test(domain)
}

// Validate email format
const isValidEmail = (email) => {
  const pattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  return pattern.test(email)
}

// Form validation - computed for reactivity
const isFormValid = computed(() => {
  return (
    form.name.length >= 2 &&
    form.domain.length > 0 && isValidDomain(form.domain) &&
    form.contact_email.length > 0 && isValidEmail(form.contact_email) &&
    form.subscription_plan_code.length > 0
  )
})

const fetchPlans = async () => {
  try {
    const response = await ssa.organizations.getPlans()
    plans.value = response.data.plans || []
    // Set default if exists
    const defaultPlan = plans.value.find(p => p.is_default)
    if (defaultPlan) {
      form.subscription_plan_code = defaultPlan.code
    }
  } catch (err) {
    console.error('Failed to fetch plans', err)
  }
}

onMounted(() => {
  fetchPlans()
})

// Submit form
const submitForm = async () => {
  if (!isFormValid.value) return
  
  loading.value = true
  error.value = null
  
  try {
    const token = localStorage.getItem('ssa_token')
    const response = await fetch(`${API_BASE}/organizations`, {
      method: 'POST',
      headers: { 
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      credentials: 'include',
      body: JSON.stringify({
        name: form.name,
        name_ar: form.name_ar || null,
        domain: form.domain,
        contact_email: form.contact_email,
        contact_phone: form.contact_phone || null,
        subscription_plan_code: form.subscription_plan_code,
        subscription_start: form.subscription_start ? new Date(form.subscription_start).toISOString() : null,
        subscription_end: form.subscription_end ? new Date(form.subscription_end).toISOString() : null
      })
    })
    
    if (!response.ok) {
      const data = await response.json()
      throw new Error(data.error || 'Failed to create organization')
    }
    
    const data = await response.json()
    createdOrg.value = data.organization
    step.value = 2 // Move to verification step
  } catch (err) {
    error.value = err.message
  } finally {
    loading.value = false
  }
}

// Go back to dashboard
const goBack = () => {
  router.push({ name: 'SSADashboard' })
}

// Go to verification page
const goToVerification = () => {
  if (createdOrg.value) {
    router.push({ 
      name: 'SSAVerifyDomain', 
      params: { id: createdOrg.value.id },
      query: { email: form.contact_email }
    })
  }
}
</script>

<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-900 p-6" dir="rtl">
    <div class="max-w-3xl mx-auto">
      <!-- Header -->
      <div class="mb-8">
        <button 
          @click="goBack"
          class="flex items-center gap-2 text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300 mb-4 transition-colors"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18"/>
          </svg>
          العودة للوحة التحكم
        </button>
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white flex items-center gap-3">
          <span class="text-4xl">🏢</span>
          إنشاء منظمة جديدة
        </h1>
        <p class="mt-2 text-gray-500 dark:text-gray-400">
          أدخل بيانات المنظمة الجديدة لإنشاء نظام ERP مخصص
        </p>
      </div>

      <!-- Progress Steps -->
      <div class="flex items-center justify-center gap-4 mb-8">
        <div class="flex items-center">
          <div :class="['w-10 h-10 rounded-full flex items-center justify-center font-bold text-white', step >= 1 ? 'bg-blue-600' : 'bg-gray-300']">1</div>
          <span class="mr-2 font-medium" :class="step >= 1 ? 'text-blue-600 dark:text-blue-400' : 'text-gray-400'">البيانات</span>
        </div>
        <div class="w-16 h-1 bg-gray-200 dark:bg-gray-700 rounded">
          <div class="h-full bg-blue-600 rounded transition-all duration-300" :style="{ width: step >= 2 ? '100%' : '0%' }"></div>
        </div>
        <div class="flex items-center">
          <div :class="['w-10 h-10 rounded-full flex items-center justify-center font-bold', step >= 2 ? 'bg-blue-600 text-white' : 'bg-gray-200 dark:bg-gray-700 text-gray-400']">2</div>
          <span class="mr-2 font-medium" :class="step >= 2 ? 'text-blue-600 dark:text-blue-400' : 'text-gray-400'">التحقق</span>
        </div>
        <div class="w-16 h-1 bg-gray-200 dark:bg-gray-700 rounded">
          <div class="h-full bg-blue-600 rounded transition-all duration-300" :style="{ width: step >= 3 ? '100%' : '0%' }"></div>
        </div>
        <div class="flex items-center">
          <div :class="['w-10 h-10 rounded-full flex items-center justify-center font-bold', step >= 3 ? 'bg-blue-600 text-white' : 'bg-gray-200 dark:bg-gray-700 text-gray-400']">3</div>
          <span class="mr-2 font-medium" :class="step >= 3 ? 'text-blue-600 dark:text-blue-400' : 'text-gray-400'">الإكمال</span>
        </div>
      </div>

      <!-- Step 1: Form -->
      <div v-if="step === 1" class="bg-white dark:bg-gray-800 rounded-2xl shadow-lg p-8 border border-gray-100 dark:border-gray-700">
        <!-- Error Alert -->
        <div v-if="error" class="mb-6 p-4 bg-red-50 dark:bg-red-900/30 border border-red-200 dark:border-red-800 rounded-xl">
          <p class="text-red-600 dark:text-red-400 flex items-center gap-2">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
            </svg>
            {{ error }}
          </p>
        </div>

        <form @submit.prevent="submitForm">
          <!-- Organization Name -->
          <div class="mb-6">
            <label class="block text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">
              اسم المنظمة <span class="text-red-500">*</span>
            </label>
            <input 
              v-model="form.name"
              type="text" 
              required
              placeholder="مثال: شركة الأفق للتقنية"
              class="w-full px-4 py-3 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all"
            />
          </div>

          <!-- Arabic Name (Optional) -->
          <div class="mb-6">
            <label class="block text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">
              الاسم بالعربية (اختياري)
            </label>
            <input 
              v-model="form.name_ar"
              type="text" 
              placeholder="مثال: شركة الأفق للتقنية"
              class="w-full px-4 py-3 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all"
              dir="rtl"
            />
          </div>

          <!-- Domain -->
          <div class="mb-6">
            <label class="block text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">
              النطاق (الدومين) <span class="text-red-500">*</span>
            </label>
            <input 
              v-model="form.domain"
              type="text" 
              required
              placeholder="مثال: company.com"
              class="w-full px-4 py-3 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all font-mono"
              dir="ltr"
            />
            <p class="mt-2 text-sm text-gray-500 dark:text-gray-400">
              سيتم استخدام هذا النطاق للوصول إلى النظام
            </p>
          </div>

          <!-- Contact Email -->
          <div class="mb-6">
            <label class="block text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">
              البريد الإلكتروني للتواصل <span class="text-red-500">*</span>
            </label>
            <input 
              v-model="form.contact_email"
              type="email" 
              required
              placeholder="admin@company.com"
              class="w-full px-4 py-3 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all"
              dir="ltr"
            />
            <p class="mt-2 text-sm text-gray-500 dark:text-gray-400">
              سيتم إرسال رمز التحقق إلى هذا البريد
            </p>
          </div>

          <!-- Contact Phone -->
          <div class="mb-8">
            <label class="block text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">
              رقم الهاتف (اختياري)
            </label>
            <input 
              v-model="form.contact_phone"
              type="tel" 
              placeholder="+218 XX XXX XXXX"
              class="w-full px-4 py-3 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all"
              dir="ltr"
            />
          </div>

          <!-- Subscription Plan -->
          <div class="mb-6">
            <label class="block text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">
              خطة الاشتراك <span class="text-red-500">*</span>
            </label>
            <select 
              v-model="form.subscription_plan_code"
              class="w-full px-4 py-3 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all"
            >
              <option value="" disabled>اختر خطة الاشتراك</option>
              <option v-for="plan in plans" :key="plan.code" :value="plan.code">
                {{ plan.name_ar || plan.name }} - {{ plan.price_monthly }} {{ plan.currency }} / شهرياً
              </option>
            </select>
          </div>

          <!-- Dates -->
          <div class="grid grid-cols-2 gap-4 mb-8">
            <div>
              <label class="block text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">
                تاريخ البدء
              </label>
              <input 
                v-model="form.subscription_start"
                type="date" 
                class="w-full px-4 py-3 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all"
              />
            </div>
            <div>
              <label class="block text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">
                تاريخ الانتهاء
              </label>
              <input 
                v-model="form.subscription_end"
                type="date" 
                class="w-full px-4 py-3 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all"
              />
            </div>
          </div>

          <!-- Submit Button -->
          <button 
            type="submit"
            :disabled="loading || !isFormValid"
            class="w-full py-4 bg-gradient-to-r from-blue-600 to-indigo-600 text-white rounded-xl font-bold text-lg shadow-lg hover:shadow-xl transition-all disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-3"
          >
            <span v-if="loading" class="w-6 h-6 border-2 border-white border-t-transparent rounded-full animate-spin"></span>
            <span v-else>
              <svg class="w-6 h-6 inline-block ml-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 9l3 3m0 0l-3 3m3-3H8m13 0a9 9 0 11-18 0 9 9 0 0118 0z"/>
              </svg>
              إنشاء المنظمة وإرسال رمز التحقق
            </span>
          </button>
        </form>
      </div>

      <!-- Step 2: Success & Verification Notice -->
      <div v-else-if="step === 2" class="bg-white dark:bg-gray-800 rounded-2xl shadow-lg p-8 border border-gray-100 dark:border-gray-700 text-center">
        <div class="w-24 h-24 mx-auto mb-6 bg-green-100 dark:bg-green-900/50 rounded-full flex items-center justify-center">
          <span class="text-5xl">✅</span>
        </div>
        
        <h2 class="text-2xl font-bold text-gray-900 dark:text-white mb-4">
          تم إنشاء المنظمة بنجاح!
        </h2>
        
        <p class="text-gray-600 dark:text-gray-400 mb-6">
          تم إرسال رمز التحقق إلى البريد الإلكتروني:
          <br/>
          <code class="text-lg font-mono text-blue-600 dark:text-blue-400">{{ form.contact_email }}</code>
        </p>
        
        <div class="bg-amber-50 dark:bg-amber-900/30 border border-amber-200 dark:border-amber-800 rounded-xl p-4 mb-6 text-right">
          <p class="text-amber-700 dark:text-amber-400 text-sm">
            ⏱️ <strong>ملاحظة:</strong> صلاحية رمز التحقق 5 دقائق فقط. يمكنك إعادة إرساله إذا انتهت الصلاحية.
          </p>
        </div>
        
        <button 
          @click="goToVerification"
          class="w-full py-4 bg-gradient-to-r from-emerald-500 to-green-600 text-white rounded-xl font-bold text-lg shadow-lg hover:shadow-xl transition-all flex items-center justify-center gap-3"
        >
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
          </svg>
          متابعة إلى صفحة التحقق
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Custom focus styles */
input:focus {
  outline: none;
}
</style>

