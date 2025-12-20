<template>
  <div class="p-6">
    <div class="flex justify-between items-center mb-6">
      <h1 class="text-2xl font-bold text-gray-800">التنبيهات</h1>
      <button 
        @click="markAllRead"
        class="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700"
        :disabled="loading"
      >
        تعليم الكل كمقروء
      </button>
    </div>

    <div v-if="loading" class="text-center py-10">
      <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mx-auto"></div>
    </div>

    <div v-else-if="notifications.length === 0" class="text-center py-10 bg-white rounded shadow text-gray-500">
      لا توجد تنبيهات
    </div>

    <div v-else class="space-y-4">
      <div 
        v-for="notification in notifications" 
        :key="notification.id"
        class="bg-white p-4 rounded shadow border-r-4 transition-all hover:bg-gray-50"
        :class="notification.is_read ? 'border-gray-200 opacity-75' : 'border-blue-500'"
        @click="markAsRead(notification)"
      >
        <div class="flex justify-between items-start">
          <div>
            <h3 class="font-bold text-gray-800" :class="{ 'text-blue-600': !notification.is_read }">
              {{ notification.title }}
            </h3>
            <p class="text-gray-600 mt-1">{{ notification.message }}</p>
            <span class="text-xs text-gray-400 mt-2 block">
              {{ new Date(notification.created_at).toLocaleString('ar-LY') }}
            </span>
          </div>
          <div v-if="!notification.is_read" class="h-3 w-3 bg-blue-500 rounded-full"></div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import interactions from '@/api/notifications'

const notifications = ref([])
const loading = ref(true)

const fetchNotifications = async () => {
  try {
    const response = await interactions.list()
    notifications.value = response.data.notifications
  } catch (error) {
    console.error('Failed to load notifications', error)
  } finally {
    loading.value = false
  }
}

const markAsRead = async (notification) => {
  if (notification.is_read) return

  try {
    await interactions.markAsRead(notification.id)
    notification.is_read = true
    // Optionally trigger a global event to update badge count
  } catch (error) {
    console.error('Failed to mark as read', error)
  }
}

const markAllRead = async () => {
  try {
    loading.value = true
    await interactions.markAllAsRead()
    notifications.value.forEach(n => n.is_read = true)
  } catch (error) {
    console.error('Failed to mark all as read', error)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchNotifications()
})
</script>
