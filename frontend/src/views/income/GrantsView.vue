<script setup>
import { ref, onMounted } from 'vue'
import api from '@/api/client'

const loading = ref(false)
const success = ref(false)
const error = ref('')

const form = ref({
  type: 'income_grant',
  amount: '',
  currency: 'LYD',
  exchange_rate: 1.0,
  container_id: '',
  project_id: '',
  description: '',
  description_ar: '',
  payee_payer: '',
  reference_number: '',
  confirm_deficit: false,
})

const containers = ref([])
const projects = ref([])

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
  } catch (err) {
    console.error('Failed to load data:', err)
    error.value = 'فشل في تحميل البيانات'
  }
})

const handleSubmit = async () => {
  error.value = ''
  success.value = false
  loading.value = true

  try {
    // Copy description_ar to description for backend compatibility
    const payload = {
      ...form.value,
      description: form.value.description_ar || form.value.description
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
    type: 'income_grant',
    amount: '',
    currency: 'LYD',
    exchange_rate: 1.0,
    container_id: '',
    project_id: '',
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
    <h1 class="text-3xl font-bold mb-6">منح ومشاريع</h1>

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
          <input v-model.number="form.amount" type="number" step="0.01" required class="input-field" placeholder="0.00" />
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">
            وجهة الاستلام <span class="text-deficit">*</span>
          </label>
          <select v-model="form.container_id" required class="input-field">
            <option value="">اختر...</option>
            <optgroup label="نقدي (كاش)">
              <option v-for="container in containers.filter(c => c.type === 'cash')" :key="container.id" :value="container.id">
                {{ container.name }} ({{ container.currency }})
              </option>
            </optgroup>
            <optgroup label="مصرفي (بنك)">
              <option v-for="container in containers.filter(c => c.type === 'bank')" :key="container.id" :value="container.id">
                {{ container.name }} ({{ container.currency }})
              </option>
            </optgroup>
          </select>
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">
            المشروع <span class="text-deficit">*</span>
          </label>
          <select v-model="form.project_id" required class="input-field">
            <option value="">اختر المشروع...</option>
            <option v-for="project in projects" :key="project.id" :value="project.id">
              {{ project.code }} - {{ project.name }}
            </option>
          </select>
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">
            البيان <span class="text-deficit">*</span>
          </label>
          <textarea v-model="form.description_ar" required rows="3" class="input-field" placeholder="وصف المنحة..."></textarea>
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">
            المانح
          </label>
          <input v-model="form.payee_payer" type="text" class="input-field" placeholder="اسم المانح" />
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">
            رقم المرجع
          </label>
          <input v-model="form.reference_number" type="text" class="input-field" placeholder="رقم الإشعار أو المرجع" />
        </div>

        <div class="flex gap-3">
          <button type="submit" :disabled="loading" class="btn-primary disabled:opacity-50 disabled:cursor-not-allowed">
            <span v-if="loading">جاري الحفظ...</span>
            <span v-else">حفظ العملية</span>
          </button>
          <button type="button" @click="resetForm" class="btn-secondary">إعادة تعيين</button>
        </div>
      </form>
    </div>
  </div>
</template>
