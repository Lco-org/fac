<script setup>
import { ref, onMounted } from 'vue'
import api from '@/api/client'

const loading = ref(true)
const saving = ref(false)
const widgets = ref([])
const error = ref(null)
const draggedIndex = ref(null)

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
  } catch (err) {
    console.error('Failed to save widget:', err)
    alert('فشل حفظ التغييرات')
  } finally {
    saving.value = false
  }
}

const saveBulkOrder = async () => {
  try {
    saving.value = true
    const updates = widgets.value.map((widget, index) => ({
      id: widget.id,
      display_order: index + 1
    }))
    
    await api.post('/admin/dashboard-widgets/bulk-order', updates)
    
    // Update local display_order values
    widgets.value.forEach((widget, index) => {
      widget.display_order = index + 1
    })
  } catch (err) {
    console.error('Failed to save order:', err)
    alert('فشل حفظ الترتيب')
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

// Drag and drop handlers
const handleDragStart = (index) => {
  draggedIndex.value = index
}

const handleDragOver = (e, index) => {
  e.preventDefault()
  if (draggedIndex.value === null || draggedIndex.value === index) return
  
  // Visual feedback
  e.currentTarget.classList.add('drag-over')
}

const handleDragLeave = (e) => {
  e.currentTarget.classList.remove('drag-over')
}

const handleDrop = (e, dropIndex) => {
  e.preventDefault()
  e.currentTarget.classList.remove('drag-over')
  
  if (draggedIndex.value === null || draggedIndex.value === dropIndex) return
  
  // Reorder array
  const draggedItem = widgets.value[draggedIndex.value]
  widgets.value.splice(draggedIndex.value, 1)
  widgets.value.splice(dropIndex, 0, draggedItem)
  
  draggedIndex.value = null
  
  // Save new order
  saveBulkOrder()
}

const handleDragEnd = () => {
  draggedIndex.value = null
  // Remove any lingering drag-over classes
  document.querySelectorAll('.drag-over').forEach(el => {
    el.classList.remove('drag-over')
  })
}

// Up/Down button handlers
const moveUp = (index) => {
  if (index === 0) return
  
  const temp = widgets.value[index]
  widgets.value[index] = widgets.value[index - 1]
  widgets.value[index - 1] = temp
  
  saveBulkOrder()
}

const moveDown = (index) => {
  if (index === widgets.value.length - 1) return
  
  const temp = widgets.value[index]
  widgets.value[index] = widgets.value[index + 1]
  widgets.value[index + 1] = temp
  
  saveBulkOrder()
}
</script>

<template>
  <div class="p-6">
    <div class="flex justify-between items-center mb-6">
      <h1 class="text-2xl font-bold text-gray-800">إعدادات لوحة التحكم</h1>
      <div class="text-sm text-gray-600">
        💡 اسحب الصفوف لإعادة الترتيب أو استخدم الأسهم
      </div>
    </div>

    <div v-if="loading" class="flex justify-center py-12">
      <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600"></div>
    </div>

    <div v-else-if="error" class="bg-red-50 text-red-700 p-4 rounded mb-6">
      {{ error }}
    </div>

    <div v-else class="bg-white rounded-lg shadow overflow-hidden">
      <table class="min-w-full divide-y divide-gray-200 responsive-table">
        <thead class="bg-gray-50">
          <tr>
            <th scope="col" class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
              الترتيب
            </th>
            <th scope="col" class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
              القسم
            </th>
            <th scope="col" class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
              الحالة
            </th>
            <th scope="col" class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
              الأدوار المسموح لها
            </th>
          </tr>
        </thead>
        <tbody class="bg-white divide-y divide-gray-200">
          <tr 
            v-for="(widget, index) in widgets" 
            :key="widget.id"
            draggable="true"
            @dragstart="handleDragStart(index)"
            @dragover="handleDragOver($event, index)"
            @dragleave="handleDragLeave"
            @drop="handleDrop($event, index)"
            @dragend="handleDragEnd"
            class="hover:bg-gray-50 transition-colors cursor-move"
            :class="{ 'opacity-50': draggedIndex === index }"
          >
            <!-- Order Controls -->
            <td class="px-6 py-4 whitespace-nowrap" data-label="الترتيب">
              <div class="flex items-center gap-2">
                <!-- Drag Handle -->
                <div class="text-gray-400 hover:text-gray-600 cursor-grab active:cursor-grabbing">
                  <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
                    <path d="M7 2a2 2 0 1 0 .001 4.001A2 2 0 0 0 7 2zm0 6a2 2 0 1 0 .001 4.001A2 2 0 0 0 7 8zm0 6a2 2 0 1 0 .001 4.001A2 2 0 0 0 7 14zm6-8a2 2 0 1 0-.001-4.001A2 2 0 0 0 13 6zm0 2a2 2 0 1 0 .001 4.001A2 2 0 0 0 13 8zm0 6a2 2 0 1 0 .001 4.001A2 2 0 0 0 13 14z"></path>
                  </svg>
                </div>
                
                <!-- Up/Down Buttons -->
                <div class="flex flex-col gap-1">
                  <button
                    @click="moveUp(index)"
                    :disabled="index === 0"
                    class="p-1 rounded hover:bg-gray-200 disabled:opacity-30 disabled:cursor-not-allowed transition-colors"
                    title="تحريك لأعلى"
                  >
                    <svg class="w-4 h-4 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 15l7-7 7 7"></path>
                    </svg>
                  </button>
                  <button
                    @click="moveDown(index)"
                    :disabled="index === widgets.length - 1"
                    class="p-1 rounded hover:bg-gray-200 disabled:opacity-30 disabled:cursor-not-allowed transition-colors"
                    title="تحريك لأسفل"
                  >
                    <svg class="w-4 h-4 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"></path>
                    </svg>
                  </button>
                </div>
                
                <!-- Order Number Badge -->
                <div class="bg-primary-100 text-primary-800 px-2 py-1 rounded-full text-xs font-semibold">
                  {{ index + 1 }}
                </div>
              </div>
            </td>

            <!-- Widget Name -->
            <td class="px-6 py-4 whitespace-nowrap" data-label="القسم">
              <div class="text-sm font-medium text-gray-900">{{ widget.widget_name }}</div>
              <div class="text-xs text-gray-500">{{ widget.widget_key }}</div>
            </td>

            <!-- Enable/Disable Toggle -->
            <td class="px-6 py-4 whitespace-nowrap" data-label="الحالة">
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

            <!-- Allowed Roles -->
            <td class="px-6 py-4" data-label="الأدوار المسموح لها">
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
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<style scoped>
.drag-over {
  border-top: 3px solid #3b82f6;
  background-color: #eff6ff;
}

tr[draggable="true"]:active {
  cursor: grabbing;
}
</style>
