<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import api from '@/api/client'

const loading = ref(false)
const sending = ref(false)
const success = ref(false)
const error = ref('')
const successMessage = ref('')

// Form data
const recipientType = ref('users')
const recipients = ref([])
const selectedIds = ref([])
const subject = ref('')
const message = ref('')

// Fetch recipients based on type
const fetchRecipients = async () => {
  loading.value = true
  error.value = ''
  recipients.value = []
  selectedIds.value = []
  
  try {
    let response
    switch (recipientType.value) {
      case 'users':
        response = await api.get('/users')
        recipients.value = response.data.map(u => ({
          id: u.id,
          name: u.full_name_ar || u.full_name,
          email: u.email,
          detail: u.role === 'admin' ? 'مدير' : u.role === 'finance' ? 'محاسب' : u.role === 'hr' ? 'موارد بشرية' : 'مراجع'
        }))
        break
      case 'employees':
        response = await api.get('/employees')
        recipients.value = response.data.filter(e => e.email).map(e => ({
          id: e.id,
          name: e.full_name_ar || e.full_name,
          email: e.email,
          detail: e.position_ar || e.position || 'موظف'
        }))
        break
      case 'members':
        response = await api.get('/members')
        recipients.value = response.data.filter(m => m.email).map(m => ({
          id: m.id,
          name: m.full_name_ar || m.full_name,
          email: m.email,
          detail: m.member_type || 'عضو'
        }))
        break
    }
  } catch (err) {
    error.value = 'فشل في تحميل القائمة'
    console.error(err)
  } finally {
    loading.value = false
  }
}

// Watch for type changes
watch(recipientType, () => {
  fetchRecipients()
})

onMounted(() => {
  fetchRecipients()
})

// Select all toggle
const allSelected = computed({
  get: () => recipients.value.length > 0 && selectedIds.value.length === recipients.value.length,
  set: (val) => {
    selectedIds.value = val ? recipients.value.map(r => r.id) : []
  }
})

// Toggle single selection
const toggleSelection = (id) => {
  const idx = selectedIds.value.indexOf(id)
  if (idx > -1) {
    selectedIds.value.splice(idx, 1)
  } else {
    selectedIds.value.push(id)
  }
}

// Send emails
const handleSend = async () => {
  if (selectedIds.value.length === 0) {
    error.value = 'يرجى اختيار مستلم واحد على الأقل'
    return
  }
  if (!subject.value.trim()) {
    error.value = 'يرجى إدخال عنوان الرسالة'
    return
  }
  if (!message.value.trim()) {
    error.value = 'يرجى إدخال محتوى الرسالة'
    return
  }
  
  sending.value = true
  error.value = ''
  success.value = false
  
  try {
    const response = await api.post('/email/bulk', {
      recipient_type: recipientType.value,
      recipient_ids: selectedIds.value,
      subject: subject.value,
      message: message.value
    })
    
    success.value = true
    successMessage.value = `تم إرسال ${response.data.sent} رسالة بنجاح${response.data.failed > 0 ? ` (فشل إرسال ${response.data.failed})` : ''}`
    
    // Reset form
    selectedIds.value = []
    subject.value = ''
    message.value = ''
  } catch (err) {
    error.value = err.response?.data?.error || 'فشل في إرسال الرسائل'
  } finally {
    sending.value = false
  }
}

// Recipient type labels
const typeLabels = {
  users: { label: 'المستخدمين', icon: '👤' },
  employees: { label: 'الموظفين', icon: '👷' },
  members: { label: 'الأعضاء المنتسبين', icon: '🤝' }
}
</script>

