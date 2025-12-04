<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRoute } from 'vue-router'
import api from '@/api/client'
import html2canvas from 'html2canvas'
import jsPDF from 'jspdf'

const route = useRoute()
const loading = ref(true)
const error = ref('')
const receiptData = ref(null)

// Organization info
const orgInfo = {
  nameAr: 'المنظمة الليبية للتعاون',
  nameEn: 'Libyan Cooperation Organization',
  phone: '+218 XXX XXX XXX',
  email: 'info@lco.org.ly',
  address: 'طرابلس، ليبيا'
}

onMounted(async () => {
  try {
    const transactionId = route.params.id
    const response = await api.get(`/receipts/${transactionId}`)
    receiptData.value = response.data
    loading.value = false
  } catch (err) {
    console.error('Failed to load receipt:', err)
    error.value = 'فشل في تحميل الإيصال'
    loading.value = false
  }
})

// Get transaction type label
const getTransactionTypeLabel = computed(() => {
  if (!receiptData.value) return ''
  
  const typeLabels = {
    'income_grant': 'إيصال قبض - منحة',
    'income_donation': 'إيصال قبض - تبرع',
    'disbursement_custody': 'إيصال صرف - عهدة',
    'disbursement_salary': 'إيصال صرف - راتب',
    'disbursement_operational': 'إيصال صرف - تشغيلية',
    'disbursement_donation': 'إيصال صرف - تبرعات',
    'disbursement_bonus': 'إيصال صرف - مكافأة',
    'procurement': 'إيصال شراء',
    'transfer': 'إيصال تحويل'
  }
  
  return typeLabels[receiptData.value.transaction.type] || 'إيصال'
})

// Check if transaction is income or expense
const isIncome = computed(() => {
  if (!receiptData.value) return false
  const type = receiptData.value.transaction.type
  return type === 'income_grant' || type === 'income_donation'
})

// Get detailed transaction description parts
const transactionDescriptionParts = computed(() => {
  if (!receiptData.value) return []
  
  const t = receiptData.value.transaction
  const c = receiptData.value.container
  const p = receiptData.value.project
  
  const containerType = c.type === 'cash' ? 'نقدي' : 'مصرفي'
  const containerName = c.name_ar || c.name
  const projectName = p.name_ar || p.name
  const beneficiary = t.payee_payer_ar || t.payee_payer
  
  // Get transaction type in Arabic
  let typeText = ''
  if (t.type === 'income_grant') typeText = 'منحة'
  else if (t.type === 'income_donation') typeText = 'تبرع'
  else if (t.type === 'disbursement_custody') typeText = 'عهدة'
  else if (t.type === 'disbursement_salary') typeText = 'راتب'
  else if (t.type === 'disbursement_operational') typeText = 'صرف تشغيلي'
  
  const parts = []
  
  if (isIncome.value) {
    parts.push(`قبض مبلغ (${typeText})`)
    parts.push(`إلى حساب ${containerType} (${containerName})`)
    parts.push(`لصالح مشروع ${projectName}`)
    if (beneficiary) parts.push(`من ${beneficiary}`)
  } else if (t.type === 'transfer') {
    parts.push('تحويل مبلغ')
    
    // Check if we have related container info
    const relatedC = receiptData.value.related_container
    
    if (t.original_transaction_id) {
      // This is credit (IN), so it's FROM related TO current
      const relatedName = relatedC ? (relatedC.name_ar || relatedC.name) : '...'
      parts.push(`إلى حساب ${containerType} (${containerName})`)
      parts.push(`من حساب ${relatedName}`)
    } else {
      // This is debit (OUT), so it's FROM current TO related
      const relatedName = relatedC ? (relatedC.name_ar || relatedC.name) : '...'
      parts.push(`من حساب ${containerType} (${containerName})`)
      parts.push(`إلى حساب ${relatedName}`)
    }
  } else {
    parts.push(`دفع مبلغ (${typeText})`)
    parts.push(`من حساب ${containerType} (${containerName})`)
    parts.push(`من مشروع ${projectName}`)
    if (beneficiary) parts.push(`إلى ${beneficiary}`)
  }
  
  return parts
})

