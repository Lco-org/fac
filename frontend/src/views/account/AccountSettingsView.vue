<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { usePreferencesStore } from '@/stores/preferences'

const router = useRouter()
const authStore = useAuthStore()
const preferencesStore = usePreferencesStore()

// Active tab
const activeTab = ref('profile')

// State
const loading = ref(false)
const error = ref('')
const success = ref('')

// ========== PROFILE TAB ==========
const avatarInput = ref(null)

const handleAvatarUpload = (event) => {
  const file = event.target.files[0]
  if (!file) return

  if (file.size > 2 * 1024 * 1024) {
    error.value = 'حجم الصورة يجب أن يكون أقل من 2 ميجابايت'
    return
  }

  const reader = new FileReader()
  reader.onload = (e) => {
    preferencesStore.setAvatar(e.target.result)
    success.value = 'تم تحديث الصورة الشخصية'
    setTimeout(() => success.value = '', 3000)
  }
  reader.readAsDataURL(file)
}

const removeAvatar = () => {
  preferencesStore.removeAvatar()
  success.value = 'تم حذف الصورة الشخصية'
  setTimeout(() => success.value = '', 3000)
}

// ========== APPEARANCE TAB ==========
const themes = [
  { id: 'light', name: 'فاتح', icon: '☀️', preview: 'bg-white text-gray-900' },
  { id: 'dark', name: 'داكن', icon: '🌙', preview: 'bg-gray-900 text-white' },
  { id: 'sepia', name: 'بني', icon: '📜', preview: 'bg-amber-50 text-amber-900' },
  { id: 'system', name: 'تلقائي', icon: '💻', preview: 'bg-gradient-to-r from-white to-gray-900 text-gray-600' },
]

const primaryColors = [
  { name: 'أزرق', value: '#1e40af' },
  { name: 'أخضر', value: '#166534' },
  { name: 'أحمر', value: '#b91c1c' },
  { name: 'بنفسجي', value: '#7c3aed' },
  { name: 'برتقالي', value: '#c2410c' },
  { name: 'وردي', value: '#be185d' },
  { name: 'سماوي', value: '#0891b2' },
  { name: 'رمادي', value: '#4b5563' },
]

const fontSizes = [
  { id: 'small', name: 'صغير', sample: 'text-sm' },
  { id: 'medium', name: 'متوسط', sample: 'text-base' },
  { id: 'large', name: 'كبير', sample: 'text-lg' },
]

const borderRadii = [
  { id: 'sharp', name: 'حاد', sample: 'rounded-sm' },
  { id: 'rounded', name: 'دائري', sample: 'rounded-lg' },
  { id: 'pill', name: 'بيضوي', sample: 'rounded-full' },
]

const customColor = ref(preferencesStore.primaryColor)

const applyCustomColor = () => {
  preferencesStore.setPrimaryColor(customColor.value)
  success.value = 'تم تطبيق اللون المخصص'
  setTimeout(() => success.value = '', 3000)
}

const resetAppearance = () => {
  preferencesStore.resetToDefaults()
  customColor.value = '#1e40af'
  success.value = 'تم إعادة الإعدادات الافتراضية'
  setTimeout(() => success.value = '', 3000)
}

// ========== SECURITY TAB ==========
const showPasswordForm = ref(false)
const oldPassword = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const twoFactorEnabled = ref(false)

onMounted(() => {
  if (authStore.user) {
    twoFactorEnabled.value = authStore.user.two_factor_enabled
  }
})

const handleToggle2FA = async () => {
  error.value = ''
  success.value = ''
  loading.value = true

  const result = await authStore.toggle2FA(!twoFactorEnabled.value)
  
  loading.value = false

  if (result.success) {
    twoFactorEnabled.value = !twoFactorEnabled.value
    success.value = result.message
    setTimeout(() => success.value = '', 3000)
  } else {
    error.value = result.error
  }
}

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

// Format date
const formatDate = (dateStr) => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleDateString('ar-SA', {
    year: 'numeric',
    month: 'long',
    day: 'numeric'
  })
}

const userInitials = computed(() => {
  const name = authStore.user?.full_name || ''
  return name.split(' ').map(n => n[0]).join('').substring(0, 2).toUpperCase() || '؟'
})

