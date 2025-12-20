<template>
  <div class="max-w-4xl mx-auto py-10 px-4">
    <div v-if="loading" class="text-center">جاري التحميل...</div>
    <div v-else class="bg-white p-8 rounded-lg shadow-lg prose max-w-none rtl" v-html="content"></div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import legalApi from '@/api/legal'
import { marked } from 'marked' // Assuming marked is installed or needs to be

const content = ref('')
const loading = ref(true)

onMounted(async () => {
  try {
    const response = await legalApi.getPrivacyPolicy()
    // Convert markdown to HTML if response contains markdown
    // For now assuming content comes as string
    content.value = response.data.content // Simple text display or use markdown parser
    
    // Quick markdown parser fallback if 'marked' isn't available
    if (response.data.content) {
       content.value = response.data.content.replace(/\n/g, '<br>').replace(/#{1,6}\s+(.*)/g, '<h3 class="text-xl font-bold mt-4 mb-2">$1</h3>')
    }
  } catch (error) {
    console.error('Failed to load privacy policy', error)
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.rtl {
  direction: rtl;
}
</style>
