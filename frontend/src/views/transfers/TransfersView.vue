<script setup>
import { ref, onMounted, computed } from 'vue'
import api from '@/api/client'

const loading = ref(false)
const success = ref(false)
const error = ref('')

const form = ref({
  from_container_id: '',
  to_container_id: '',
  amount: '',
  currency: 'LYD',
  description: '',
  reference_number: '',
})

const containers = ref([])

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
        currency: c.currency,
        balance: c.balance
      }))
    }
  } catch (err) {
    console.error('Failed to load containers:', err)
    error.value = 'فشل في تحميل الحسابات'
  }
})

// Filter containers based on source selection
const availableDestinations = computed(() => {
  if (!form.value.from_container_id) {
    return []
  }
  
  const sourceContainer = containers.value.find(c => c.id === form.value.from_container_id)
  if (!sourceContainer) {
    return []
  }
  
  // Filter: same currency and different container
  return containers.value.filter(c => 
    c.currency === sourceContainer.currency && 
    c.id !== form.value.from_container_id
  )
})

// Update currency when source changes
const onSourceChange = () => {
  const sourceContainer = containers.value.find(c => c.id === form.value.from_container_id)
  if (sourceContainer) {
    form.value.currency = sourceContainer.currency
    // Reset destination if it's no longer valid
    if (form.value.to_container_id) {
      const destContainer = containers.value.find(c => c.id === form.value.to_container_id)
      if (!destContainer || destContainer.currency !== sourceContainer.currency) {
        form.value.to_container_id = ''
      }
    }
  }
}

const handleSubmit = async () => {
  error.value = ''
  success.value = false
  loading.value = true

  try {
    const response = await api.post('/transfers', form.value)
    success.value = true
    
    // Update container balances locally
    const sourceIdx = containers.value.findIndex(c => c.id === form.value.from_container_id)
    const destIdx = containers.value.findIndex(c => c.id === form.value.to_container_id)
    if (sourceIdx !== -1) {
      containers.value[sourceIdx].balance -= parseFloat(form.value.amount)
    }
    if (destIdx !== -1) {
      containers.value[destIdx].balance += parseFloat(form.value.amount)
    }
    
    resetForm()
  } catch (err) {
    error.value = err.response?.data?.error || 'فشل في إنشاء التحويل'
  } finally {
    loading.value = false
  }
}

const resetForm = () => {
  form.value = {
    from_container_id: '',
    to_container_id: '',
    amount: '',
    currency: 'LYD',
    description: '',
    reference_number: '',
  }
}

// Get container display info
const getContainerInfo = (containerId) => {
  const container = containers.value.find(c => c.id === containerId)
  if (!container) return ''
  return `${container.name} (${container.currency}) - رصيد: ${container.balance.toFixed(2)}`
}
</script>

<template>
  <div>
    <h1 class="text-3xl font-bold mb-6">التحويلات بين الحسابات</h1>

    <div class="card max-w-2xl">
      <div v-if="success" class="bg-surplus-light border border-surplus text-surplus-dark px-4 py-3 rounded-lg mb-4">
        ✅ تم التحويل بنجاح
      </div>

      <div v-if="error" class="bg-deficit-light border border-deficit text-deficit-dark px-4 py-3 rounded-lg mb-4">
        ❌ {{ error }}
      </div>

      <form @submit.prevent="handleSubmit" class="space-y-4">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">
            من حساب (المصدر) <span class="text-deficit">*</span>
          </label>
          <select 
            v-model="form.from_container_id" 
            @change="onSourceChange"
            required 
            class="input-field"
          >
            <option value="">اختر الحساب المصدر...</option>
            <optgroup label="نقدي (كاش)">
              <option 
                v-for="container in containers.filter(c => c.type === 'cash')" 
                :key="container.id" 
                :value="container.id"
              >
                {{ container.name }} ({{ container.currency }}) - رصيد: {{ container.balance.toFixed(2) }}
              </option>
            </optgroup>
            <optgroup label="مصرفي (بنك)">
              <option 
                v-for="container in containers.filter(c => c.type === 'bank')" 
                :key="container.id" 
                :value="container.id"
              >
                {{ container.name }} ({{ container.currency }}) - رصيد: {{ container.balance.toFixed(2) }}
              </option>
            </optgroup>
          </select>
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">
            إلى حساب (الوجهة) <span class="text-deficit">*</span>
          </label>
          <select 
            v-model="form.to_container_id" 
            required 
            class="input-field"
            :disabled="!form.from_container_id"
          >
            <option value="">
              {{ form.from_container_id ? 'اختر الحساب الوجهة...' : 'اختر الحساب المصدر أولاً' }}
            </option>
            <optgroup label="نقدي (كاش)" v-if="availableDestinations.filter(c => c.type === 'cash').length > 0">
              <option 
                v-for="container in availableDestinations.filter(c => c.type === 'cash')" 
                :key="container.id" 
                :value="container.id"
              >
                {{ container.name }} ({{ container.currency }}) - رصيد: {{ container.balance.toFixed(2) }}
              </option>
            </optgroup>
            <optgroup label="مصرفي (بنك)" v-if="availableDestinations.filter(c => c.type === 'bank').length > 0">
              <option 
                v-for="container in availableDestinations.filter(c => c.type === 'bank')" 
                :key="container.id" 
                :value="container.id"
              >
                {{ container.name }} ({{ container.currency }}) - رصيد: {{ container.balance.toFixed(2) }}
              </option>
            </optgroup>
          </select>
          <p class="text-sm text-gray-500 mt-1">
            ملاحظة: يمكن التحويل فقط بين حسابات بنفس العملة
          </p>
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">
            المبلغ <span class="text-deficit">*</span>
          </label>
          <input 
            v-model.number="form.amount" 
            type="number" 
            step="0.01" 
            required 
            class="input-field" 
            placeholder="0.00" 
          />
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">
            العملة
          </label>
          <input 
            :value="form.currency" 
            type="text" 
            readonly 
            disabled
            class="input-field bg-gray-100 cursor-not-allowed" 
          />
          <p class="text-sm text-gray-500 mt-1">
            العملة تحدد تلقائياً حسب الحساب المصدر
          </p>
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">
            البيان <span class="text-deficit">*</span>
          </label>
          <textarea 
            v-model="form.description" 
            required 
            rows="3" 
            class="input-field" 
            placeholder="وصف التحويل..."
          ></textarea>
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">
            رقم المرجع
          </label>
          <input 
            v-model="form.reference_number" 
            type="text" 
            class="input-field" 
            placeholder="رقم الإشعار أو المرجع (اختياري)" 
          />
        </div>

        <div class="flex gap-3">
          <button 
            type="submit" 
            :disabled="loading" 
            class="btn-primary disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <span v-if="loading">جاري التحويل...</span>
            <span v-else>تنفيذ التحويل</span>
          </button>
          <button 
            type="button" 
            @click="resetForm" 
            class="btn-secondary"
          >
            إعادة تعيين
          </button>
        </div>
      </form>
    </div>
  </div>
</template>
