<script setup>
import { ref, onMounted } from 'vue'
import api from '@/api/client'

const loading = ref(false)
const success = ref(false)
const error = ref('')

// Asset Donation Toggle
const isAssetDonation = ref(false)
const assetCategories = ref([])
const assetTypes = ref([])
const selectedAssets = ref([]) // Array of { asset_type_id, quantity, name, description }

const form = ref({
  type: 'income_donation',
  amount: '',
  currency: 'LYD',
  exchange_rate: 1.0,
  container_id: '',
  project_id: '',
  member_id: '',
  description: '',
  description_ar: '',
  payee_payer: '',
  reference_number: '',
  confirm_deficit: false,
})

const containers = ref([])
const projects = ref([])
const members = ref([])
const employees = ref([])

onMounted(async () => {
  try {
    const dailyPos = await api.get('/reports/daily-position')
    if (dailyPos.data) {
      const allContainers = [
        ...(dailyPos.data.cash_containers || []),
        ...(dailyPos.data.bank_containers || [])
      ]
      containers.value = allContainers.map(c => ({
        id: c.id,
        name: c.name_ar || c.name,
        type: c.type,
        currency: c.currency
      }))
    }

    const projectSum = await api.get('/reports/project-summary')
    if (projectSum.data && Array.isArray(projectSum.data)) {
      projects.value = projectSum.data.map(p => ({
        id: p.id,
        code: p.code,
        name: p.name_ar || p.name,
        currency: p.currency || 'LYD'
      }))
    }

    // Load members for subscription/donation linking - sorted alphabetically
    const membersRes = await api.get('/members')
    if (membersRes.data && Array.isArray(membersRes.data)) {
      members.value = membersRes.data
        .filter(m => m.is_active)
        .map(m => ({
          id: m.id,
          name: m.full_name,
          number: m.membership_number,
          type: m.member_type,
          category: 'member'
        }))
        .sort((a, b) => a.name.localeCompare(b.name, 'ar'))
    }

    // Load employees - sorted alphabetically
    const employeesRes = await api.get('/employees')
    if (employeesRes.data && Array.isArray(employeesRes.data)) {
      employees.value = employeesRes.data
        .filter(e => e.is_active)
        .map(e => ({
          id: e.id,
          name: e.full_name,
          number: e.employee_number,
          position: e.position,
          category: 'employee'
        }))
        .sort((a, b) => a.name.localeCompare(b.name, 'ar'))
    }

    // Load asset categories and types for in-kind donations
    try {
      const catRes = await api.get('/asset-categories')
      if (catRes.data?.categories) {
        assetCategories.value = catRes.data.categories
      }
      const typesRes = await api.get('/asset-types')
      if (typesRes.data?.types) {
        assetTypes.value = typesRes.data.types
      }
    } catch (assetErr) {
      console.warn('Asset categories not available:', assetErr)
    }
  } catch (err) {
    console.error('Failed to load data:', err)
  }
})

// Auto-fill payer name when member/employee selected
const onMemberChange = () => {
  // Search in both employees and members
  const selectedMember = members.value.find(m => m.id === form.value.member_id)
  const selectedEmployee = employees.value.find(e => e.id === form.value.member_id)
  
  if (selectedMember) {
    form.value.payee_payer = selectedMember.name
  } else if (selectedEmployee) {
    form.value.payee_payer = selectedEmployee.name
  }
}

