<script setup>
import { ref, onMounted } from 'vue'
import api from '@/api/client'

const loading = ref(true)
const saving = ref(false)
const widgets = ref([])
const error = ref(null)

const availableRoles = [
  { value: 'admin', label: 'مدير النظام' },
  { value: 'finance', label: 'المسؤول المالي' },
  { value: 'auditor', label: 'المراجع' }
]

onMounted(async () => {
  await loadWidgets()
})

const loadWidgets = async () => {
  try {
    loading.value = true
    const response = await api.get('/admin/dashboard-widgets')
    widgets.value = response.data.sort((a, b) => a.display_order - b.display_order)
  } catch (err) {
    console.error('Failed to load widgets:', err)
    error.value = 'فشل تحميل إعدادات لوحة التحكم'
  } finally {
    loading.value = false
  }
}

const saveWidget = async (widget) => {
  try {
    saving.value = true
    await api.put(`/admin/dashboard-widgets/${widget.id}`, {
      is_enabled: widget.is_enabled,
      allowed_roles: widget.allowed_roles,
      display_order: parseInt(widget.display_order)
    })
    // Show success feedback (could be a toast)
    console.log('Widget saved successfully')
  } catch (err) {
    console.error('Failed to save widget:', err)
    alert('فشل حفظ التغييرات')
  } finally {
    saving.value = false
  }
}

const toggleRole = (widget, roleValue) => {
  const index = widget.allowed_roles.indexOf(roleValue)
  if (index === -1) {
    widget.allowed_roles.push(roleValue)
  } else {
    widget.allowed_roles.splice(index, 1)
  }
  saveWidget(widget)
}

const updateOrder = (widget) => {
  saveWidget(widget)
  // Re-sort locally to reflect changes immediately
  widgets.value.sort((a, b) => a.display_order - b.display_order)
}
</script>

<template>
  <div class="p-6">
    <h1 class="text-2xl font-bold mb-6 text-gray-800">إعدادات لوحة التحكم</h1>

    <div v-if="loading" class="flex justify-center py-12">
      <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600"></div>
    </div>

    <div v-else-if="error" class="bg-red-50 text-red-700 p-4 rounded mb-6">
      {{ error }}
    </div>

    <div v-else class="bg-white rounded-lg shadow overflow-hidden">
      <table class="min-w-full divide-y divide-gray-200">
        <thead class="bg-gray-50">
          <tr>
            <th scope="col" class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
              القسم
            </th>
            <th scope="col" class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
              الحالة
            </th>
            <th scope="col" class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
              الأدوار المسموح لها
            </th>
            <th scope="col" class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
              الترتيب
            </th>
          </tr>
        </thead>
        <tbody class="bg-white divide-y divide-gray-200">
          <tr v-for="widget in widgets" :key="widget.id">
            <td class="px-6 py-4 whitespace-nowrap">
              <div class="text-sm font-medium text-gray-900">{{ widget.widget_name }}</div>
              <div class="text-xs text-gray-500">{{ widget.widget_key }}</div>
            </td>
            <td class="px-6 py-4 whitespace-nowrap">
              <button 
                @click="widget.is_enabled = !widget.is_enabled; saveWidget(widget)"
                class="relative inline-flex flex-shrink-0 h-6 w-11 border-2 border-transparent rounded-full cursor-pointer transition-colors ease-in-out duration-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500"
                :class="widget.is_enabled ? 'bg-green-600' : 'bg-gray-200'"
              >
                <span 
                  class="pointer-events-none inline-block h-5 w-5 rounded-full bg-white shadow transform ring-0 transition ease-in-out duration-200"
                  :class="widget.is_enabled ? 'translate-x-0' : '-translate-x-5'"
                ></span>
              </button>
            </td>
            <td class="px-6 py-4">
              <div class="flex flex-wrap gap-2">
                <button
                  v-for="role in availableRoles"
                  :key="role.value"
                  @click="toggleRole(widget, role.value)"
                  class="px-3 py-1 rounded-full text-xs font-medium transition-colors"
                  :class="widget.allowed_roles.includes(role.value) 
                    ? 'bg-blue-100 text-blue-800 hover:bg-blue-200' 
                    : 'bg-gray-100 text-gray-600 hover:bg-gray-200'"
                >
                  {{ role.label }}
                </button>
              </div>
            </td>
            <td class="px-6 py-4 whitespace-nowrap">
              <input 
                type="number" 
                v-model="widget.display_order" 
                @change="updateOrder(widget)"
                class="w-16 rounded border-gray-300 text-center text-sm focus:ring-primary-500 focus:border-primary-500"
                min="1"
              >
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
