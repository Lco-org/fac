<template>
  <div class="bg-gray-50 rounded-lg border border-gray-200 p-4">
    <div class="flex justify-between items-center mb-4">
      <h3 class="font-bold text-gray-700">المرفقات</h3>
      <div class="relative">
        <input 
          type="file" 
          ref="fileInput" 
          @change="handleFileUpload" 
          class="hidden" 
          multiple
        >
        <button 
          @click="$refs.fileInput.click()"
          class="bg-white border border-gray-300 text-gray-700 hover:bg-gray-50 px-3 py-1 rounded text-sm flex items-center shadow-sm"
          :disabled="uploading"
        >
          <span v-if="uploading" class="ml-2 animate-spin h-3 w-3 border-b-2 border-gray-700 rounded-full"></span>
          <span v-else class="ml-2 text-lg">+</span>
          إضافة مرفق
        </button>
      </div>
    </div>

    <!-- Error Message -->
    <div v-if="error" class="mb-3 text-red-600 text-sm bg-red-50 p-2 rounded">
      {{ error }}
    </div>

    <!-- Upload Progress -->
    <div v-if="uploading" class="mb-4">
      <div class="w-full bg-gray-200 rounded-full h-2.5">
        <div class="bg-blue-600 h-2.5 rounded-full" :style="{ width: uploadProgress + '%' }"></div>
      </div>
      <p class="text-xs text-gray-500 mt-1 text-center">{{ uploadProgress }}%</p>
    </div>

    <!-- List -->
    <div v-if="loading" class="text-center py-4 text-gray-500 text-sm">جاري التحميل...</div>
    
    <div v-else-if="attachments.length === 0" class="text-center py-4 text-gray-400 text-sm border-2 border-dashed border-gray-200 rounded">
      لا توجد مرفقات
    </div>

    <div v-else class="space-y-2">
      <div v-for="file in attachments" :key="file.id" class="flex justify-between items-center bg-white p-2 rounded border border-gray-100 shadow-sm">
        <div class="flex items-center overflow-hidden">
          <div class="p-2 bg-gray-100 rounded ml-3">
             <span class="text-xs font-bold text-gray-500">{{ getFileExtension(file.original_name) }}</span>
          </div>
          <div class="truncate">
            <a :href="getDownloadUrl(file.id)" target="_blank" class="text-sm text-blue-600 hover:underline block truncate" :title="file.original_name">
              {{ file.original_name }}
            </a>
            <span class="text-xs text-gray-400">{{ formatSize(file.file_size) }} • {{ formatDate(file.created_at) }}</span>
          </div>
        </div>
        <button 
          @click="deleteAttachment(file.id)" 
          class="text-gray-400 hover:text-red-500 p-1"
          title="حذف"
        >
          ×
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import attachmentApi from '@/api/attachments'

const props = defineProps({
  entityType: {
    type: String,
    required: true
  },
  entityId: {
    type: String,
    required: true
  }
})

const attachments = ref([])
const loading = ref(false)
const uploading = ref(false)
const uploadProgress = ref(0)
const error = ref(null)

const fetchAttachments = async () => {
  if (!props.entityId) return
  
  loading.value = true
  try {
    const response = await attachmentApi.list(props.entityType, props.entityId)
    attachments.value = response.data || []
  } catch (err) {
    console.error('Failed to list attachments', err)
  } finally {
    loading.value = false
  }
}

const handleFileUpload = async (event) => {
  const files = event.target.files
  if (!files.length) return

  uploading.value = true
  uploadProgress.value = 0
  error.value = null

  try {
    // Upload one by one (simplified)
    for (let i = 0; i < files.length; i++) {
        await attachmentApi.upload(
            props.entityType, 
            props.entityId, 
            files[i], 
            (progressEvent) => {
                uploadProgress.value = Math.round((progressEvent.loaded * 100) / progressEvent.total)
            }
        )
    }
    await fetchAttachments() // Refresh list
  } catch (err) {
    console.error('Upload failed', err)
    error.value = 'فشل رفع الملف. تأكد من أن الحجم المسموح به والصيغة صحيحة.'
  } finally {
    uploading.value = false
    uploadProgress.value = 0
    event.target.value = '' // Reset input
  }
}

const deleteAttachment = async (id) => {
  if (!confirm('هل أنت متأكد من حذف هذا المرفق؟')) return

  try {
    await attachmentApi.delete(id)
    attachments.value = attachments.value.filter(a => a.id !== id)
  } catch (err) {
    console.error('Delete failed', err)
    alert('فشل الحذف')
  }
}

const getDownloadUrl = (id) => {
    return attachmentApi.getDownloadURL(id)
}

const getFileExtension = (filename) => {
    return filename.split('.').pop().toUpperCase()
}

const formatSize = (bytes) => {
    if (bytes === 0) return '0 B'
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}

const formatDate = (dateString) => {
    return new Date(dateString).toLocaleDateString('ar-LY')
}

// Fetch when ID changes (e.g. navigation)
watch(() => props.entityId, (newVal) => {
    if (newVal) fetchAttachments()
})

onMounted(() => {
    fetchAttachments()
})
</script>
