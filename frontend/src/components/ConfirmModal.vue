<template>
  <Teleport to="body">
    <Transition name="modal">
      <div v-if="modelValue" class="modal-overlay" @click.self="handleCancel">
        <div class="modal-container" :class="{ 'modal-danger': danger }">
          <!-- Header -->
          <div class="modal-header">
            <div class="modal-icon" :class="{ 'icon-danger': danger, 'icon-warning': !danger }">
              <svg v-if="danger" xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
              </svg>
              <svg v-else xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.228 9c.549-1.165 2.03-2 3.772-2 2.21 0 4 1.343 4 3 0 1.4-1.278 2.575-3.006 2.907-.542.104-.994.54-.994 1.093m0 3h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
            <h3 class="modal-title">{{ title }}</h3>
          </div>
          
          <!-- Body -->
          <div class="modal-body">
            <p class="modal-message">{{ message }}</p>
            
            <!-- Two-step confirmation -->
            <div v-if="requireConfirmation && !confirmed" class="confirm-checkbox">
              <label class="confirm-label">
                <input type="checkbox" v-model="confirmed" class="confirm-input" />
                <span>{{ confirmationText || 'أؤكد أنني أريد المتابعة' }}</span>
              </label>
            </div>
          </div>
          
          <!-- Footer -->
          <div class="modal-footer">
            <button 
              type="button" 
              class="btn btn-cancel" 
              @click="handleCancel"
              :disabled="loading"
            >
              {{ cancelText || 'إلغاء' }}
            </button>
            <button 
              type="button" 
              class="btn" 
              :class="danger ? 'btn-danger' : 'btn-primary'"
              @click="handleConfirm"
              :disabled="(requireConfirmation && !confirmed) || loading"
            >
              <span v-if="loading" class="spinner"></span>
              {{ confirmText || 'تأكيد' }}
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup>
import { ref, watch } from 'vue'

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  },
  title: {
    type: String,
    default: 'تأكيد'
  },
  message: {
    type: String,
    default: 'هل أنت متأكد من هذا الإجراء؟'
  },
  confirmText: {
    type: String,
    default: 'تأكيد'
  },
  cancelText: {
    type: String,
    default: 'إلغاء'
  },
  danger: {
    type: Boolean,
    default: false
  },
  requireConfirmation: {
    type: Boolean,
    default: false
  },
  confirmationText: {
    type: String,
    default: ''
  },
  loading: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['update:modelValue', 'confirm', 'cancel'])

const confirmed = ref(false)

// Reset confirmation when modal closes
watch(() => props.modelValue, (newVal) => {
  if (!newVal) {
    confirmed.value = false
  }
})

const handleConfirm = () => {
  emit('confirm')
}

const handleCancel = () => {
  emit('update:modelValue', false)
  emit('cancel')
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
  backdrop-filter: blur(2px);
}

.modal-container {
  background: white;
  border-radius: 12px;
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
  width: 100%;
  max-width: 400px;
  margin: 1rem;
  overflow: hidden;
}

.modal-danger {
  border-top: 4px solid #ef4444;
}

.modal-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 20px 20px 12px;
}

.modal-icon {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.icon-danger {
  background-color: #fef2f2;
  color: #ef4444;
}

.icon-warning {
  background-color: #fff7ed;
  color: #f97316;
}

.modal-title {
  font-size: 1.125rem;
  font-weight: 600;
  color: #1f2937;
  margin: 0;
}

.modal-body {
  padding: 0 20px 20px;
}

.modal-message {
  color: #6b7280;
  margin: 0 0 16px;
  line-height: 1.6;
}

.confirm-checkbox {
  background-color: #f9fafb;
  border-radius: 8px;
  padding: 12px;
}

.confirm-label {
  display: flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
  font-size: 0.875rem;
  color: #374151;
}

.confirm-input {
  width: 18px;
  height: 18px;
  accent-color: #ef4444;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 20px;
  background-color: #f9fafb;
  border-top: 1px solid #e5e7eb;
}

.btn {
  padding: 10px 20px;
  border-radius: 8px;
  font-weight: 500;
  font-size: 0.875rem;
  cursor: pointer;
  border: none;
  transition: all 0.2s;
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-cancel {
  background-color: white;
  color: #374151;
  border: 1px solid #d1d5db;
}

.btn-cancel:hover:not(:disabled) {
  background-color: #f3f4f6;
}

.btn-primary {
  background-color: #3b82f6;
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background-color: #2563eb;
}

.btn-danger {
  background-color: #ef4444;
  color: white;
}

.btn-danger:hover:not(:disabled) {
  background-color: #dc2626;
}

.spinner {
  width: 16px;
  height: 16px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

/* Transitions */
.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.2s ease;
}

.modal-enter-active .modal-container,
.modal-leave-active .modal-container {
  transition: transform 0.2s ease;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}

.modal-enter-from .modal-container,
.modal-leave-to .modal-container {
  transform: scale(0.95);
}
</style>