// Convert number to Arabic words
const numberToArabicWords = (num) => {
  if (num === 0) return 'صفر'
  
  const ones = ['', 'واحد', 'اثنان', 'ثلاثة', 'أربعة', 'خمسة', 'ستة', 'سبعة', 'ثمانية', 'تسعة']
  const tens = ['', 'عشرة', 'عشرون', 'ثلاثون', 'أربعون', 'خمسون', 'ستون', 'سبعون', 'ثمانون', 'تسعون']
  const hundreds = ['', 'مائة', 'مئتان', 'ثلاثمائة', 'أربعمائة', 'خمسمائة', 'ستمائة', 'سبعمائة', 'ثمانمائة', 'تسعمائة']
  
  const int = Math.floor(num)
  const decimal = Math.round((num - int) * 100)
  
  let result = ''
  
  // Thousands
  if (int >= 1000) {
    const th = Math.floor(int / 1000)
    if (th === 1) result += 'ألف '
    else if (th === 2) result += 'ألفان '
    else if (th <= 10) result += ones[th] + ' آلاف '
    else result += th + ' ألف '
  }
  
  // Hundreds
  const h = Math.floor((int % 1000) / 100)
  if (h > 0) result += hundreds[h] + ' '
  
  // Tens and ones
  const remainder = int % 100
  if (remainder >= 20) {
    result += tens[Math.floor(remainder / 10)] + ' '
    if (remainder % 10 > 0) result += ones[remainder % 10] + ' '
  } else if (remainder >= 10) {
    result += tens[1] + ' '
    if (remainder > 10) result += ones[remainder - 10] + ' '
  } else if (remainder > 0) {
    result += ones[remainder] + ' '
  }
  
  return { integer: result.trim() || 'صفر', decimal: decimal }
}

// Get currency info (symbol and subunit name)
const getCurrencyInfo = (currency) => {
  const currencyData = {
    'LYD': { symbol: 'د.ل', main: 'دينار ليبي', sub: 'درهم' },
    'USD': { symbol: '$', main: 'دولار أمريكي', sub: 'سنت' },
    'EUR': { symbol: '€', main: 'يورو', sub: 'سنت' },
    'GBP': { symbol: '£', main: 'جنيه إسترليني', sub: 'بنس' }
  }
  return currencyData[currency] || { symbol: currency, main: currency, sub: 'فلس' }
}

// Format amount in words
const amountInWords = computed(() => {
  if (!receiptData.value) return ''
  const amount = receiptData.value.transaction.amount
  const currency = receiptData.value.transaction.currency
  
  const parts = numberToArabicWords(amount)
  const currencyInfo = getCurrencyInfo(currency)
  
  let result = parts.integer + ' ' + currencyInfo.main
  
  if (parts.decimal > 0) {
    // For decimals, use the subunit name
    const decimalWord = numberToArabicWords(parts.decimal).integer
    result += ' و' + decimalWord + ' ' + currencyInfo.sub
  }
  
  return result
})

// Print receipt
const printReceipt = () => {
  window.print()
}

// Export as JPG
const exportAsJPG = async () => {
  try {
    const element = document.getElementById('receipt-content')
    const canvas = await html2canvas(element, {
      scale: 2,
      backgroundColor: '#ffffff',
      logging: false
    })
    
    const link = document.createElement('a')
    link.download = `receipt-${receiptData.value.transaction.transaction_number}.jpg`
    link.href = canvas.toDataURL('image/jpeg', 0.95)
    link.click()
  } catch (err) {
    console.error('Failed to export as JPG:', err)
    alert('فشل في تصدير الإيصال كصورة')
  }
}

// Export as PDF
const exportAsPDF = async () => {
  try {
    const element = document.getElementById('receipt-content')
    const canvas = await html2canvas(element, {
      scale: 2,
      backgroundColor: '#ffffff',
      logging: false,
      windowWidth: element.scrollWidth,
      windowHeight: element.scrollHeight
    })
    
    const pdf = new jsPDF({
      orientation: 'landscape',
      unit: 'mm',
      format: 'a5'
    })
    
    // A5 landscape dimensions: 210mm x 148mm
    const pdfWidth = 210
    const pdfHeight = 148
    
    // Add margins (e.g., 10mm on each side)
    const marginX = 10
    const contentWidth = pdfWidth - (2 * marginX)
    
    const imgWidth = contentWidth
    const imgHeight = (canvas.height * contentWidth) / canvas.width
    
    // Center vertically
    const yOffset = imgHeight < pdfHeight ? (pdfHeight - imgHeight) / 2 : 10 // Default top margin if tall
    
    pdf.addImage(canvas.toDataURL('image/jpeg', 0.95), 'JPEG', marginX, yOffset, imgWidth, imgHeight)
    pdf.save(`receipt-${receiptData.value.transaction.transaction_number}.pdf`)
  } catch (err) {
    console.error('Failed to export as PDF:', err)
    alert('فشل في تصدير الإيصال كملف PDF')
  }
}