const handleSubmit = async () => {
  error.value = ''
  success.value = false
  loading.value = true

  try {
    const payload = {
      ...form.value,
      description: form.value.description_ar || form.value.description,
      member_id: form.value.member_id || null
    }

    // For asset donations, set amount to 0 and add assets
    if (isAssetDonation.value) {
      payload.amount = 0
      payload.assets = selectedAssets.value.filter(a => a.asset_type_id && a.quantity > 0)
    }

    const response = await api.post('/transactions', payload)
    success.value = true
    
    // Create assets if this is an asset donation
    if (isAssetDonation.value && selectedAssets.value.length > 0) {
      for (const asset of selectedAssets.value) {
        if (asset.asset_type_id && asset.quantity > 0) {
          await api.post('/assets', {
            asset_type_id: asset.asset_type_id,
            project_id: form.value.project_id,
            transaction_id: response.data.transaction?.id,
            name: asset.name || '',
            name_ar: asset.name || '',
            quantity: asset.quantity,
            unit_price: asset.unit_price || 0,
            total_value: (asset.quantity * (asset.unit_price || 0)),
            currency: form.value.currency,
            condition: 'new',
            status: 'available',
            acquisition_type: 'donation',
            acquisition_date: new Date().toISOString().split('T')[0],
            supplier: form.value.payee_payer || ''
          })
        }
      }
    }
    
    // Open receipt in new tab
    if (response.data.transaction && response.data.transaction.id) {
      const receiptUrl = `/receipt/${response.data.transaction.id}`
      window.open(receiptUrl, '_blank')
    }
    
    resetForm()
  } catch (err) {
    error.value = err.response?.data?.error || 'فشل في إنشاء العملية'
  } finally {
    loading.value = false
  }
}

const resetForm = () => {
  form.value = {
    type: 'income_donation',
    amount: '',
    currency: 'LYD',
    exchange_rate: 1.0,
    container_id: '',
    project_id: '',
    member_id: '',
    description: '',
    description_ar: '',
    payee_payer: '',
    reference_number: '',
    confirm_deficit: false,
  }
  isAssetDonation.value = false
  selectedAssets.value = []
}

// Asset helpers
const addAsset = () => {
  selectedAssets.value.push({
    asset_type_id: '',
    name: '',
    quantity: 1,
    unit_price: 0
  })
}

const removeAsset = (index) => {
  selectedAssets.value.splice(index, 1)
}

const getTypesForCategory = (categoryId) => {
  return assetTypes.value.filter(t => t.category_id === categoryId)
}
</script>

