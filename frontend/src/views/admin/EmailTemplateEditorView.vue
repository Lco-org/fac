<script setup>
import { ref, onMounted, computed } from 'vue'

// State
const loading = ref(true)
const saving = ref(false)
const error = ref(null)
const success = ref(null)
const templates = ref([])
const selectedTemplate = ref(null)
const previewMode = ref(false)
const previewHtml = ref('')

// Sample variables for preview
const sampleVariables = ref({
  user_name: 'أحمد محمد',
  org_name: 'شركة التقنية المتقدمة',
  code: '123456',
  username: 'admin',
  login_url: 'https://erp.company.com/login',
  amount: '5,000 LYD',
  transaction_id: 'TXN-2024-001',
  date: new Date().toLocaleDateString('ar-LY')
})

// API Base
const API_BASE = '/api/v1/organization'

// Fetch templates
const fetchTemplates = async () => {
  loading.value = true
  error.value = null
  try {
    const response = await fetch(`${API_BASE}/email-templates`, { credentials: 'include' })
    if (!response.ok) throw new Error('فشل تحميل القوالب')
    const data = await response.json()
    templates.value = data.templates || []
    if (templates.value.length > 0 && !selectedTemplate.value) {
      selectTemplate(templates.value[0])
    }
  } catch (err) {
    error.value = err.message
  } finally {
    loading.value = false
  }
}

// Select template for editing
const selectTemplate = (template) => {
  selectedTemplate.value = { ...template }
  previewMode.value = false
}

