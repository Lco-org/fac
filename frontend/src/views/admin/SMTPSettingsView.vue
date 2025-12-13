<script setup>
import { ref, onMounted } from 'vue'

// State
const loading = ref(true)
const saving = ref(false)
const testing = ref(false)
const error = ref(null)
const success = ref(null)
const smtpSettings = ref([])
const showForm = ref(false)
const editingId = ref(null)

const form = ref({
  name: '',
  smtp_host: '',
  smtp_port: 587,
  username: '',
  password: '',
  from_address: '',
  from_name: '',
  use_tls: true,
  is_default: false
})

const testEmail = ref('')
const showTestModal = ref(false)
const testingId = ref(null)

// API Base
const API_BASE = '/api/v1/organization'

// Fetch SMTP settings
const fetchSettings = async () => {
  loading.value = true
  error.value = null
  try {
    const response = await fetch(`${API_BASE}/smtp-settings`, { credentials: 'include' })
    if (!response.ok) throw new Error('فشل تحميل الإعدادات')
    const data = await response.json()
    smtpSettings.value = data.smtp_settings || []
  } catch (err) {
    error.value = err.message
  } finally {
    loading.value = false
  }
}

// Reset form
const resetForm = () => {
  form.value = {
    name: '',
    smtp_host: '',
    smtp_port: 587,
    username: '',
    password: '',
    from_address: '',
    from_name: '',
    use_tls: true,
    is_default: false
  }
  editingId.value = null
}

// Open form for new
const openNewForm = () => {
  resetForm()
  showForm.value = true
}

// Open form for edit
const editSettings = (settings) => {
  form.value = {
    name: settings.name,
    smtp_host: settings.smtp_host,
    smtp_port: settings.smtp_port,
    username: settings.username,
    password: '', // Don't show existing password
    from_address: settings.from_address,
    from_name: settings.from_name || '',
    use_tls: settings.use_tls,
    is_default: settings.is_default
  }
  editingId.value = settings.id
  showForm.value = true
}

// Save settings
const saveSettings = async () => {
  saving.value = true
  error.value = null
  success.value = null
  
  try {
    const url = editingId.value 
      ? `${API_BASE}/smtp-settings/${editingId.value}`
      : `${API_BASE}/smtp-settings`
    
    const response = await fetch(url, {
      method: editingId.value ? 'PUT' : 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify(form.value)
    })
    
    if (!response.ok) throw new Error('فشل حفظ الإعدادات')
    
    success.value = editingId.value ? 'تم تحديث الإعدادات' : 'تم إنشاء إعدادات SMTP جديدة'
    showForm.value = false
    resetForm()
    await fetchSettings()
  } catch (err) {
    error.value = err.message
  } finally {
    saving.value = false
  }
}

// Delete settings
const deleteSettings = async (id) => {
  if (!confirm('هل تريد حذف إعدادات SMTP هذه؟')) return
  
  try {
    await fetch(`${API_BASE}/smtp-settings/${id}`, {
      method: 'DELETE',
      credentials: 'include'
    })
    await fetchSettings()
    success.value = 'تم حذف الإعدادات'
  } catch (err) {
    error.value = err.message
  }
}

// Open test modal
const openTestModal = (id) => {
  testingId.value = id
  testEmail.value = ''
  showTestModal.value = true
}

// Send test email
const sendTestEmail = async () => {
  if (!testEmail.value) return
  
  testing.value = true
  error.value = null
  
  try {
    const response = await fetch(`${API_BASE}/smtp-settings/${testingId.value}/test`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify({ to_email: testEmail.value })
    })
    
    if (!response.ok) throw new Error('فشل إرسال الاختبار')
    
    success.value = `تم إرسال رسالة اختبار إلى ${testEmail.value}`
    showTestModal.value = false
  } catch (err) {
    error.value = err.message
  } finally {
    testing.value = false
  }
}

onMounted(() => {
  fetchSettings()
})
</script>