<template>
  <div>
    <h1 class="text-3xl font-bold mb-6">تبرعات واشتراكات</h1>

    <div class="card max-w-2xl">
      <div v-if="success" class="bg-surplus-light border border-surplus text-surplus-dark px-4 py-3 rounded-lg mb-4">
        ✅ تمت العملية بنجاح
      </div>

      <div v-if="error" class="bg-deficit-light border border-deficit text-deficit-dark px-4 py-3 rounded-lg mb-4">
        ❌ {{ error }}
      </div>

      <form @submit.prevent="handleSubmit" class="space-y-4">
        <!-- Asset Donation Toggle -->
        <div class="flex items-center gap-3 p-3 bg-blue-50 dark:bg-blue-900/20 rounded-lg border border-blue-200 dark:border-blue-800">
          <input 
            type="checkbox" 
            id="isAssetDonation" 
            v-model="isAssetDonation" 
            class="w-5 h-5 text-blue-600 rounded focus:ring-blue-500"
          />
          <label for="isAssetDonation" class="text-sm font-medium text-blue-800 dark:text-blue-300 cursor-pointer">
            📦 منحة عينية (مواد / أصول) - بدون مبلغ مالي
          </label>
        </div>

        <!-- Amount Field - Hidden for Asset Donations -->
        <div v-if="!isAssetDonation">
          <label class="block text-sm font-medium text-gray-700 mb-2">
            المبلغ <span class="text-deficit">*</span>
          </label>
          <input v-model.number="form.amount" type="number" step="0.01" :required="!isAssetDonation" class="input-field" />
        </div>

        <!-- Asset List Section -->
        <div v-if="isAssetDonation" class="space-y-4 p-4 bg-gradient-to-br from-emerald-50 to-green-50 dark:from-emerald-900/20 dark:to-green-900/20 rounded-xl border border-emerald-200 dark:border-emerald-800">
          <div class="flex items-center justify-between">
            <h3 class="text-lg font-bold text-emerald-800 dark:text-emerald-300 flex items-center gap-2">
              <span>📦</span> المواد / الأصول المتبرع بها
            </h3>
            <button 
              type="button" 
              @click="addAsset" 
              class="px-3 py-1 bg-emerald-600 text-white rounded-lg text-sm hover:bg-emerald-700 transition-colors"
            >
              + إضافة مادة
            </button>
          </div>

          <div v-if="selectedAssets.length === 0" class="text-center py-6 text-gray-500">
            <p>اضغط "إضافة مادة" لإضافة الأصول المتبرع بها</p>
          </div>

          <!-- Asset Items -->
          <div v-for="(asset, index) in selectedAssets" :key="index" class="p-3 bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 space-y-3">
            <div class="flex justify-between items-center">
              <span class="text-sm font-medium text-gray-600 dark:text-gray-400">مادة #{{ index + 1 }}</span>
              <button type="button" @click="removeAsset(index)" class="text-red-500 hover:text-red-700 text-sm">✕ حذف</button>
            </div>
            
            <div class="grid grid-cols-2 gap-3">
              <div>
                <label class="block text-xs text-gray-500 mb-1">نوع المادة</label>
                <select v-model="asset.asset_type_id" class="input-field text-sm">
                  <option value="">اختر...</option>
                  <optgroup v-for="cat in assetCategories" :key="cat.id" :label="cat.name_ar || cat.name">
                    <option v-for="t in getTypesForCategory(cat.id)" :key="t.id" :value="t.id">
                      {{ t.name_ar || t.name }}
                    </option>
                  </optgroup>
                </select>
              </div>
              <div>
                <label class="block text-xs text-gray-500 mb-1">الكمية</label>
                <input v-model.number="asset.quantity" type="number" min="1" class="input-field text-sm" />
              </div>
            </div>

            <div class="grid grid-cols-2 gap-3">
              <div>
                <label class="block text-xs text-gray-500 mb-1">الاسم / الوصف</label>
                <input v-model="asset.name" type="text" placeholder="مثل: لابتوب Dell" class="input-field text-sm" />
              </div>
              <div>
                <label class="block text-xs text-gray-500 mb-1">القيمة التقديرية (اختياري)</label>
                <input v-model.number="asset.unit_price" type="number" step="0.01" min="0" class="input-field text-sm" />
              </div>
            </div>
          </div>
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">
            وجهة الاستلام <span class="text-deficit">*</span>
          </label>
          <select v-model="form.container_id" required class="input-field">
            <option value="">اختر...</option>
            <optgroup label="💰 نقدي (كاش)">
              <option v-for="c in containers.filter(x => x.type === 'cash')" :key="c.id" :value="c.id">{{ c.name }} ({{ c.currency }})</option>
            </optgroup>
            <optgroup label="🏦 مصرفي (بنك)">
              <option v-for="c in containers.filter(x => x.type === 'bank')" :key="c.id" :value="c.id">{{ c.name }} ({{ c.currency }})</option>
            </optgroup>
          </select>
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">
            المشروع <span class="text-deficit">*</span>
          </label>
          <select v-model="form.project_id" required class="input-field">
            <option value="">اختر...</option>
            <option v-for="p in projects.filter(proj => {
              if (!form.container_id) return true;
              const container = containers.find(c => c.id === form.container_id);
              return !container || (proj.currency || 'LYD') === container.currency;
            })" :key="p.id" :value="p.id">{{ p.code }} - {{ p.name }}</option>
          </select>
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">
            البيان <span class="text-deficit">*</span>
          </label>
          <textarea v-model="form.description_ar" required rows="3" class="input-field"></textarea>
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">الموظف / العضو (للاشتراكات)</label>
          <select v-model="form.member_id" @change="onMemberChange" class="input-field">
            <option value="">-- غير محدد --</option>
            <optgroup label="👤 الموظفين">
              <option v-for="e in employees" :key="'emp-' + e.id" :value="e.id">
                {{ e.number }} - {{ e.name }} {{ e.position ? '(' + e.position + ')' : '' }}
              </option>
            </optgroup>
            <optgroup label="🤝 الأعضاء المنتسبين">
              <option v-for="m in members" :key="'mem-' + m.id" :value="m.id">
                {{ m.number }} - {{ m.name }} ({{ m.type }})
              </option>
            </optgroup>
          </select>
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">المتبرع</label>
          <input v-model="form.payee_payer" type="text" class="input-field" />
        </div>

        <div class="flex gap-3">
          <button type="submit" :disabled="loading" class="btn-primary disabled:opacity-50">
            {{ loading ? 'جاري الحفظ...' : 'حفظ' }}
          </button>
          <button type="button" @click="resetForm" class="btn-secondary">إعادة تعيين</button>
        </div>
      </form>
    </div>
  </div>
</template>