// Format date with standard numerals
const formatDate = (date) => {
  if (!date) return ''
  const d = new Date(date)
  const day = d.getDate()
  const month = d.toLocaleDateString('ar-EG', { month: 'long' })
  const year = d.getFullYear()
  return `${day} ${month} ${year}`
}

// Format time with standard numerals
const formatTime = (date) => {
  if (!date) return ''
  const d = new Date(date)
  const hours = String(d.getHours()).padStart(2, '0')
  const minutes = String(d.getMinutes()).padStart(2, '0')
  return `${hours}:${minutes}`
}
</script>

<template>
  <div class="receipt-page">
    <!-- Loading state -->
    <div v-if="loading" class="loading-container">
      <div class="loading-spinner"></div>
      <p>جاري تحميل الإيصال...</p>
    </div>

    <!-- Error state -->
    <div v-else-if="error" class="error-container">
      <div class="error-message">
        <h2>❌ خطأ</h2>
        <p>{{ error }}</p>
        <button @click="$router.go(-1)" class="btn-secondary">العودة</button>
      </div>
    </div>

    <!-- Receipt content -->
    <div v-else class="receipt-wrapper">
      <!-- Action buttons (hidden in print) -->
      <div class="action-buttons no-print">
        <button @click="printReceipt" class="btn-primary">
          🖨️ طباعة
        </button>
        <button @click="exportAsPDF" class="btn-success">
          📄 حفظ PDF
        </button>
        <button @click="exportAsJPG" class="btn-info">
          🖼️ حفظ JPG
        </button>
        <button @click="$router.go(-1)" class="btn-secondary">
          🔙 إغلاق
        </button>
      </div>

      <!-- Receipt content (A5 landscape) -->
      <div id="receipt-content" class="receipt-content">
        <!-- Header -->
        <div class="receipt-header">
          <div class="org-logo">
            <div class="logo-placeholder">LOGO</div>
          </div>
          <div class="org-info">
            <h1>{{ orgInfo.nameAr }}</h1>
            <p class="org-contact">{{ orgInfo.phone }} | {{ orgInfo.email }}</p>
            <p class="org-address">{{ orgInfo.address }}</p>
          </div>
          <div class="receipt-type" :class="{ 'income-type': isIncome, 'expense-type': !isIncome }">
            <h2>{{ getTransactionTypeLabel }}</h2>
          </div>
        </div>

        <div class="divider"></div>

        <!-- Transaction header info -->
        <div class="transaction-header">
          <div class="header-row">
            <div class="header-item">
              <span class="label">رقم العملية:</span>
              <span class="value-large">{{ receiptData.transaction.transaction_number }}</span>
            </div>
            <div class="header-item">
              <span class="label">التاريخ:</span>
              <span class="value-large">{{ formatDate(receiptData.transaction.transaction_date) }}</span>
            </div>
          </div>
        </div>

        <!-- Detailed transaction description -->
        <div class="transaction-details-box">
          <div class="details-line" v-for="(line, index) in transactionDescriptionParts" :key="index">
            {{ line }}
          </div>
        </div>

        <!-- Amount section -->
        <div class="amount-section">
          <div class="amount-box">
            <div class="amount-value">
              {{ receiptData.transaction.amount.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 }) }}
              {{ getCurrencyInfo(receiptData.transaction.currency).symbol }}
            </div>
            <div class="amount-words">{{ amountInWords }}</div>
          </div>
        </div>

        <!-- Statement/Description -->
        <div class="statement-section" v-if="receiptData.transaction.description_ar || receiptData.transaction.description">
          <div class="statement-label">البيان:</div>
          <div class="statement-text">{{ receiptData.transaction.description_ar || receiptData.transaction.description }}</div>
        </div>

        <!-- Additional details grid -->
        <div class="details-grid">
          <div class="detail-row" v-if="receiptData.transaction.payee_payer_ar || receiptData.transaction.payee_payer">
            <span class="detail-label">{{ isIncome ? 'المانح/الدافع:' : 'المستفيد:' }}</span>
            <span class="detail-value">{{ receiptData.transaction.payee_payer_ar || receiptData.transaction.payee_payer }}</span>
          </div>

          <div class="detail-row" v-if="receiptData.transaction.reference_number">
            <span class="detail-label">رقم المرجع/الشيك:</span>
            <span class="detail-value">{{ receiptData.transaction.reference_number }}</span>
          </div>
        </div>

        <div class="divider"></div>

        <!-- Signatures -->
        <div class="signatures-section">
          <div class="signature-box">
            <div class="signature-label">المحاسب</div>
            <div class="signature-line"></div>
            <div class="signature-name">{{ receiptData.created_by_user.full_name_ar || receiptData.created_by_user.full_name }}</div>
          </div>
          <div class="signature-box">
            <div class="signature-label">المدير المالي</div>
            <div class="signature-line"></div>
          </div>
          <div class="signature-box">
            <div class="signature-label">المدير العام</div>
            <div class="signature-line"></div>
          </div>
        </div>

        <!-- Footer -->
        <div class="receipt-footer">
          <p>تم إنشاء هذا الإيصال في: {{ formatDate(receiptData.transaction.created_at) }} - {{ formatTime(receiptData.transaction.created_at) }}</p>
        </div>
      </div>
    </div>
  </div>