const getRoleName = (role) => {
  const roles = {
    admin: 'مدير النظام',
    finance: 'محاسب',
    hr: 'موارد بشرية',
    auditor: 'مراجع'
  }
  return roles[role] || role
}
</script>

<template>
  <div class="max-w-4xl mx-auto">
    <h1 class="text-2xl font-bold mb-6">⚙️ إعدادات الحساب</h1>

    <!-- Messages -->
    <div v-if="error" class="mb-4 bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg">
      ⚠️ {{ error }}
    </div>
    <div v-if="success" class="mb-4 bg-green-50 border border-green-200 text-green-700 px-4 py-3 rounded-lg">
      ✅ {{ success }}
    </div>

    <!-- Tabs -->
    <div class="card mb-6">
      <div class="flex border-b overflow-x-auto">
        <button
          v-for="tab in [
            { id: 'profile', name: 'الملف الشخصي', icon: '👤' },
            { id: 'appearance', name: 'المظهر', icon: '🎨' },
            { id: 'security', name: 'الأمان', icon: '🔐' },
          ]"
          :key="tab.id"
          @click="activeTab = tab.id"
          :class="[
            'flex items-center gap-2 px-6 py-3 font-medium transition-colors whitespace-nowrap',
            activeTab === tab.id 
              ? 'border-b-2 border-primary-600 text-primary-600 bg-primary-50' 
              : 'text-gray-500 hover:text-gray-700 hover:bg-gray-50'
          ]"
        >
          <span>{{ tab.icon }}</span>
          <span>{{ tab.name }}</span>
        </button>
      </div>

      <!-- Profile Tab -->
      <div v-if="activeTab === 'profile'" class="p-6">
        <div class="flex flex-col md:flex-row gap-8">
          <!-- Avatar Section -->
          <div class="flex flex-col items-center">
            <div class="relative mb-4">
              <div 
                v-if="preferencesStore.avatar"
                class="w-32 h-32 rounded-full overflow-hidden border-4 border-primary-200"
              >
                <img :src="preferencesStore.avatar" alt="Avatar" class="w-full h-full object-cover" />
              </div>
              <div 
                v-else 
                class="w-32 h-32 rounded-full bg-primary-100 text-primary-600 flex items-center justify-center text-4xl font-bold border-4 border-primary-200"
              >
                {{ userInitials }}
              </div>
              <button
                v-if="preferencesStore.avatar"
                @click="removeAvatar"
                class="absolute -top-1 -right-1 w-8 h-8 bg-red-500 text-white rounded-full flex items-center justify-center hover:bg-red-600 transition-colors"
                title="حذف الصورة"
              >
                ✕
              </button>
            </div>
            <input
              ref="avatarInput"
              type="file"
              accept="image/*"
              class="hidden"
              @change="handleAvatarUpload"
            />
            <button
              @click="avatarInput.click()"
              class="btn-secondary text-sm"
            >
              📷 {{ preferencesStore.avatar ? 'تغيير الصورة' : 'إضافة صورة' }}
            </button>
            <p class="text-xs text-gray-500 mt-2">الحد الأقصى: 2 ميجابايت</p>
          </div>

          <!-- Profile Info -->
          <div class="flex-1 space-y-4">
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div class="bg-gray-50 rounded-lg p-4">
                <label class="text-sm text-gray-500">الاسم الكامل</label>
                <p class="font-semibold text-lg">{{ authStore.user?.full_name || '-' }}</p>
              </div>
              <div class="bg-gray-50 rounded-lg p-4">
                <label class="text-sm text-gray-500">اسم المستخدم</label>
                <p class="font-semibold text-lg">{{ authStore.user?.username || '-' }}</p>
              </div>
              <div class="bg-gray-50 rounded-lg p-4">
                <label class="text-sm text-gray-500">البريد الإلكتروني</label>
                <p class="font-semibold text-lg">{{ authStore.user?.email || '-' }}</p>
              </div>
              <div class="bg-gray-50 rounded-lg p-4">
                <label class="text-sm text-gray-500">الصلاحية</label>
                <p class="font-semibold text-lg">
                  <span class="inline-flex items-center gap-1 px-2 py-1 bg-primary-100 text-primary-700 rounded-full text-sm">
                    {{ getRoleName(authStore.userRole) }}
                  </span>
                </p>
              </div>
              <div class="bg-gray-50 rounded-lg p-4 md:col-span-2">
                <label class="text-sm text-gray-500">تاريخ إنشاء الحساب</label>
                <p class="font-semibold text-lg">{{ formatDate(authStore.user?.created_at) }}</p>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Appearance Tab -->
      <div v-if="activeTab === 'appearance'" class="p-6 space-y-8">
        <!-- Theme Selection -->
        <div>
          <h3 class="text-lg font-semibold mb-4">🎨 نمط العرض</h3>
          <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
            <button
              v-for="t in themes"
              :key="t.id"
              @click="preferencesStore.setTheme(t.id)"
              :class="[
                'p-4 rounded-xl border-2 transition-all text-center',
                preferencesStore.theme === t.id 
                  ? 'border-primary-500 ring-2 ring-primary-200' 
                  : 'border-gray-200 hover:border-gray-300'
              ]"
            >
              <div :class="['w-full h-12 rounded-lg mb-2 flex items-center justify-center text-2xl', t.preview]">
                {{ t.icon }}
              </div>
              <span class="font-medium">{{ t.name }}</span>
            </button>
          </div>
        </div>

        <!-- Primary Color -->
        <div>
          <h3 class="text-lg font-semibold mb-4">🖌️ اللون الرئيسي</h3>
          <div class="grid grid-cols-4 md:grid-cols-8 gap-3 mb-4">
            <button
              v-for="color in primaryColors"
              :key="color.value"
              @click="preferencesStore.setPrimaryColor(color.value)"
              :style="{ backgroundColor: color.value }"
              :class="[
                'w-12 h-12 rounded-full transition-transform hover:scale-110',
                preferencesStore.primaryColor === color.value 
                  ? 'ring-4 ring-offset-2 ring-gray-400' 
                  : ''
              ]"
              :title="color.name"
            />
          </div>
          <div class="flex items-center gap-4">
            <label class="text-sm text-gray-600">لون مخصص:</label>
            <input
              v-model="customColor"
              type="color"
              class="w-12 h-10 rounded cursor-pointer border-0"
            />
            <button @click="applyCustomColor" class="btn-secondary text-sm">
              تطبيق
            </button>
          </div>
        </div>

        <!-- Font Size -->
        <div>
          <h3 class="text-lg font-semibold mb-4">📝 حجم الخط</h3>
          <div class="flex gap-4">
            <button
              v-for="size in fontSizes"
              :key="size.id"
              @click="preferencesStore.setFontSize(size.id)"
              :class="[
                'px-6 py-3 rounded-lg border-2 transition-all',
                size.sample,
                preferencesStore.fontSize === size.id 
                  ? 'border-primary-500 bg-primary-50' 
                  : 'border-gray-200 hover:border-gray-300'
              ]"
            >
              {{ size.name }}
            </button>
          </div>
        </div>

        <!-- Border Radius -->
        <div>
          <h3 class="text-lg font-semibold mb-4">⬜ شكل الأزرار</h3>
          <div class="flex gap-4">
            <button
              v-for="radius in borderRadii"
              :key="radius.id"
              @click="preferencesStore.setBorderRadius(radius.id)"
              :class="[
                'px-6 py-3 border-2 transition-all bg-primary-100 text-primary-700',
                radius.sample,
                preferencesStore.borderRadius === radius.id 
                  ? 'border-primary-500 ring-2 ring-primary-200' 
                  : 'border-gray-200 hover:border-gray-300'
              ]"
            >
              {{ radius.name }}
            </button>
          </div>
        </div>

        <!-- Alert Colors Section -->
        <div class="border-t pt-6">
          <h3 class="text-lg font-semibold mb-4">🚦 ألوان التنبيهات</h3>
          
          <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
            <!-- Surplus Color (Positive) -->
            <div class="bg-gray-50 rounded-xl p-4">
              <label class="block text-sm font-medium mb-3">✅ لون الدائن (الأرصدة الموجبة)</label>
              <div class="flex flex-wrap gap-2 mb-3">
                <button
                  v-for="color in ['#22c55e', '#16a34a', '#15803d', '#10b981', '#059669', '#0d9488']"
                  :key="color"
                  @click="preferencesStore.setSurplusColor(color)"
                  :style="{ backgroundColor: color }"
                  :class="[
                    'w-10 h-10 rounded-full transition-transform hover:scale-110',
                    preferencesStore.surplusColor === color 
                      ? 'ring-4 ring-offset-2 ring-gray-400' 
                      : ''
                  ]"
                />
              </div>
              <div class="flex items-center gap-3">
                <input
                  :value="preferencesStore.surplusColor"
                  @input="preferencesStore.setSurplusColor($event.target.value)"
                  type="color"
                  class="w-10 h-10 rounded cursor-pointer border-0"
                />
                <span class="text-sm text-gray-500">لون مخصص</span>
                <div 
                  class="flex-1 h-8 rounded-lg flex items-center justify-center text-white font-bold"
                  :style="{ backgroundColor: preferencesStore.surplusColor }"
                >
                  +1,000
                </div>
              </div>
            </div>

            <!-- Deficit Color (Negative) -->
            <div class="bg-gray-50 rounded-xl p-4">
              <label class="block text-sm font-medium mb-3">⚠️ لون العجز (الأرصدة السالبة)</label>
              <div class="flex flex-wrap gap-2 mb-3">
                <button
                  v-for="color in ['#ef4444', '#dc2626', '#b91c1c', '#f97316', '#ea580c', '#c2410c']"
                  :key="color"
                  @click="preferencesStore.setDeficitColor(color)"
                  :style="{ backgroundColor: color }"
                  :class="[
                    'w-10 h-10 rounded-full transition-transform hover:scale-110',
                    preferencesStore.deficitColor === color 
                      ? 'ring-4 ring-offset-2 ring-gray-400' 
                      : ''
                  ]"
                />
              </div>
              <div class="flex items-center gap-3">
                <input
                  :value="preferencesStore.deficitColor"
                  @input="preferencesStore.setDeficitColor($event.target.value)"
                  type="color"
                  class="w-10 h-10 rounded cursor-pointer border-0"
                />
                <span class="text-sm text-gray-500">لون مخصص</span>
                <div 
                  class="flex-1 h-8 rounded-lg flex items-center justify-center text-white font-bold"
                  :style="{ backgroundColor: preferencesStore.deficitColor }"
                >
                  -1,000
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Reset Button -->
        <div class="pt-4 border-t">
          <button @click="resetAppearance" class="text-gray-500 hover:text-gray-700 text-sm">
            🔄 إعادة الإعدادات الافتراضية
          </button>
        </div>
      </div>

      <!-- Security Tab -->
      <div v-if="activeTab === 'security'" class="p-6 space-y-6">
        <!-- Two-Factor Authentication -->
        <div class="bg-gray-50 rounded-xl p-6">
          <h3 class="text-lg font-semibold mb-4 flex items-center gap-2">
            <span class="text-2xl">🔐</span>
            التحقق الثنائي (2FA)
          </h3>
          <p class="text-gray-600 mb-4">
            عند تفعيل التحقق الثنائي، سيتم إرسال رمز تحقق إلى بريدك الإلكتروني في كل مرة تقوم فيها بتسجيل الدخول.
          </p>
          
          <div class="flex items-center justify-between p-4 bg-white rounded-lg border">
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
        <div class="bg-gray-50 rounded-xl p-6">
          <h3 class="text-lg font-semibold mb-4 flex items-center gap-2">
            <span class="text-2xl">🔑</span>
            تغيير كلمة المرور
          </h3>
          
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
        <div class="bg-red-50 rounded-xl p-6 border-2 border-red-200">
          <h3 class="text-lg font-semibold mb-4 flex items-center gap-2 text-red-700">
            <span class="text-2xl">⚠️</span>
            إلغاء جميع الجلسات
          </h3>
          <p class="text-gray-600 mb-4">
            سيتم تسجيل خروجك من جميع الأجهزة المتصلة بحسابك. استخدم هذا الخيار إذا كنت تشك في اختراق حسابك.
          </p>
          
          <button
            @click="handleRevokeAllSessions"
            :disabled="loading"
            class="bg-red-500 hover:bg-red-600 text-white font-medium py-2 px-4 rounded-lg transition-colors disabled:opacity-50"
          >
            {{ loading ? 'جاري الإلغاء...' : 'إلغاء جميع الجلسات' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