<template>
  <div class="messaging-page">
    <div class="page-header">
      <h1>📧 مراسلات النظام</h1>
      <p class="subtitle">إرسال رسائل بريدية للمستخدمين أو الموظفين أو الأعضاء</p>
    </div>

    <div class="messaging-container">
      <!-- Left Panel: Recipients -->
      <div class="recipients-panel">
        <div class="panel-header">
          <h2>المستلمين</h2>
          
          <!-- Type Selector -->
          <div class="type-selector">
            <button 
              v-for="(info, key) in typeLabels" 
              :key="key"
              :class="['type-btn', { active: recipientType === key }]"
              @click="recipientType = key"
            >
              <span class="type-icon">{{ info.icon }}</span>
              <span>{{ info.label }}</span>
            </button>
          </div>
        </div>
        
        <!-- Select All -->
        <div class="select-all-row" v-if="recipients.length > 0">
          <label class="checkbox-label">
            <input type="checkbox" v-model="allSelected" />
            <span class="checkmark"></span>
            <span>تحديد الكل ({{ selectedIds.length }}/{{ recipients.length }})</span>
          </label>
        </div>
        
        <!-- Recipients List -->
        <div class="recipients-list" v-if="!loading">
          <div 
            v-for="recipient in recipients" 
            :key="recipient.id"
            :class="['recipient-card', { selected: selectedIds.includes(recipient.id) }]"
            @click="toggleSelection(recipient.id)"
          >
            <div class="recipient-checkbox">
              <input 
                type="checkbox" 
                :checked="selectedIds.includes(recipient.id)"
                @click.stop="toggleSelection(recipient.id)"
              />
            </div>
            <div class="recipient-info">
              <div class="recipient-name">{{ recipient.name }}</div>
              <div class="recipient-email">{{ recipient.email }}</div>
            </div>
            <div class="recipient-detail">{{ recipient.detail }}</div>
          </div>
          
          <div v-if="recipients.length === 0" class="empty-state">
            <span class="empty-icon">📭</span>
            <p>لا يوجد مستلمين في هذه القائمة</p>
          </div>
        </div>
        
        <div v-else class="loading-state">
          <div class="spinner"></div>
          <p>جاري التحميل...</p>
        </div>
      </div>
      
      <!-- Right Panel: Message Composer -->
      <div class="composer-panel">
        <div class="panel-header">
          <h2>✍️ صياغة الرسالة</h2>
        </div>
        
        <!-- Alerts -->
        <div v-if="success" class="alert success">
          ✅ {{ successMessage }}
        </div>
        <div v-if="error" class="alert error">
          ❌ {{ error }}
        </div>
        
        <!-- Selected Count -->
        <div class="selected-summary" v-if="selectedIds.length > 0">
          <span class="badge">{{ selectedIds.length }}</span>
          <span>مستلم تم تحديدهم</span>
        </div>
        
        <!-- Subject -->
        <div class="form-group">
          <label>عنوان الرسالة <span class="required">*</span></label>
          <input 
            v-model="subject" 
            type="text" 
            placeholder="أدخل عنوان الرسالة..."
            class="input-field"
          />
        </div>
        
        <!-- Message Body -->
        <div class="form-group">
          <label>محتوى الرسالة <span class="required">*</span></label>
          <textarea 
            v-model="message" 
            rows="10" 
            placeholder="اكتب رسالتك هنا..."
            class="input-field message-textarea"
          ></textarea>
        </div>
        
        <!-- Preview -->
        <div class="preview-section" v-if="message.trim()">
          <h3>معاينة الرسالة:</h3>
          <div class="preview-box">
            <div class="preview-header">
              <span>📧</span>
              <strong>{{ subject || 'بدون عنوان' }}</strong>
            </div>
            <div class="preview-content">
              {{ message }}
            </div>
          </div>
        </div>
        
        <!-- Send Button -->
        <div class="actions">
          <button 
            @click="handleSend" 
            :disabled="sending || selectedIds.length === 0"
            class="send-btn"
          >
            <span v-if="sending">⏳ جاري الإرسال...</span>
            <span v-else>📤 إرسال الرسالة</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.messaging-page {
  padding: 20px;
  direction: rtl;
}

.page-header {
  margin-bottom: 30px;
}

.page-header h1 {
  font-size: 28px;
  font-weight: 700;
  margin: 0;
  color: #1f2937;
}

.page-header .subtitle {
  color: #6b7280;
  margin-top: 8px;
  font-size: 14px;
}

.messaging-container {
  display: grid;
  grid-template-columns: 1fr 1.2fr;
  gap: 25px;
}

/* Panels */
.recipients-panel,
.composer-panel {
  background: white;
  border-radius: 16px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.06);
  overflow: hidden;
}