</template>
<style scoped>
/* General page styles */
.receipt-page {
  min-height: 100vh;
  background-color: #f3f4f6;
  padding: 20px;
  direction: rtl;
  font-family: 'Cairo', 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
  display: flex;
  justify-content: center;
  align-items: center;
}

.loading-container,
.error-container {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 60vh;
  flex-direction: column;
}

.loading-spinner {
  width: 50px;
  height: 50px;
  border: 4px solid #e0e0e0;
  border-top-color: #3b82f6;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.error-message {
  text-align: center;
  padding: 40px;
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.1);
}

/* Action buttons */
.action-buttons {
  display: flex;
  gap: 10px;
  justify-content: center;
  margin-bottom: 20px;
  flex-wrap: wrap;
}

.action-buttons button {
  padding: 10px 20px;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 600;
  transition: all 0.3s;
}

.btn-primary {
  background: #3b82f6;
  color: white;
}

.btn-primary:hover {
  background: #2563eb;
}

.btn-success {
  background: #10b981;
  color: white;
}

.btn-success:hover {
  background: #059669;
}

.btn-info {
  background: #8b5cf6;
  color: white;
}

.btn-info:hover {
  background: #7c3aed;
}

.btn-secondary {
  background: #6b7280;
  color: white;
}

.btn-secondary:hover {
  background: #4b5563;
}

/* Receipt wrapper */
.receipt-wrapper {
  max-width: 210mm;
  margin: 0 auto;
}

/* Receipt content - A5 landscape */
.receipt-content {
  width: 210mm;
  min-height: 148mm;
  background: white;
  padding: 15mm;
  margin: 0 auto;
  box-shadow: 0 4px 16px rgba(0,0,0,0.1);
  border-radius: 8px;
}

/* Header */
.receipt-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
}

.org-logo .logo-placeholder {
  width: 70px;
  height: 70px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-weight: bold;
  font-size: 11px;
}

.org-info {
  flex: 1;
  text-align: center;
  margin: 0 20px;
}

.org-info h1 {
  font-size: 22px;
  margin: 0 0 6px 0;
  color: #1f2937;
  font-weight: 700;
}

.org-contact,
.org-address {
  font-size: 11px;
  color: #6b7280;
  margin: 3px 0;
}

.receipt-type {
  min-width: 140px;
  text-align: center;
  padding: 10px 16px;
  border-radius: 8px;
}

.receipt-type.income-type {
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
}

.receipt-type.expense-type {
  background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%);
}

