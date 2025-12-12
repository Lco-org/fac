<script setup>
import { ref, onMounted, watch } from 'vue'
import api from '@/api/client'
import DeficitWarningModal from '@/components/DeficitWarningModal.vue'

const loading = ref(false)
const success = ref(false)
const error = ref('')
const showDeficitWarning = ref(false)
const deficitWarningData = ref(null)

// Salary Summary State
const summaryLoading = ref(false)
const salarySummary = ref(null)
const selectedMonth = ref(new Date().toISOString().slice(0, 7)) // YYYY-MM

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

// Fetch Salary Summary
const fetchSalarySummary = async () => {
  if (!form.value.employee_id || !selectedMonth.value) {
    salarySummary.value = null
    return
  }

  summaryLoading.value = true
  try {
    const res = await api.get(`/hr/salaries?month=${selectedMonth.value}&employee_id=${form.value.employee_id}`)
    if (res.data && res.data.length > 0) {
      salarySummary.value = res.data[0]
      // Auto-fill amount with Net Salary
      if (salarySummary.value.net_salary > 0) {
        form.value.amount = salarySummary.value.net_salary
      }
    } else {
      salarySummary.value = null
    }
  } catch (err) {
    console.error('Failed to fetch salary summary', err)
  } finally {
    summaryLoading.value = false
  }
}

// Watchers
watch([() => form.value.employee_id, selectedMonth], () => {
  if (form.value.employee_id) {
    fetchSalarySummary()
  } else {
    salarySummary.value = null
  }
})

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
  salarySummary.value = null
}
</script>

<template>
  <div>
    <h1 class="text-3xl font-bold mb-6">صرف مرتبات</h1>

    <div class="card max-w-4xl">
      <div v-if="success" class="bg-surplus-light border border-surplus text-surplus-dark px-4 py-3 rounded-lg mb-4">
        ✅ تمت العملية بنجاح
      </div>

      <div v-if="error" class="bg-deficit-light border border-deficit text-deficit-dark px-4 py-3 rounded-lg mb-4">
        ❌ {{ error }}
      </div>

      <form @submit.prevent="handleSubmit" class="space-y-6">
        
        <!-- Employee & Month Selection Row -->
        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">الشهر (لحساب المستحقات)</label>
              <input type="month" v-model="selectedMonth" class="input-field" />
            </div>

            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">الموظف <span class="text-deficit">*</span></label>
              <select v-model="form.employee_id" required class="input-field">
                <option value="">اختر الموظف...</option>
                <option v-for="e in employees" :key="e.id" :value="e.id">
                  {{ e.employee_number }} - {{ e.name }}
                </option>
              </select>
            </div>
        </div>

        <!-- Mini Summary Table -->
        <div v-if="summaryLoading" class="text-center py-4 text-gray-500">جاري احتساب المستحقات...</div>
        <div v-else-if="salarySummary" class="bg-gray-50 p-4 rounded-lg border border-gray-200">
            <h3 class="text-sm font-bold text-gray-700 mb-3">ملخص مستحقات شهر {{ selectedMonth }}:</h3>
            <div class="grid grid-cols-2 md:grid-cols-4 gap-4 text-center">
                <div class="bg-white p-2 rounded shadow-sm">
                    <div class="text-xs text-gray-500">الراتب الأساسي</div>
                    <div class="font-bold text-gray-800">{{ salarySummary.base_salary.toLocaleString() }}</div>
                </div>
                <div class="bg-white p-2 rounded shadow-sm border-r-2 border-green-400">
                    <div class="text-xs text-gray-500">(+) المكافآت</div>
                    <div class="font-bold text-green-600">{{ salarySummary.total_bonuses.toLocaleString() }}</div>
                </div>
                <div class="bg-white p-2 rounded shadow-sm border-r-2 border-red-400">
                    <div class="text-xs text-gray-500">(-) الخصومات</div>
                    <div class="font-bold text-red-600">{{ salarySummary.total_deductions.toLocaleString() }}</div>
                </div>
                <div class="bg-white p-2 rounded shadow-sm ring-2 ring-blue-100">
                    <div class="text-xs text-gray-500">صافي الراتب</div>
                    <div class="font-bold text-blue-800 text-lg">{{ salarySummary.net_salary.toLocaleString() }}</div>
                </div>
            </div>
        </div>

        <!-- Payment Details -->
        <div class="border-t pt-6 grid grid-cols-1 md:grid-cols-2 gap-6">
            <div>
            <label class="block text-sm font-medium text-gray-700 mb-2">المبلغ المراد صرفه <span class="text-deficit">*</span></label>
            <input v-model.number="form.amount" type="number" step="0.01" required class="input-field font-bold" />
            <p class="text-xs text-gray-500 mt-1">يتم تعبئته تلقائياً بصافي الراتب</p>
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
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">البيان <span class="text-deficit">*</span></label>
          <textarea v-model="form.description_ar" required rows="3" class="input-field" placeholder="مسير رواتب شهر..."></textarea>
        </div>

        <div class="flex gap-3">
          <button type="submit" :disabled="loading" class="btn-primary disabled:opacity-50">
            {{ loading ? 'جاري الحفظ...' : 'حفظ وصرف' }}
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
