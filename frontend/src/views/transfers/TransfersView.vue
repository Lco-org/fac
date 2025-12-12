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
        balance: c.balance,
        bank_name: c.bank_name_ar || c.bank_name
      }))
    }
  } catch (err) {
    console.error('Failed to load containers:', err)
    error.value = 'فشل في تحميل الحسابات'
  }
})

// Computed helper for selected containers
const selectedSource = computed(() => 
  containers.value.find(c => c.id === form.value.from_container_id)
)

const selectedDest = computed(() => 
  containers.value.find(c => c.id === form.value.to_container_id)
)

// Computed projected balance for destination
const projectedDestBalance = computed(() => {
  if (!selectedDest.value || !form.value.amount) return null
  return selectedDest.value.balance + parseFloat(form.value.amount)
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
    
    // Open receipt in new tab for the debit transaction
    if (response.data.debit_transaction && response.data.debit_transaction.id) {
      const receiptUrl = `/receipt/${response.data.debit_transaction.id}`
      window.open(receiptUrl, '_blank')
    }
    
    // Update container balances locally (OPTIMISTIC UI UPDATE)
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
        <!-- Source Account -->
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
              <option v-for="c in containers.filter(c => c.type === 'cash')" :key="c.id" :value="c.id">
                {{ c.name }} ({{ c.currency }})
              </option>
            </optgroup>
            <optgroup label="مصرفي (بنك)">
              <option v-for="c in containers.filter(c => c.type === 'bank')" :key="c.id" :value="c.id">
                {{ c.name }} ({{ c.currency }})
              </option>
            </optgroup>
          </select>
          
          <!-- Detailed Source Info -->
          <div v-if="selectedSource" class="mt-2 bg-blue-50 p-3 rounded border border-blue-100 text-sm grid grid-cols-2 gap-2">
            <div>
              <span class="text-gray-500 block text-xs">اسم المصرف/الخزينة:</span>
              <span class="font-bold text-gray-800">{{ selectedSource.bank_name || selectedSource.name }}</span>
            </div>
            <div>
              <span class="text-gray-500 block text-xs">الرصيد الحالي:</span>
              <span :class="selectedSource.balance < 0 ? 'text-red-600' : 'text-green-600'" class="font-bold">
                {{ selectedSource.balance.toLocaleString() }} {{ selectedSource.currency }}
                <span v-if="selectedSource.balance < 0" class="text-xs mr-1 bg-red-100 text-red-800 px-1 rounded">(عجز)</span>
              </span>
            </div>

          </div>
        </div>

        <!-- Destination Account -->
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
            <option v-for="c in availableDestinations" :key="c.id" :value="c.id">
              {{ c.name }} ({{ c.currency }})
            </option>
          </select>
          
          <!-- Detailed Destination Info -->
          <div v-if="selectedDest" class="mt-2 bg-green-50 p-3 rounded border border-green-100 text-sm grid grid-cols-2 gap-2">
             <div>
              <span class="text-gray-500 block text-xs">اسم المصرف/الخزينة:</span>
              <span class="font-bold text-gray-800">{{ selectedDest.bank_name || selectedDest.name }}</span>
            </div>
            <div>
              <span class="text-gray-500 block text-xs">الرصيد الحالي:</span>
              <span :class="selectedDest.balance < 0 ? 'text-red-600' : 'text-green-600'" class="font-bold">
                {{ selectedDest.balance.toLocaleString() }} {{ selectedDest.currency }}
                <span v-if="selectedDest.balance < 0" class="text-xs mr-1 bg-red-100 text-red-800 px-1 rounded">(عجز)</span>
              </span>
            </div>

          </div>
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">
            المبلغ <span class="text-deficit">*</span>
          </label>
          <input v-model.number="form.amount" type="number" step="0.01" required class="input-field" placeholder="0.00" />
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">البيان <span class="text-deficit">*</span></label>
          <textarea v-model="form.description" required rows="3" class="input-field" placeholder="وصف التحويل..."></textarea>
        </div>

        <div>
           <label class="block text-sm font-medium text-gray-700 mb-2">رقم المرجع</label>
           <input v-model="form.reference_number" type="text" class="input-field" placeholder="رقم الإشعار أو المرجع (اختياري)" />
        </div>

        <!-- Transfer Summary Box -->
        <div v-if="form.amount && selectedSource && selectedDest" class="bg-gradient-to-l from-blue-50 to-white p-4 rounded-lg border-r-4 border-blue-500 shadow-sm my-4">
          <h3 class="font-bold text-blue-800 mb-2 text-sm">ملخص العملية قبل التنفيذ:</h3>
          <p class="text-sm text-gray-700 leading-relaxed">
            سيتم تحويل مبلغ <span class="font-bold text-blue-700">{{ form.amount.toLocaleString() }} {{ form.currency }}</span>
            من <span class="font-bold">{{ selectedSource.name }}</span>
            إلى <span class="font-bold">{{ selectedDest.name }}</span>.
          </p>
          <div class="mt-2 text-sm pt-2 border-t border-blue-100">
             <span class="text-gray-500">الرصيد المتوقع في {{ selectedDest.name }}:</span>
             <span class="font-bold text-green-700 mr-2">{{ projectedDestBalance?.toLocaleString() }} {{ form.currency }}</span>
             <span class="text-xs text-gray-400 mr-1">(زيادة +{{ form.amount }})</span>
          </div>
        </div>

        <div class="flex gap-3">
          <button type="submit" :disabled="loading" class="btn-primary disabled:opacity-50 disabled:cursor-not-allowed w-full md:w-auto">
            <span v-if="loading">جاري التحويل...</span>
            <span v-else>تنفيذ التحويل 💸</span>
          </button>
          <button type="button" @click="resetForm" class="btn-secondary">إعادة تعيين</button>
        </div>
      </form>
    </div>
  </div>
</template>