.receipt-type h2 {
  font-size: 15px;
  margin: 0;
  color: white;
  font-weight: 700;
}

.divider {
  height: 2px;
  background: linear-gradient(90deg, transparent, #d1d5db, transparent);
  margin: 12px 0;
}

/* Transaction header */
.transaction-header {
  margin: 15px 0;
}

.header-row {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 20px;
}

.header-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.header-item .label {
  font-size: 12px;
  color: #6b7280;
  font-weight: 600;
}

.header-item .value-large {
  font-size: 18px;
  color: #1f2937;
  font-weight: 700;
}

/* Transaction details box */
.transaction-details-box {
  background: linear-gradient(135deg, #eff6ff 0%, #dbeafe 100%);
  border: 2px solid #3b82f6;
  border-radius: 8px;
  padding: 12px 18px;
  margin: 15px 0;
  line-height: 1.9;
}

.details-line {
  font-size: 13px;
  color: #1e40af;
  font-weight: 600;
  margin: 3px 0;
}

/* Amount section */
.amount-section {
  margin: 20px 0;
}

.amount-box {
  background: linear-gradient(135deg, #fef3c7 0%, #fde68a 100%);
  border: 3px solid #f59e0b;
  border-radius: 12px;
  padding: 20px;
  text-align: center;
}

.amount-value {
  font-size: 36px;
  font-weight: 700;
  color: #92400e;
  margin-bottom: 8px;
}

.amount-words {
  font-size: 16px;
  color: #78350f;
  font-weight: 600;
}

/* Statement section */
.statement-section {
  background: #f9fafb;
  border-right: 4px solid #6b7280;
  padding: 15px;
  margin: 15px 0;
  border-radius: 6px;
}

.statement-label {
  font-size: 13px;
  color: #4b5563;
  font-weight: 700;
  margin-bottom: 8px;
}

.statement-text {
  font-size: 15px;
  color: #1f2937;
  line-height: 1.6;
}

/* Details grid */
.details-grid {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin: 15px 0;
}

.detail-row {
  display: flex;
  padding: 10px 15px;
  background: #f3f4f6;
  border-radius: 6px;
  border-right: 3px solid #9ca3af;
}

.detail-label {
  font-size: 13px;
  color: #4b5563;
  font-weight: 600;
  min-width: 130px;
}

.detail-value {
  font-size: 14px;
  color: #1f2937;
  font-weight: 500;
  flex: 1;
}

/* Signatures */
.signatures-section {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 25px;
  margin: 25px 0 15px 0;
}

.signature-box {
  text-align: center;
}

.signature-label {
  font-size: 12px;
  color: #6b7280;
  margin-bottom: 25px;
  font-weight: 600;
}

.signature-line {
  border-top: 2px solid #000;
  margin: 0 10px;
  min-height: 35px;
}

.signature-name {
  font-size: 11px;
  color: #4b5563;
  margin-top: 6px;
  font-weight: 500;
}

/* Footer */
.receipt-footer {
  text-align: center;
  padding-top: 12px;
  border-top: 1px solid #e5e7eb;
}

.receipt-footer p {
  font-size: 10px;
  color: #9ca3af;
  margin: 0;
}

/* Print styles */
@media print {
  .no-print {
    display: none !important;
  }

  .receipt-page {
    padding: 0;
    background: white;
    display: flex;
    justify-content: center;
    align-items: center;
    height: 100vh; /* Use full height to center vertically */
    margin: 0;
  }

  .receipt-content {
    box-shadow: none;
    border: none; /* Optional: remove border for print if desired, or keep it */
    padding: 0;
    margin: 0;
    width: 90%; /* Prevent sticking to edges */
    max-width: 190mm; /* A5 width minus margins */
  }

  @page {
    size: A5 landscape;
    margin: 0; /* We handle margins via CSS */
  }
}

/* Responsive */
@media screen and (max-width: 900px) {
  .receipt-content {
    width: 100%;
    padding: 12mm;
  }

  .receipt-header {
    flex-direction: column;
    gap: 12px;
  }

  .org-info h1 {
    font-size: 18px;
  }

  .header-row {
    grid-template-columns: 1fr;
  }

  .signatures-section {
    grid-template-columns: 1fr;
    gap: 18px;
  }
}
</style>
