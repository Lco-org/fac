<script setup>
import { ref, onMounted } from 'vue'
import api from '@/api/client'
import DeficitWarningModal from '@/components/DeficitWarningModal.vue'

const loading = ref(false)
const success = ref(false)
const error = ref('')
const showDeficitWarning = ref(false)
const deficitWarningData = ref(null)

const form = ref({
  type: 'disbursement_salary',
  amount: '',
  currency: 'LYD',
  exchange_rate: 1.0,
  container_id: '',
  project_id: '',
  employee_id: '',
  description: '',
  description_ar: '',
  payee_payer: '',
  reference_number: '',
  confirm_deficit: false,
})

const containers = ref([])
const projects = ref([])
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
        name: p.name_ar || p.name
      }))
    }

    // Load employees for salary selection
    const employeesRes = await api.get('/employees')
    if (employeesRes.data && Array.isArray(employeesRes.data)) {
      employees.value = employeesRes.data.filter(e => e.is_active).map(e => ({
        id: e.id,
        name: e.full_name,
        employee_number: e.employee_number,
        base_salary: e.base_salary || 0
      }))
    }
  } catch (err) {
    console.error('Failed to load data:', err)
  }
})

// Auto-fill amount when employee selected
const onEmployeeChange = () => {
  const selectedEmployee = employees.value.find(e => e.id === form.value.employee_id)
  if (selectedEmployee && selectedEmployee.base_salary > 0) {
    form.value.amount = selectedEmployee.base_salary
    // form.value.payee_payer = selectedEmployee.name // Removed auto-fill
  }
}

const handleSubmit = async () => {
  error.value = ''
  success.value = false
  loading.value = true

  // Validate employee selection
  if (!form.value.employee_id) {
    error.value = 'يجب اختيار الموظف لصرف المرتب'
    loading.value = false
    return
  }

  try {
    const payload = {
      ...form.value,
      description: form.value.description_ar || form.value.description
    }
    const response = await api.post('/transactions', payload)

    if (response.status === 202 && response.data.deficit_warning) {
      deficitWarningData.value = response.data
      showDeficitWarning.value = true
      loading.value = false
      return
    }

    success.value = true
    
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

const handleDeficitConfirm = async () => {
  showDeficitWarning.value = false
  form.value.confirm_deficit = true
  await handleSubmit()
  form.value.confirm_deficit = false
}

const handleDeficitCancel = () => {
  showDeficitWarning.value = false
  deficitWarningData.value = null
}

const resetForm = () => {
  form.value = {
    type: 'disbursement_salary',
    amount: '',
    currency: 'LYD',
    exchange_rate: 1.0,
    container_id: '',
    project_id: '',
    employee_id: '',
    description: '',
    description_ar: '',
    payee_payer: '',
    reference_number: '',
    confirm_deficit: false,
  }
}
</script>

<template>
  <div>
    <h1 class="text-3xl font-bold mb-6">صرف مرتبات</h1>

    <div class="card max-w-2xl">
      <div v-if="success" class="bg-surplus-light border border-surplus text-surplus-dark px-4 py-3 rounded-lg mb-4">
        ✅ تمت العملية بنجاح
      </div>

      <div v-if="error" class="bg-deficit-light border border-deficit text-deficit-dark px-4 py-3 rounded-lg mb-4">
        ❌ {{ error }}
      </div>

      <form @submit.prevent="handleSubmit" class="space-y-4">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">الموظف <span class="text-deficit">*</span></label>
          <select v-model="form.employee_id" @change="onEmployeeChange" required class="input-field">
            <option value="">اختر الموظف...</option>
            <option v-for="e in employees" :key="e.id" :value="e.id">
              {{ e.employee_number }} - {{ e.name }}
              <template v-if="e.base_salary > 0"> (الراتب: {{ e.base_salary }})</template>
            </option>
          </select>
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">المبلغ <span class="text-deficit">*</span></label>
          <input v-model.number="form.amount" type="number" step="0.01" required class="input-field" />
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">طريقة الدفع <span class="text-deficit">*</span></label>
          <select v-model="form.container_id" required class="input-field">
            <option value="">اختر...</option>
            <optgroup label="نقدي">
              <option v-for="c in containers.filter(x => x.type === 'cash')" :key="c.id" :value="c.id">{{ c.name }} ({{ c.currency }})</option>
            </optgroup>
            <optgroup label="🏦 مصرفي (بنك)">
              <option v-for="c in containers.filter(x => x.type === 'bank')" :key="c.id" :value="c.id">{{ c.name }} ({{ c.currency }})</option>
            </optgroup>
          </select>
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">المشروع <span class="text-deficit">*</span></label>
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
          <label class="block text-sm font-medium text-gray-700 mb-2">ملاحظات إضافية</label>
          <input v-model="form.payee_payer" type="text" class="input-field" placeholder="أي ملاحظات أخرى..." />
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">البيان <span class="text-deficit">*</span></label>
          <textarea v-model="form.description_ar" required rows="3" class="input-field" placeholder="مسير رواتب شهر..."></textarea>
        </div>

        <div class="flex gap-3">
          <button type="submit" :disabled="loading" class="btn-primary disabled:opacity-50">
            {{ loading ? 'جاري الحفظ...' : 'حفظ' }}
          </button>
          <button type="button" @click="resetForm" class="btn-secondary">إعادة تعيين</button>
        </div>
      </form>
    </div>

    <DeficitWarningModal
      :show="showDeficitWarning"
      :message="deficitWarningData?.message"
      :current-balance="0"
      :amount="form.amount"
      :new-balance="0"
      :currency="form.currency"
      @confirm="handleDeficitConfirm"
      @cancel="handleDeficitCancel"
    />
  </div>
</template>