.panel-header {
  background: linear-gradient(135deg, #1e40af 0%, #3b82f6 100%);
  color: white;
  padding: 20px;
}

.panel-header h2 {
  margin: 0 0 15px 0;
  font-size: 18px;
}

/* Type Selector */
.type-selector {
  display: flex;
  gap: 8px;
}

.type-btn {
  background: rgba(255, 255, 255, 0.15);
  border: 1px solid rgba(255, 255, 255, 0.3);
  color: white;
  padding: 8px 16px;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
}

.type-btn:hover {
  background: rgba(255, 255, 255, 0.25);
}

.type-btn.active {
  background: white;
  color: #1e40af;
  font-weight: 600;
}

.type-icon {
  font-size: 16px;
}

/* Select All */
.select-all-row {
  padding: 15px 20px;
  background: #f8fafc;
  border-bottom: 1px solid #e5e7eb;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
  font-size: 14px;
  color: #374151;
}

/* Recipients List */
.recipients-list {
  max-height: 500px;
  overflow-y: auto;
  padding: 10px;
}

.recipient-card {
  display: flex;
  align-items: center;
  gap: 15px;
  padding: 15px;
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.2s;
  border: 2px solid transparent;
  margin-bottom: 8px;
  background: #f9fafb;
}

.recipient-card:hover {
  background: #f0f9ff;
  border-color: #bae6fd;
}

.recipient-card.selected {
  background: #dbeafe;
  border-color: #3b82f6;
}

.recipient-checkbox input {
  width: 18px;
  height: 18px;
  cursor: pointer;
}

.recipient-info {
  flex: 1;
}

.recipient-name {
  font-weight: 600;
  color: #1f2937;
  font-size: 14px;
}

.recipient-email {
  font-size: 12px;
  color: #6b7280;
  margin-top: 2px;
}

.recipient-detail {
  font-size: 11px;
  color: white;
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  padding: 4px 10px;
  border-radius: 20px;
}

/* Empty/Loading States */
.empty-state,
.loading-state {
  text-align: center;
  padding: 40px;
  color: #6b7280;
}

.empty-icon {
  font-size: 48px;
  display: block;
  margin-bottom: 15px;
}

.spinner {
  width: 40px;
  height: 40px;
  border: 4px solid #e5e7eb;
  border-top-color: #3b82f6;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin: 0 auto 15px;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* Composer Panel */
.composer-panel .panel-header {
  background: linear-gradient(135deg, #059669 0%, #10b981 100%);
}

.alert {
  margin: 15px 20px;
  padding: 12px 16px;
  border-radius: 8px;
  font-size: 14px;
}

.alert.success {
  background: #d1fae5;
  color: #065f46;
  border: 1px solid #6ee7b7;
}

.alert.error {
  background: #fee2e2;
  color: #991b1b;
  border: 1px solid #fca5a5;
}

.selected-summary {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 20px;
  background: #eff6ff;
  color: #1e40af;
  font-size: 14px;
}

.badge {
  background: #3b82f6;
  color: white;
  padding: 4px 12px;
  border-radius: 20px;
  font-weight: 700;
}

.form-group {
  padding: 0 20px;
  margin-bottom: 20px;
}

.form-group:first-of-type {
  margin-top: 20px;
}

.form-group label {
  display: block;
  font-weight: 600;
  color: #374151;
  margin-bottom: 8px;
  font-size: 14px;
}

.required {
  color: #dc2626;
}

.input-field {
  width: 100%;
  padding: 12px 15px;
  border: 2px solid #e5e7eb;
  border-radius: 10px;
  font-size: 14px;
  transition: all 0.2s;
  box-sizing: border-box;
}

.input-field:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.message-textarea {
  resize: vertical;
  min-height: 150px;
  font-family: inherit;
  line-height: 1.6;
}

/* Preview */
.preview-section {
  padding: 0 20px;
  margin-bottom: 20px;
}

.preview-section h3 {
  font-size: 13px;
  color: #6b7280;
  margin-bottom: 10px;
}

.preview-box {
  background: #f8fafc;
  border: 1px solid #e5e7eb;
  border-radius: 10px;
  overflow: hidden;
}

.preview-header {
  background: linear-gradient(135deg, #1e40af 0%, #3b82f6 100%);
  color: white;
  padding: 12px 15px;
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 14px;
}

.preview-content {
  padding: 15px;
  font-size: 13px;
  color: #374151;
  line-height: 1.7;
  white-space: pre-wrap;
}

/* Actions */
.actions {
  padding: 20px;
  border-top: 1px solid #e5e7eb;
}

.send-btn {
  width: 100%;
  padding: 15px;
  background: linear-gradient(135deg, #059669 0%, #10b981 100%);
  color: white;
  border: none;
  border-radius: 10px;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
}

.send-btn:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(16, 185, 129, 0.3);
}

.send-btn:disabled {
  background: #9ca3af;
  cursor: not-allowed;
}

/* Responsive */
@media (max-width: 900px) {
  .messaging-container {
    grid-template-columns: 1fr;
  }
}
</style>
