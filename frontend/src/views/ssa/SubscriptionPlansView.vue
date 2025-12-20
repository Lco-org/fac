<script setup>
import { ref, onMounted, reactive } from 'vue'
import ssa from '@/api/ssa'

// State
const loading = ref(true)
const plans = ref([])
const error = ref(null)

// Modal State
const showModal = ref(false)
const modalMode = ref('create') // 'create' or 'edit'
const saving = ref(false)
const modalError = ref(null)

// Form Data
const form = reactive({
  id: null,
  name: '',
  name_ar: '',
  code: '',
  price_monthly: 0,
  price_yearly: 0,
  currency: 'LYD',
  max_users: 5,
  max_storage_gb: 1,
  features: '', // Comma separated for simplicity
  is_active: true,
  is_default: false,
  sort_order: 0,
  description: '',
  description_ar: ''
})

// Fetch Plans
const fetchPlans = async () => {
  loading.value = true
  error.value = null
  try {
    const response = await ssa.plans.list()
    // Backend returns { plans: [...] } or just [...] depending on implementation?
    // Handler returns json(fiber.Map{"plans": plans})
    plans.value = response.data.plans || []
  } catch (err) {
    error.value = 'فشل تحميل خطط الاشتراك'
    console.error(err)
  } finally {
    loading.value = false
  }
}

// Open Modal
const openCreateModal = () => {
  modalMode.value = 'create'
  resetForm()
  showModal.value = true
}

const openEditModal = (plan) => {
  modalMode.value = 'edit'
  form.id = plan.id
  form.name = plan.name
  form.name_ar = plan.name_ar || ''
  form.code = plan.code
  form.price_monthly = plan.price_monthly
  form.price_yearly = plan.price_yearly
  form.currency = plan.currency
  form.max_users = plan.max_users
  form.max_storage_gb = plan.max_storage_gb
  form.features = Array.isArray(plan.features) ? plan.features.join(', ') : ''
  form.is_active = plan.is_active
  form.is_default = plan.is_default
  form.sort_order = plan.sort_order
  form.description = plan.description || ''
  form.description_ar = plan.description_ar || ''
  showModal.value = true
}

const closeModal = () => {
  showModal.value = false
  resetForm()
}

const resetForm = () => {
  form.id = null
  form.name = ''
  form.name_ar = ''
  form.code = ''
  form.price_monthly = 0
  form.price_yearly = 0
  form.currency = 'LYD'
  form.max_users = 5
  form.max_storage_gb = 1
  form.features = ''
  form.is_active = true
  form.is_default = false
  form.sort_order = 0
  form.description = ''
  form.description_ar = ''
  modalError.value = null
}

// Submit Form
const submitForm = async () => {
  saving.value = true
  modalError.value = null
  
  try {
    const payload = { ...form }
    // Convert features string to array
    payload.features = form.features.split(',').map(f => f.trim()).filter(f => f.length > 0)
    
    // Convert numeric fields
    payload.price_monthly = parseFloat(payload.price_monthly)
    payload.price_yearly = parseFloat(payload.price_yearly)
    payload.max_users = parseInt(payload.max_users)
    payload.max_storage_gb = parseInt(payload.max_storage_gb)
    
    // Remove ID from payload for create
    if (modalMode.value === 'create') {
      delete payload.id
      await ssa.plans.create(payload)
    } else {
      await ssa.plans.update(form.id, payload)
    }
    
    closeModal()
    fetchPlans()
  } catch (err) {
    modalError.value = err.response?.data?.error || err.message || 'حدث خطأ أثناء الحفظ'
  } finally {
    saving.value = false
  }
}

// Delete Plan
const deletePlan = async (id) => {
  if (!confirm('هل أنت متأكد من حذف هذه الخطة؟ لا يمكن التراجع عن هذا الإجراء.')) return
  
  try {
    await ssa.plans.delete(id)
    fetchPlans()
  } catch (err) {
    alert(err.response?.data?.error || 'فشل حذف الخطة')
  }
}

onMounted(() => {
  fetchPlans()
})
</script>

