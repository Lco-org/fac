<script setup>
import { ref, onMounted } from 'vue'
import api from '@/api/client'

const loading = ref(false)
const success = ref(false)
const error = ref('')

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
        name: p.name_ar || p.name
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
    const response = await api.post('/transactions', payload)
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
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">
            المبلغ <span class="text-deficit">*</span>
          </label>
          <input v-model.number="form.amount" type="number" step="0.01" required class="input-field" />
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
            <option v-for="p in projects" :key="p.id" :value="p.id">{{ p.code }} - {{ p.name }}</option>
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