<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-900 p-6" dir="rtl">
    <div class="max-w-4xl mx-auto">
      <!-- Header -->
      <div class="flex justify-between items-center mb-8">
        <div>
          <h1 class="text-3xl font-bold text-gray-900 dark:text-white flex items-center gap-3">
            <span class="text-4xl">📮</span>
            إعدادات البريد الإلكتروني (SMTP)
          </h1>
          <p class="mt-2 text-gray-500 dark:text-gray-400">
            إعداد خادم البريد الإلكتروني للمنظمة
          </p>
        </div>
        <button 
          @click="openNewForm"
          class="px-6 py-3 bg-gradient-to-r from-blue-600 to-indigo-600 text-white rounded-xl font-bold shadow-lg hover:shadow-xl transition-all flex items-center gap-2"
        >
          <span class="text-xl">➕</span>
          إضافة SMTP
        </button>
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

      <!-- SMTP List -->
      <div v-else class="space-y-4">
        <div v-if="smtpSettings.length === 0" class="bg-white dark:bg-gray-800 rounded-2xl shadow-lg p-12 text-center">
          <div class="w-24 h-24 mx-auto mb-4 bg-gray-100 dark:bg-gray-700 rounded-full flex items-center justify-center">
            <span class="text-4xl">📧</span>
          </div>
          <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-2">لا توجد إعدادات SMTP</h3>
          <p class="text-gray-500 dark:text-gray-400 mb-6">أضف إعدادات SMTP للبدء بإرسال البريد</p>
          <button @click="openNewForm" class="px-6 py-3 bg-blue-600 text-white rounded-lg">
            إضافة إعدادات SMTP
          </button>
        </div>

        <div 
          v-for="smtp in smtpSettings" 
          :key="smtp.id"
          class="bg-white dark:bg-gray-800 rounded-2xl shadow-lg border border-gray-100 dark:border-gray-700 overflow-hidden"
        >
          <div class="p-6">
            <div class="flex justify-between items-start">
              <div class="flex items-center gap-4">
                <div class="w-14 h-14 bg-gradient-to-br from-purple-500 to-pink-600 rounded-xl flex items-center justify-center text-white text-2xl shadow-lg">
                  📮
                </div>
                <div>
                  <div class="flex items-center gap-2">
                    <h3 class="text-lg font-bold text-gray-900 dark:text-white">{{ smtp.name }}</h3>
                    <span v-if="smtp.is_default" class="px-2 py-0.5 bg-green-100 dark:bg-green-900 text-green-700 dark:text-green-300 text-xs font-medium rounded-full">
                      افتراضي
                    </span>
                  </div>
                  <p class="text-gray-500 dark:text-gray-400 text-sm font-mono">
                    {{ smtp.smtp_host }}:{{ smtp.smtp_port }}
                  </p>
                </div>
              </div>
              <div class="flex gap-2">
                <button 
                  @click="openTestModal(smtp.id)"
                  class="p-2 bg-blue-50 dark:bg-blue-900/30 text-blue-600 dark:text-blue-400 rounded-lg hover:bg-blue-100 dark:hover:bg-blue-900/50"
                  title="اختبار"
                >
                  🧪
                </button>
                <button 
                  @click="editSettings(smtp)"
                  class="p-2 bg-amber-50 dark:bg-amber-900/30 text-amber-600 dark:text-amber-400 rounded-lg hover:bg-amber-100 dark:hover:bg-amber-900/50"
                  title="تعديل"
                >
                  ✏️
                </button>
                <button 
                  @click="deleteSettings(smtp.id)"
                  class="p-2 bg-red-50 dark:bg-red-900/30 text-red-600 dark:text-red-400 rounded-lg hover:bg-red-100 dark:hover:bg-red-900/50"
                  title="حذف"
                >
                  🗑️
                </button>
              </div>
            </div>

            <div class="mt-4 grid grid-cols-2 gap-4 text-sm">
              <div class="flex justify-between p-3 bg-gray-50 dark:bg-gray-700 rounded-lg">
                <span class="text-gray-500 dark:text-gray-400">البريد المرسل:</span>
                <span class="text-gray-900 dark:text-white font-mono">{{ smtp.from_address }}</span>
              </div>
              <div class="flex justify-between p-3 bg-gray-50 dark:bg-gray-700 rounded-lg">
                <span class="text-gray-500 dark:text-gray-400">TLS:</span>
                <span :class="smtp.use_tls ? 'text-green-600' : 'text-red-600'">
                  {{ smtp.use_tls ? '✓ مفعل' : '✗ معطل' }}
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Form Modal -->
      <div v-if="showForm" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
        <div class="bg-white dark:bg-gray-800 rounded-2xl shadow-2xl w-full max-w-xl max-h-[90vh] overflow-auto">
          <div class="p-6 border-b border-gray-100 dark:border-gray-700">
            <h2 class="text-xl font-bold text-gray-900 dark:text-white">
              {{ editingId ? 'تعديل إعدادات SMTP' : 'إضافة إعدادات SMTP جديدة' }}
            </h2>
          </div>
          
          <form @submit.prevent="saveSettings" class="p-6 space-y-4">
            <div class="grid grid-cols-2 gap-4">
              <div class="col-span-2">
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">الاسم</label>
                <input v-model="form.name" required class="w-full px-4 py-3 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-xl" placeholder="SMTP الرئيسي"/>
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">الخادم (Host)</label>
                <input v-model="form.smtp_host" required class="w-full px-4 py-3 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-xl" dir="ltr" placeholder="smtp.gmail.com"/>
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">المنفذ (Port)</label>
                <input v-model.number="form.smtp_port" type="number" required class="w-full px-4 py-3 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-xl" dir="ltr"/>
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">اسم المستخدم</label>
                <input v-model="form.username" class="w-full px-4 py-3 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-xl" dir="ltr"/>
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">كلمة المرور</label>
                <input v-model="form.password" type="password" class="w-full px-4 py-3 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-xl" dir="ltr" :placeholder="editingId ? '(لا تغيير)' : ''"/>
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">البريد المرسل</label>
                <input v-model="form.from_address" type="email" required class="w-full px-4 py-3 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-xl" dir="ltr"/>
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">اسم المرسل</label>
                <input v-model="form.from_name" class="w-full px-4 py-3 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-xl"/>
              </div>
            </div>
            <div class="flex gap-4">
              <label class="flex items-center gap-2 cursor-pointer">
                <input v-model="form.use_tls" type="checkbox" class="w-5 h-5 rounded border-gray-300"/>
                <span class="text-gray-700 dark:text-gray-300">استخدام TLS</span>
              </label>
              <label class="flex items-center gap-2 cursor-pointer">
                <input v-model="form.is_default" type="checkbox" class="w-5 h-5 rounded border-gray-300"/>
                <span class="text-gray-700 dark:text-gray-300">افتراضي</span>
              </label>
            </div>
            <div class="flex gap-3 pt-4">
              <button type="submit" :disabled="saving" class="flex-1 py-3 bg-blue-600 text-white rounded-xl font-bold disabled:opacity-50">
                {{ saving ? 'جاري الحفظ...' : '💾 حفظ' }}
              </button>
              <button type="button" @click="showForm = false" class="px-6 py-3 bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded-xl font-bold">
                إلغاء
              </button>
            </div>
          </form>
        </div>
      </div>

      <!-- Test Modal -->
      <div v-if="showTestModal" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
        <div class="bg-white dark:bg-gray-800 rounded-2xl shadow-2xl w-full max-w-md">
          <div class="p-6 border-b border-gray-100 dark:border-gray-700">
            <h2 class="text-xl font-bold text-gray-900 dark:text-white">🧪 اختبار SMTP</h2>
          </div>
          <div class="p-6">
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              البريد الإلكتروني للاختبار
            </label>
            <input 
              v-model="testEmail" 
              type="email" 
              class="w-full px-4 py-3 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-xl mb-4" 
              dir="ltr"
              placeholder="test@example.com"
            />
            <div class="flex gap-3">
              <button 
                @click="sendTestEmail" 
                :disabled="testing || !testEmail"
                class="flex-1 py-3 bg-blue-600 text-white rounded-xl font-bold disabled:opacity-50"
              >
                {{ testing ? 'جاري الإرسال...' : '📤 إرسال' }}
              </button>
              <button 
                @click="showTestModal = false" 
                class="px-6 py-3 bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded-xl font-bold"
              >
                إلغاء
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