// Save template
const saveTemplate = async () => {
  if (!selectedTemplate.value) return
  
  saving.value = true
  error.value = null
  success.value = null
  
  try {
    const response = await fetch(`${API_BASE}/email-templates/${selectedTemplate.value.id}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify({
        name: selectedTemplate.value.name,
        name_ar: selectedTemplate.value.name_ar,
        subject: selectedTemplate.value.subject,
        subject_ar: selectedTemplate.value.subject_ar,
        body_html: selectedTemplate.value.body_html,
        is_active: selectedTemplate.value.is_active
      })
    })
    
    if (!response.ok) throw new Error('فشل حفظ القالب')
    success.value = 'تم حفظ القالب بنجاح'
    await fetchTemplates()
  } catch (err) {
    error.value = err.message
  } finally {
    saving.value = false
  }
}

// Preview template
const previewTemplate = async () => {
  if (!selectedTemplate.value) return
  
  try {
    const response = await fetch(`${API_BASE}/email-templates/preview`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify({
        template_key: selectedTemplate.value.template_key,
        variables: sampleVariables.value
      })
    })
    
    if (!response.ok) throw new Error('فشل المعاينة')
    const data = await response.json()
    previewHtml.value = data.body
    previewMode.value = true
  } catch (err) {
    // If preview API fails, do client-side preview
    let html = selectedTemplate.value.body_html
    for (const [key, value] of Object.entries(sampleVariables.value)) {
      html = html.replace(new RegExp(`{{${key}}}`, 'g'), value)
    }
    previewHtml.value = html
    previewMode.value = true
  }
}

// Reset template to default
const resetTemplate = async () => {
  if (!selectedTemplate.value || !confirm('هل تريد إعادة القالب للشكل الافتراضي؟')) return
  
  try {
    await fetch(`${API_BASE}/email-templates/${selectedTemplate.value.template_key}/reset`, {
      method: 'POST',
      credentials: 'include'
    })
    await fetchTemplates()
    success.value = 'تم إعادة القالب للشكل الافتراضي'
  } catch (err) {
    error.value = err.message
  }
}

// Template icon
const getTemplateIcon = (key) => {
  const icons = {
    'welcome_user': '👋',
    '2fa_code': '🔐',
    'password_reset': '🔑',
    'transaction_receipt': '🧾'
  }
  return icons[key] || '📧'
}

onMounted(() => {
  fetchTemplates()
})
</script>

<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-900 p-6" dir="rtl">
    <div class="max-w-7xl mx-auto">
      <!-- Header -->
      <div class="mb-8">
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white flex items-center gap-3">
          <span class="text-4xl">📧</span>
          إدارة قوالب البريد الإلكتروني
        </h1>
        <p class="mt-2 text-gray-500 dark:text-gray-400">
          تخصيص قوالب رسائل البريد الإلكتروني المرسلة من النظام
        </p>
      </div>

      <!-- Alerts -->
      <div v-if="error" class="mb-4 p-4 bg-red-50 dark:bg-red-900/30 border border-red-200 dark:border-red-800 rounded-xl">
        <p class="text-red-600 dark:text-red-400">❌ {{ error }}</p>
      </div>
      <div v-if="success" class="mb-4 p-4 bg-green-50 dark:bg-green-900/30 border border-green-200 dark:border-green-800 rounded-xl">
        <p class="text-green-600 dark:text-green-400">✅ {{ success }}</p>
      </div>

      <!-- Loading -->
      <div v-if="loading" class="flex justify-center p-12">
        <div class="w-12 h-12 border-4 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
      </div>

      <!-- Main Content -->
      <div v-else class="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <!-- Template List -->
        <div class="lg:col-span-1 bg-white dark:bg-gray-800 rounded-2xl shadow-lg border border-gray-100 dark:border-gray-700 overflow-hidden">
          <div class="p-4 border-b border-gray-100 dark:border-gray-700 bg-gray-50 dark:bg-gray-900">
            <h2 class="font-semibold text-gray-900 dark:text-white">القوالب المتاحة</h2>
          </div>
          <div class="divide-y divide-gray-100 dark:divide-gray-700">
            <button
              v-for="template in templates"
              :key="template.id"
              @click="selectTemplate(template)"
              :class="[
                'w-full p-4 text-right hover:bg-gray-50 dark:hover:bg-gray-700/50 transition-colors',
                selectedTemplate?.id === template.id ? 'bg-blue-50 dark:bg-blue-900/30 border-r-4 border-blue-500' : ''
              ]"
            >
              <div class="flex items-center gap-3">
                <span class="text-2xl">{{ getTemplateIcon(template.template_key) }}</span>
                <div>
                  <p class="font-medium text-gray-900 dark:text-white">{{ template.name_ar || template.name }}</p>
                  <p class="text-sm text-gray-500 dark:text-gray-400">{{ template.template_key }}</p>
                </div>
              </div>
            </button>
          </div>
        </div>

        <!-- Editor -->
        <div class="lg:col-span-2 bg-white dark:bg-gray-800 rounded-2xl shadow-lg border border-gray-100 dark:border-gray-700 overflow-hidden">
          <div v-if="selectedTemplate" class="h-full flex flex-col">
            <!-- Editor Header -->
            <div class="p-4 border-b border-gray-100 dark:border-gray-700 bg-gray-50 dark:bg-gray-900 flex justify-between items-center">
              <div class="flex items-center gap-3">
                <span class="text-2xl">{{ getTemplateIcon(selectedTemplate.template_key) }}</span>
                <div>
                  <h2 class="font-semibold text-gray-900 dark:text-white">{{ selectedTemplate.name_ar || selectedTemplate.name }}</h2>
                  <p class="text-sm text-gray-500 dark:text-gray-400">{{ selectedTemplate.description }}</p>
                </div>
              </div>
              <div class="flex gap-2">
                <button @click="previewMode = !previewMode" class="px-4 py-2 bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded-lg hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors">
                  {{ previewMode ? '📝 تعديل' : '👁️ معاينة' }}
                </button>
                <button @click="resetTemplate" class="px-4 py-2 bg-amber-100 dark:bg-amber-900/30 text-amber-700 dark:text-amber-400 rounded-lg hover:bg-amber-200 dark:hover:bg-amber-900/50 transition-colors">
                  ↩️ إعادة تعيين
                </button>
              </div>
            </div>

            <!-- Preview Mode -->
            <div v-if="previewMode" class="flex-1 p-6 overflow-auto">
              <div class="bg-gray-100 dark:bg-gray-900 rounded-xl p-6 min-h-[400px]">
                <div v-html="previewHtml || selectedTemplate.body_html" class="email-preview"></div>
              </div>
            </div>

            <!-- Edit Mode -->
            <div v-else class="flex-1 p-6 overflow-auto space-y-6">
              <!-- Subject -->
              <div>
                <label class="block text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">
                  عنوان الرسالة (Subject)
                </label>
                <input 
                  v-model="selectedTemplate.subject"
                  type="text"
                  class="w-full px-4 py-3 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  dir="ltr"
                />
              </div>

              <!-- Arabic Subject -->
              <div>
                <label class="block text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">
                  عنوان الرسالة بالعربية
                </label>
                <input 
                  v-model="selectedTemplate.subject_ar"
                  type="text"
                  class="w-full px-4 py-3 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                />
              </div>

              <!-- Body HTML -->
              <div>
                <label class="block text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">
                  محتوى الرسالة (HTML)
                </label>
                <textarea 
                  v-model="selectedTemplate.body_html"
                  rows="12"
                  class="w-full px-4 py-3 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-transparent font-mono text-sm"
                  dir="ltr"
                ></textarea>
              </div>

              <!-- Variables Info -->
              <div class="p-4 bg-blue-50 dark:bg-blue-900/30 border border-blue-200 dark:border-blue-800 rounded-xl">
                <p class="font-medium text-blue-700 dark:text-blue-300 mb-2">💡 المتغيرات المتاحة:</p>
                <div class="flex flex-wrap gap-2">
                  <code v-for="variable in selectedTemplate.variables" :key="variable" class="px-2 py-1 bg-blue-100 dark:bg-blue-800 text-blue-800 dark:text-blue-200 rounded text-sm">
                    <span v-text="'{{' + variable + '}}'"></span>
                  </code>
                </div>
              </div>
            </div>

            <!-- Save Button -->
            <div class="p-4 border-t border-gray-100 dark:border-gray-700 bg-gray-50 dark:bg-gray-900">
              <button 
                @click="saveTemplate"
                :disabled="saving"
                class="w-full py-3 bg-gradient-to-r from-blue-600 to-indigo-600 text-white rounded-xl font-bold shadow-lg hover:shadow-xl transition-all disabled:opacity-50"
              >
                <span v-if="saving" class="flex items-center justify-center gap-2">
                  <span class="w-5 h-5 border-2 border-white border-t-transparent rounded-full animate-spin"></span>
                  جاري الحفظ...
                </span>
                <span v-else>💾 حفظ التغييرات</span>
              </button>
            </div>
          </div>

          <!-- No Selection -->
          <div v-else class="flex items-center justify-center h-96 text-gray-500 dark:text-gray-400">
            <p>اختر قالب من القائمة للتعديل</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.email-preview :deep(*) {
  max-width: 100%;
}
</style>
