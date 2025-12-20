<template>
  <div class="p-6">
    <div class="flex justify-between items-center mb-6">
      <h1 class="text-2xl font-bold text-gray-800">النسخ الاحتياطي</h1>
      <button 
        @click="createBackup"
        class="bg-green-600 text-white px-4 py-2 rounded hover:bg-green-700 flex items-center"
        :disabled="creating"
      >
        <span v-if="creating" class="animate-spin h-4 w-4 border-b-2 border-white rounded-full ml-2"></span>
        نسخة احتياطية جديدة
      </button>
    </div>

    <div v-if="loading" class="text-center py-10">جاري التحميل...</div>

    <div v-else class="bg-white rounded-lg shadow overflow-hidden">
      <table class="min-w-full divide-y divide-gray-200">
        <thead class="bg-gray-50">
          <tr>
            <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">التاريخ</th>
            <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">اسم الملف</th>
            <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">الحجم</th>
            <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">الحالة</th>
          </tr>
        </thead>
        <tbody class="bg-white divide-y divide-gray-200">
          <tr v-for="backup in backups" :key="backup.id">
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
              {{ new Date(backup.created_at).toLocaleString('ar-LY') }}
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900 font-mono">
              {{ backup.filename }}
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
              {{ (backup.size_bytes / 1024 / 1024).toFixed(2) }} MB
            </td>
            <td class="px-6 py-4 whitespace-nowrap">
              <span 
                class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full"
                :class="backup.status === 'completed' ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'"
              >
                {{ backup.status === 'completed' ? 'مكتمل' : 'فشل' }}
              </span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import backupApi from '@/api/backups'

const backups = ref([])
const loading = ref(true)
const creating = ref(false)

const fetchBackups = async () => {
  try {
    const response = await backupApi.list()
    backups.value = response.data.backups || []
  } catch (error) {
    console.error('Failed to list backups', error)
  } finally {
    loading.value = false
  }
}

const createBackup = async () => {
  try {
    creating.value = true
    await backupApi.create()
    await fetchBackups() // Refresh list
    alert('تم إنشاء النسخة الاحتياطية بنجاح')
  } catch (error) {
    console.error('Failed to create backup', error)
    alert('فشل إنشاء النسخة الاحتياطية')
  } finally {
    creating.value = false
  }
}

onMounted(() => {
  fetchBackups()
})
</script>
