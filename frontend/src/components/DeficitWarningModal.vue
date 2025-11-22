<script setup>
import { ref } from 'vue'

const props = defineProps({
  show: Boolean,
  message: String,
  currentBalance: Number,
  amount: Number,
  newBalance: Number,
  currency: String,
})

const emit = defineEmits(['confirm', 'cancel'])

const handleConfirm = () => {
  emit('confirm')
}

const handleCancel = () => {
  emit('cancel')
}
</script>

<template>
  <Teleport to="body">
    <Transition name="modal">
      <div v-if="show" class="fixed inset-0 z-50 overflow-y-auto">
        <!-- Overlay -->
        <div class="fixed inset-0 bg-black bg-opacity-50 transition-opacity" @click="handleCancel"></div>

        <!-- Modal -->
        <div class="flex min-h-full items-center justify-center p-4">
          <div class="relative bg-white rounded-lg shadow-xl max-w-md w-full p-6">
            <!-- Icon -->
            <div class="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-deficit-light mb-4">
              <span class="text-2xl">⚠️</span>
            </div>

            <!-- Title -->
            <h3 class="text-lg font-bold text-center text-gray-900 mb-4">
              تحذير: الرصيد غير كافٍ
            </h3>

            <!-- Message -->
            <div class="bg-gray-50 rounded-lg p-4 mb-4 space-y-2">
              <p class="text-sm text-gray-700">{{ message }}</p>
              
              <div class="border-t border-gray-200 pt-2 mt-2">
                <div class="flex justify-between text-sm">
                  <span class="text-gray-600">الرصيد الحالي:</span>
                  <span class="font-semibold">{{ currentBalance?.toFixed(2) }} {{ currency }}</span>
                </div>
                <div class="flex justify-between text-sm">
                  <span class="text-gray-600">المبلغ المطلوب:</span>
                  <span class="font-semibold">{{ amount?.toFixed(2) }} {{ currency }}</span>
                </div>
                <div class="flex justify-between text-sm border-t border-gray-200 pt-2 mt-2">
                  <span class="text-gray-600">الرصيد بعد العملية:</span>
                  <span class="font-bold text-deficit">{{ newBalance?.toFixed(2) }} {{ currency }}</span>
                </div>
              </div>
            </div>

            <!-- Warning Text -->
            <div class="bg-deficit-light border border-deficit rounded-lg p-3 mb-6">
              <p class="text-sm text-deficit-dark">
                <strong>تنبيه:</strong> سيتم تسجيل هذه العملية بالعجز (رصيد سالب). 
                يجب سداد هذا العجز عند توفر السيولة.
              </p>
            </div>

            <!-- Buttons -->
            <div class="flex gap-3">
              <button
                @click="handleCancel"
                class="flex-1 btn-secondary"
              >
                إلغاء
              </button>
              <button
                @click="handleConfirm"
                class="flex-1 btn-danger"
              >
                متابعة بالعجز
              </button>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.3s ease;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}
</style>