<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-900 p-6" dir="rtl">
    <div class="max-w-7xl mx-auto">
      <!-- Header -->
      <div class="flex justify-between items-center mb-8">
        <div>
          <h1 class="text-3xl font-bold text-gray-900 dark:text-white">خطط الاشتراك</h1>
          <p class="text-gray-600 dark:text-gray-400 mt-1">إدارة باقات الاشتراك والميزات المتاحة</p>
        </div>
        <button 
          @click="openCreateModal"
          class="px-6 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors flex items-center gap-2 font-medium shadow-sm"
        >
          <span>+</span>
          إضافة خطة جديدة
        </button>
      </div>

      <!-- Loading -->
      <div v-if="loading" class="flex justify-center p-12">
        <div class="w-10 h-10 border-4 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
      </div>

      <!-- Error -->
      <div v-else-if="error" class="bg-red-50 dark:bg-red-900/30 text-red-600 dark:text-red-400 p-4 rounded-xl text-center">
        {{ error }}
        <button @click="fetchPlans" class="block mx-auto mt-2 text-sm underline">إعادة المحاولة</button>
      </div>

      <!-- Empty State -->
      <div v-else-if="plans.length === 0" class="text-center py-12 bg-white dark:bg-gray-800 rounded-2xl border border-gray-200 dark:border-gray-700 border-dashed">
        <p class="text-gray-500 dark:text-gray-400 mb-4">لا توجد خطط اشتراك حالياً</p>
        <button @click="openCreateModal" class="text-blue-600 hover:text-blue-700 font-medium">إنشاء أول خطة</button>
      </div>

      <!-- Plans Grid -->
      <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        <div 
          v-for="plan in plans" 
          :key="plan.id" 
          class="bg-white dark:bg-gray-800 rounded-2xl shadow-lg border border-gray-100 dark:border-gray-700 overflow-hidden relative flex flex-col"
        >
          <!-- Active Badge -->
          <div v-if="!plan.is_active" class="absolute top-4 left-4 bg-gray-200 dark:bg-gray-700 text-gray-600 dark:text-gray-400 px-2 py-1 rounded text-xs font-bold">
            موقوفة
          </div>
          <div v-if="plan.is_default" class="absolute top-4 right-4 bg-blue-100 dark:bg-blue-900/50 text-blue-600 dark:text-blue-400 px-2 py-1 rounded text-xs font-bold">
             افتراضية
          </div>

          <div class="p-6 flex-1">
            <h3 class="text-xl font-bold text-gray-900 dark:text-white mb-1">{{ plan.name_ar || plan.name }}</h3>
            <code class="text-xs bg-gray-100 dark:bg-gray-700 px-2 py-1 rounded text-gray-500 font-mono">{{ plan.code }}</code>
            
            <div class="mt-4 flex items-baseline gap-1">
              <span class="text-3xl font-bold text-gray-900 dark:text-white">{{ plan.price_monthly }}</span>
              <span class="text-sm text-gray-500">{{ plan.currency }} / شهرياً</span>
            </div>
            <div class="text-sm text-gray-500 mb-4">
              {{ plan.price_yearly }} {{ plan.currency }} / سنوياً
            </div>

            <div class="space-y-2 border-t border-gray-100 dark:border-gray-700 pt-4">
              <div class="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-300">
                <span>👤</span>
                <span>حد المستخدمين: {{ plan.max_users }}</span>
              </div>
              <div class="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-300">
                <span>💾</span>
                <span>المساحة: {{ plan.max_storage_gb }} GB</span>
              </div>
            </div>

            <div class="mt-4">
              <p class="text-xs font-bold text-gray-400 uppercase tracking-wider mb-2">الميزات</p>
              <ul class="space-y-1">
                <li v-for="feature in plan.features" :key="feature" class="text-sm text-gray-600 dark:text-gray-300 flex items-center gap-2">
                  <span class="text-green-500">✓</span>
                  {{ feature }}
                </li>
              </ul>
            </div>
          </div>

          <!-- Actions -->
          <div class="p-4 bg-gray-50 dark:bg-gray-700/50 flex gap-2 border-t border-gray-100 dark:border-gray-700">
            <button 
              @click="openEditModal(plan)"
              class="flex-1 py-2 bg-white dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-lg text-gray-700 dark:text-gray-200 hover:bg-gray-50 dark:hover:bg-gray-600 text-sm font-medium transition-colors"
            >
              تعديل
            </button>
            <button 
              @click="deletePlan(plan.id)"
              class="px-4 py-2 bg-white dark:bg-gray-700 border border-red-200 dark:border-red-900/50 text-red-600 hover:bg-red-50 dark:hover:bg-red-900/20 rounded-lg text-sm font-medium transition-colors"
            >
              حذف
            </button>
          </div>
        </div>
      </div>

      <!-- CRUD Modal -->
      <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-black/60 backdrop-blur-sm" @click="closeModal"></div>
        <div class="relative bg-white dark:bg-gray-800 rounded-2xl shadow-2xl max-w-2xl w-full max-h-[90vh] overflow-y-auto">
          <div class="p-6 border-b border-gray-100 dark:border-gray-700 flex justify-between items-center sticky top-0 bg-white dark:bg-gray-800 z-10">
            <h2 class="text-xl font-bold text-gray-900 dark:text-white">
              {{ modalMode === 'create' ? 'إضافة خطة جديدة' : 'تعديل الخطة' }}
            </h2>
            <button @click="closeModal" class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-200">
              <span class="text-2xl">×</span>
            </button>
          </div>
          
          <div class="p-6 space-y-4">
            <!-- Names -->
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">الاسم (English)</label>
                <input v-model="form.name" type="text" class="w-full p-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-white" placeholder="Basic Plan">
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">الاسم (العربية)</label>
                <input v-model="form.name_ar" type="text" class="w-full p-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-white" placeholder="الباقة الأساسية">
              </div>
            </div>

            <!-- Code & Currency -->
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">الكود (Code)</label>
                <input v-model="form.code" type="text" class="w-full p-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-white font-mono uppercase" placeholder="BASIC">
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">العملة</label>
                <select v-model="form.currency" class="w-full p-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-white">
                  <option value="LYD">LYD (دينار ليبي)</option>
                  <option value="USD">USD (دولار أمريكي)</option>
                </select>
              </div>
            </div>

            <!-- Prices -->
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">السعر الشهري</label>
                <input v-model="form.price_monthly" type="number" min="0" step="0.01" class="w-full p-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-white">
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">السعر السنوي</label>
                <input v-model="form.price_yearly" type="number" min="0" step="0.01" class="w-full p-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-white">
              </div>
            </div>

            <!-- Limits -->
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">الحد الأقصى للمستخدمين</label>
                <input v-model="form.max_users" type="number" min="1" class="w-full p-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-white">
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">سعة التخزين (GB)</label>
                <input v-model="form.max_storage_gb" type="number" min="1" class="w-full p-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-white">
              </div>
            </div>

            <!-- Features -->
            <div>
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">الميزات (نص مفصول بفاصلة)</label>
              <textarea v-model="form.features" rows="3" class="w-full p-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-white" placeholder="مثال: invoices, reports, api_access"></textarea>
              <p class="text-xs text-gray-500 mt-1">أدخل الأسماء البرمجية للميزات، مفصول بفاصلة (,)</p>
            </div>

            <!-- Checkboxes -->
            <div class="flex gap-6">
              <label class="flex items-center gap-2 cursor-pointer">
                <input v-model="form.is_active" type="checkbox" class="w-5 h-5 rounded text-blue-600 focus:ring-blue-500 border-gray-300">
                <span class="text-gray-700 dark:text-gray-300">نشط</span>
              </label>
              <label class="flex items-center gap-2 cursor-pointer">
                <input v-model="form.is_default" type="checkbox" class="w-5 h-5 rounded text-blue-600 focus:ring-blue-500 border-gray-300">
                <span class="text-gray-700 dark:text-gray-300">خطة افتراضية</span>
              </label>
            </div>

            <div v-if="modalError" class="p-3 bg-red-50 text-red-600 rounded-lg text-sm">
              {{ modalError }}
            </div>
          </div>

          <div class="p-6 border-t border-gray-100 dark:border-gray-700 flex gap-3 justify-end sticky bottom-0 bg-white dark:bg-gray-800 z-10">
            <button @click="closeModal" class="px-4 py-2 text-gray-600 hover:text-gray-800 dark:text-gray-400 dark:hover:text-gray-200">إلغاء</button>
            <button 
              @click="submitForm" 
              :disabled="saving"
              class="px-6 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg disabled:opacity-50"
            >
              {{ saving ? 'جاري الحفظ...' : 'حفظ التغييرات' }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
