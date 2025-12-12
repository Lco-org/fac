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

  // Get transaction type in Arabic text only
  const getTransactionTypeText = computed(() => {
    if (!receiptData.value) return ''
    const t = receiptData.value.transaction
    if (t.type === 'income_grant') return 'منحة'
    else if (t.type === 'income_donation') return 'تبرع'
    else if (t.type === 'disbursement_custody') return 'عهدة'
    else if (t.type === 'disbursement_salary') return 'راتب'
    else if (t.type === 'disbursement_operational') return 'تشغيلية'
    else if (t.type === 'disbursement_donation') return 'تبرعات'
    else if (t.type === 'disbursement_bonus') return 'مكافأة'
    else if (t.type === 'procurement') return 'شراء'
    else if (t.type === 'transfer') return 'تحويل'
    return 'عام'
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
    
    let beneficiary = t.payee_payer_ar || t.payee_payer
    if (receiptData.value.employee) {
      beneficiary = receiptData.value.employee.full_name_ar || receiptData.value.employee.full_name
    } else if (receiptData.value.member) {
      beneficiary = receiptData.value.member.full_name_ar || receiptData.value.member.full_name
    }
    
    // Use the computed property
    let typeText = getTransactionTypeText.value
    
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
  if (num === 0) return { integer: 'صفر', decimal: 0 }
  
  const ones = ['', 'واحد', 'اثنان', 'ثلاثة', 'أربعة', 'خمسة', 'ستة', 'سبعة', 'ثمانية', 'تسعة']
  const tens = ['', 'عشرة', 'عشرون', 'ثلاثون', 'أربعون', 'خمسون', 'ستون', 'سبعون', 'ثمانون', 'تسعون']
  const hundreds = ['', 'مائة', 'مئتان', 'ثلاثمائة', 'أربعمائة', 'خمسمائة', 'ستمائة', 'سبعمائة', 'ثمانمائة', 'تسعمائة']

  const convertThreeDigits = (n) => {
    let s = ''
    // Hundreds
    let h = Math.floor(n / 100)
    let rem = n % 100
    if (h > 0) {
      s += hundreds[h]
      if (rem > 0) s += ' و'
    }
    
    // Tens and Ones
    if (rem > 0) {
      if (rem < 10) {
        s += ones[rem]
      } else if (rem === 10) {
        s += 'عشرة'
      } else if (rem === 11) {
        s += 'أحد عشر'
      } else if (rem === 12) {
        s += 'اثنا عشر'
      } else if (rem < 20) {
        s += ones[rem % 10] + ' عشر'
      } else { // 20-99
        let t = Math.floor(rem / 10)
        let o = rem % 10
        if (o > 0) s += ones[o] + ' و'
        s += tens[t]
      }
    }
    return s
  }

  const int = Math.floor(num)
  const decimal = Math.round((num - int) * 100)
  
  let result = ''
  let remaining = int
  
  // Billions
  let bil = Math.floor(remaining / 1000000000)
  remaining = remaining % 1000000000
  if (bil > 0) {
    if (bil === 1) result += 'مليار'
    else if (bil === 2) result += 'ملياران'
    else if (bil <= 10) result += convertThreeDigits(bil) + ' مليارات'
    else result += convertThreeDigits(bil) + ' مليار'
    if (remaining > 0) result += ' و'
  }

  // Millions
  let mil = Math.floor(remaining / 1000000)
  remaining = remaining % 1000000
  if (mil > 0) {
    if (mil === 1) result += 'مليون'
    else if (mil === 2) result += 'مليونان'
    else if (mil <= 10) result += convertThreeDigits(mil) + ' ملايين'
    else result += convertThreeDigits(mil) + ' مليون'
    if (remaining > 0) result += ' و'
  }
  
  // Thousands
  let th = Math.floor(remaining / 1000)
  remaining = remaining % 1000
  if (th > 0) {
    if (th === 1) result += 'ألف'
    else if (th === 2) result += 'ألفان'
    else if (th <= 10) result += convertThreeDigits(th) + ' آلاف'
    else result += convertThreeDigits(th) + ' ألف'
    if (remaining > 0) result += ' و'
  }
  
  // Hundreds/Tens/Ones
  if (remaining > 0) {
    result += convertThreeDigits(remaining)
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
          <div class="org-info">
            <h1>{{ orgInfo.nameAr }}</h1>
            <p class="org-contact">{{ orgInfo.phone }} | {{ orgInfo.email }}</p>
            <p class="org-address">{{ orgInfo.address }}</p>
          </div>
          <div class="org-logo">
            <div class="logo-placeholder">LOGO</div>
          </div>
        </div>

        <div class="divider"></div>

        <!-- Sub Header -->
        <div class="sub-header">
          <div class="receipt-type-label">
            <h2>{{ getTransactionTypeLabel }}</h2>
          </div>
          <div class="receipt-meta">
            <div class="meta-row">
              <span class="meta-label">رقم العملية:</span>
              <span class="meta-value">{{ receiptData.transaction.transaction_number }}</span>
            </div>
            <div class="meta-row">
              <span class="meta-label">التاريخ:</span>
              <span class="meta-value">{{ formatDate(receiptData.transaction.transaction_date) }}</span>
            </div>
          </div>
        </div>

        <!-- Centered Amount Section -->
        <div class="centered-amount-section">
          <div class="amount-label-large">
            {{ isIncome ? 'قيمة القبض' : (receiptData.transaction.type === 'transfer' ? 'مبلغ التحويل' : 'قيمة الصرف') }} 
            <!-- Add specific type text if needed, though header already has it. User asked for "Payment Amount (Type)" -->
            <span class="amount-type-suffix">({{ getTransactionTypeText }})</span>
          </div>
          <div class="amount-value-large">
            {{ receiptData.transaction.amount.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 }) }}
            <span class="currency-symbol">{{ getCurrencyInfo(receiptData.transaction.currency).symbol }}</span>
          </div>
        </div>

        <!-- Amount in Words (Compact & Centered) -->
        <div class="amount-words-centered">
          <span class="words-label">{{ isIncome ? 'استلمنا منكم مبلغ:' : 'صرف له مبلغ:' }}</span>
          <span class="words-value">{{ amountInWords }}</span>
        </div>

        <!-- Main Data Grid (Revised Logic) -->
        <div class="main-data-grid three-col">
          <!-- Column 1: Financial Account (The Org's Side) -->
          <div class="grid-cell">
            <span class="cell-label">{{ receiptData.transaction.type === 'transfer' ? 'من حساب:' : (isIncome ? 'إلى حساب (الخزينة/المصرف):' : 'من حساب (الخزينة/المصرف):') }}</span>
            <span class="cell-value">
              {{ receiptData.container.name_ar || receiptData.container.name }}
              <span class="sub-text">({{ receiptData.container.type === 'cash' ? 'نقدي' : 'مصرفي' }})</span>
            </span>
          </div>
          
          <!-- Column 2: The Other Party (Beneficiary or Source) -->
          <div class="grid-cell">
            <span class="cell-label">
              {{ isIncome ? 'استلمنا من:' : 'صرف لأمر:' }}
            </span>
            <span class="cell-value">
              <!-- Priority: Employee -> Member -> PayeePayer (External) -->
              <template v-if="receiptData.employee">{{ receiptData.employee.full_name_ar || receiptData.employee.full_name }} <span class="sub-text">(موظف)</span></template>
              <template v-else-if="receiptData.member">{{ receiptData.member.full_name_ar || receiptData.member.full_name }} <span class="sub-text">(عضو)</span></template>
              <template v-else-if="receiptData.transaction.payee_payer_ar || receiptData.transaction.payee_payer">{{ receiptData.transaction.payee_payer_ar || receiptData.transaction.payee_payer }}</template>
              <template v-else>-</template>
            </span>
          </div>

          <!-- Column 3: Project / Context -->
          <div class="grid-cell">
            <span class="cell-label">لصالح (المشروع) : </span>
            <span class="cell-value">
              {{ receiptData.project.name_ar || receiptData.project.name }}
            </span>
          </div>
        </div>

        <!-- Statement/Description -->
        <div class="statement-section" v-if="receiptData.transaction.description_ar || receiptData.transaction.description">
          <div class="statement-label">وذلك مقابل (البيان):</div>
          <div class="statement-text">{{ receiptData.transaction.description_ar || receiptData.transaction.description }}</div>
        </div>

        <!-- Optional: Cheque/Reference Info if exists -->
        <div class="ref-section" v-if="receiptData.transaction.reference_number">
          <span class="ref-label">رقم المرجع / الصك:</span>
          <span class="ref-value">{{ receiptData.transaction.reference_number }}</span>
        </div>

        <div class="divider"></div>

        <!-- Signatures (Improved) -->
        <div class="signatures-section">
          <div class="signature-box">
            <div class="signature-line"></div>
            <div class="signature-label">المستلم</div>
          </div>
          <div class="signature-box">
            <div class="signature-line"></div>
            <div class="signature-label">المحاسب</div>
            <div class="signature-name">{{ receiptData.created_by_user.full_name_ar || receiptData.created_by_user.full_name }}</div>
          </div>
          <div class="signature-box">
            <div class="signature-line"></div>
            <div class="signature-label">المدير المالي</div>
          </div>
          <div class="signature-box">
            <div class="signature-line"></div>
            <div class="signature-label">المدير العام</div>
          </div>
        </div>

        <!-- Footer -->
        <div class="receipt-footer">
          <p>حرر بواسطة النظام في: {{ formatDate(receiptData.transaction.created_at) }} - {{ formatTime(receiptData.transaction.created_at) }}</p>
        </div>
      </div>
    </div>
  </div>
</template>
<style scoped>
/* Keep existing container styles ... */
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

.receipt-content {
  width: 210mm;
  min-height: 148mm;
  background: white;
  padding: 15mm;
  box-shadow: 0 4px 16px rgba(0,0,0,0.1);
  border-radius: 8px;
}

/* Header Adjustments */
.receipt-header {
  display: flex;
  justify-content: space-between;
  align-items: center; /* Center Vertically */
  margin-bottom: 20px; /* More space */
}

/* Org Info on Right (Start in RTL) */
.org-info {
  text-align: right; /* Explicitly right aligned */
  flex: 1;
}

.org-info h1 {
  font-size: 24px;
  font-weight: 800;
  margin: 0 0 6px 0;
  color: #111;
}

.org-logo .logo-placeholder {
  width: 80px;
  height: 80px;
  background: linear-gradient(135deg, #2563eb 0%, #1e40af 100%);
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-weight: bold;
  font-size: 12px;
  box-shadow: 0 4px 6px rgba(0,0,0,0.1);
}

/* Sub Header */
.sub-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  margin: 10px 0 25px 0;
  padding-bottom: 15px;
  border-bottom: 2px solid #f3f4f6;
}

.receipt-type-label h2 {
  font-size: 20px;
  font-weight: 800;
  margin: 0;
  color: #2563eb;
}

.receipt-meta {
  display: flex;
  gap: 15px; /* Side by side now? Or stacked? User said "Transaction # (Left) Date (Left)" - usually stacked or inline */
  /* User previous request: "Display Transaction #: ... directly under it Date: ..." -> Stacked */
  flex-direction: column; 
  align-items: flex-end;
  gap: 6px;
}

.meta-row {
  display: flex;
  align-items: center;
  gap: 10px;
  background: #f8fafc;
  padding: 4px 12px;
  border-radius: 6px;
  border: 1px solid #e2e8f0;
}

.meta-label {
  font-weight: 700;
  color: #64748b;
  font-size: 11px;
}

.meta-value {
  font-weight: 700;
  color: #1e293b;
  font-size: 13px;
  font-family: monospace;
}

/* Centered Amount Section */
.centered-amount-section {
  text-align: center;
  margin-bottom: 10px;
}

.amount-label-large {
  font-size: 14px;
  color: #64748b;
  font-weight: 700;
  margin-bottom: 6px;
}

.amount-type-suffix {
  color: #2563eb;
}

.amount-value-large {
  font-size: 28px;
  font-weight: 800;
  color: #1e3a8a;
  background: linear-gradient(to right, transparent, #eff6ff, transparent);
  padding: 5px 0;
  border-radius: 8px; /* Visual touch */
}

.currency-symbol {
  font-size: 18px;
  color: #3b82f6;
  margin-right: 5px;
}

/* Amount Words Centered */
.amount-words-centered {
  text-align: center;
  margin-bottom: 25px;
  background: #fefce8;
  border: 1px dashed #facc15;
  padding: 8px;
  border-radius: 6px;
  display: inline-block; /* Shrink to fit */
  width: 100%; /* Or stick to full width? Center text */
}

.words-label {
  font-size: 12px;
  font-weight: 700;
  color: #a16207;
  margin-left: 8px;
}

.words-value {
  font-size: 14px;
  font-weight: 700;
  color: #713f12;
}

/* Main Data Grid - 3 Columns */
.main-data-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 15px;
  margin-bottom: 25px;
}

.grid-cell {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  padding: 12px 16px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.02);
}

.cell-label {
  font-size: 11px;
  color: #64748b;
  font-weight: 700;
  text-transform: uppercase;
}

.cell-value {
  font-size: 14px;
  color: #0f172a;
  font-weight: 600;
}

.sub-text {
  font-size: 11px;
  color: #64748b;
  font-weight: 400;
}

/* Statement */
.statement-section {
  background: #f9fafb;
  border-right: 4px solid #94a3b8;
  padding: 15px;
  margin: 15px 0;
  border-radius: 6px;
}

.statement-label {
  font-size: 12px;
  color: #475569;
  font-weight: 700;
  margin-bottom: 6px;
}

.statement-text {
  font-size: 14px;
  color: #334155;
  line-height: 1.6;
}

/* Reference */
.ref-section {
  margin: 10px 0;
  display: flex;
  gap: 10px;
  font-size: 13px;
}

.ref-label {
  font-weight: 700;
  color: #64748b;
}

/* Signatures */
.signatures-section {
  display: flex;
  justify-content: space-around;
  margin-top: 40px;
  margin-bottom: 20px;
  padding-top: 20px;
  border-top: 1px solid #e2e8f0;
}

.signature-box {
  width: 20%;
  text-align: center;
}

.signature-label {
  font-size: 13px;
  color: #64748b;
  font-weight: 700;
  margin-top: 8px;
}

.signature-line {
  border-bottom: 2px solid #cbd5e1;
  margin-bottom: 8px;
  height: 40px;
}

/* Print overrides */
@media print {
  .no-print { display: none !important; }
  .receipt-page {
    background: white;
    padding: 0;
    height: auto;
    display: block; 
  }
  .receipt-content {
    box-shadow: none;
    border: none;
    width: 100%;
    margin: 0;
    padding: 0;
  }
  
  .amount-value-large {
    background: none;
    border: 1px solid #eee;
  }
  
  .amount-words-centered {
    border: 1px solid #ccc;
    background: none;
  }
  
  .main-data-grid {
    gap: 10px;
  }
  
  .grid-cell {
    border: 1px solid #ccc;
    box-shadow: none;
  }
}

/* Responsive */
@media screen and (max-width: 900px) {
  .receipt-content {
    width: 100%;
    padding: 10mm;
  }
  .receipt-header {
    flex-direction: column-reverse; /* Stack logo on top or info? Usually Logo top on mobile */
    text-align: center;
  }
  .org-info { text-align: center; }
  
  .main-data-grid {
    grid-template-columns: 1fr;
  }
  
  .signatures-section {
    flex-wrap: wrap;
    gap: 20px;
  }
  .signature-box {
    width: 45%;
  }
}
</style>
