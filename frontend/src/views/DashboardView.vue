<script setup>
import { ref, onMounted, computed } from 'vue'
import api from '@/api/client'
import { format } from 'date-fns'
import { ar } from 'date-fns/locale'
import { useAuthStore } from '@/stores/auth'
import * as XLSX from 'xlsx-js-style'

const authStore = useAuthStore()
const loading = ref(true)
const error = ref(null)
const dashboardData = ref(null)
const widgetConfig = ref([])

// Transaction Types Mapping (Arabic)
const transactionTypes = {
  'income_grant': 'منحة',
  'income_donation': 'تبرع',
  'disbursement_custody': 'عهدة',
  'disbursement_salary': 'مرتبات',
  'disbursement_operational': 'تشغيلية',
  'disbursement_donation': 'تبرع صادر',
  'disbursement_bonus': 'مكافأة',
  'procurement': 'مشتريات',
  'transfer': 'تحويل'
}

// Widget key mapping
const widgetKeys = {
  FINANCIAL_SUMMARY: 'financial_summary',
  BANK_ACCOUNTS: 'banks',
  CASH_ACCOUNTS: 'cash_accounts',
  PROJECTS: 'projects',
  RECENT_INCOME: 'recent_income',
  RECENT_EXPENSES: 'recent_expenses'
}

onMounted(async () => {
  try {
    loading.value = true
    
    // Fetch both dashboard data and widget configuration
    const [dashboardResponse, widgetResponse] = await Promise.all([
      api.get('/reports/dashboard'),
      api.get('/admin/dashboard-widgets').catch(() => ({ data: [] })) // Gracefully handle if not admin
    ])
    
    dashboardData.value = dashboardResponse.data
    const widgets = Array.isArray(widgetResponse.data) ? widgetResponse.data : []
    widgetConfig.value = widgets.sort((a, b) => a.display_order - b.display_order)
  } catch (err) {
    console.error('Failed to load dashboard data:', err)
    error.value = 'فشل تحميل بيانات لوحة التحكم. يرجى المحاولة مرة أخرى.'
  } finally {
    loading.value = false
  }
})

// Check if a widget should be displayed
const isWidgetVisible = (widgetKey) => {
  if (!widgetConfig.value || widgetConfig.value.length === 0) {
    return true // Show all widgets if config not loaded
  }
  
  const widget = widgetConfig.value.find(w => w.widget_key === widgetKey)
  if (!widget) return true // Show if not configured
  
  const userRole = authStore.user?.role
  return widget.is_enabled && widget.allowed_roles.includes(userRole)
}

// Get ordered visible widgets
const orderedWidgets = computed(() => {
  if (!widgetConfig.value || widgetConfig.value.length === 0) {
    // Default order if no config
    return [
      widgetKeys.FINANCIAL_SUMMARY,
      widgetKeys.BANK_ACCOUNTS,
      widgetKeys.CASH_ACCOUNTS,
      widgetKeys.PROJECTS,
      widgetKeys.RECENT_TRANSACTIONS
    ]
  }
  
  const userRole = authStore.user?.role
  return widgetConfig.value
    .filter(w => w.is_enabled && w.allowed_roles.includes(userRole))
    .sort((a, b) => a.display_order - b.display_order)
    .map(w => w.widget_key)
})

const formatCurrency = (amount, currency) => {
  console.log(`Formatting: ${amount} | Currency: '${currency}'`)
  const code = (currency || 'LYD').trim().toUpperCase()
  let locale = 'ar-LY'
  
  // Use specific locales for foreign currencies to ensure correct symbols
  if (code === 'USD') locale = 'en-US'
  if (code === 'EUR') locale = 'en-IE'
  if (code === 'GBP') locale = 'en-GB'
  
  try {
    return new Intl.NumberFormat(locale, {
      style: 'currency',
      currency: code
    }).format(amount)
  } catch (e) {
    console.error('Currency formatting error:', e)
    return `${amount} ${code}`
  }
}

const formatDate = (dateString) => {
  if (!dateString) return '-'
  return format(new Date(dateString), 'dd MMMM yyyy', { locale: ar })
}

const getTransactionLabel = (type) => {
  return transactionTypes[type] || type
}

const exporting = ref(false)

// Helper function to apply cell styling
const applyHeaderStyle = (cell) => {
  cell.s = {
    fill: { fgColor: { rgb: "4F46E5" } }, // Primary blue
    font: { bold: true, color: { rgb: "FFFFFF" }, sz: 12 },
    alignment: { horizontal: "center", vertical: "center" },
    border: {
      top: { style: "thin", color: { rgb: "000000" } },
      bottom: { style: "thin", color: { rgb: "000000" } },
      left: { style: "thin", color: { rgb: "000000" } },
      right: { style: "thin", color: { rgb: "000000" } }
    }
  }
  return cell
}

const applySectionHeaderStyle = (cell) => {
  cell.s = {
    fill: { fgColor: { rgb: "E0E7FF" } }, // Light blue
    font: { bold: true, sz: 14 },
    alignment: { horizontal: "right", vertical: "center" },
    border: {
      top: { style: "medium", color: { rgb: "4F46E5" } },
      bottom: { style: "medium", color: { rgb: "4F46E5" } },
      left: { style: "medium", color: { rgb: "4F46E5" } },
      right: { style: "medium", color: { rgb: "4F46E5" } }
    }
  }
  return cell
}

const applyDataCellStyle = (cell, isNumber = false, color = null) => {
  cell.s = {
    alignment: { horizontal: isNumber ? "right" : "right", vertical: "center" },
    border: {
      top: { style: "thin", color: { rgb: "D1D5DB" } },
      bottom: { style: "thin", color: { rgb: "D1D5DB" } },
      left: { style: "thin", color: { rgb: "D1D5DB" } },
      right: { style: "thin", color: { rgb: "D1D5DB" } }
    }
  }
  if (isNumber) {
    cell.s.font = { bold: true }
    if (color === 'green') {
      cell.s.font.color = { rgb: "059669" } // Green-600
    } else if (color === 'red') {
      cell.s.font.color = { rgb: "DC2626" } // Red-600
    }
  }
  return cell
}

// Helper function to set RTL (Right-to-Left) for individual sheets
const setRTL = (ws) => {
  // Set RTL view for this specific sheet
  if (!ws['!views']) ws['!views'] = [{}]
  ws['!views'][0] = { rtl: true }
}

// Export dashboard data to Excel with formatting
const exportToExcel = () => {
  try {
    exporting.value = true
    const wb = XLSX.utils.book_new()
    
    // Set workbook to RTL (Right-to-Left) for all sheets
    if (!wb.Workbook) wb.Workbook = {}
    if (!wb.Workbook.Views) wb.Workbook.Views = []
    wb.Workbook.Views[0] = { RTL: true }
    
    // Financial Summary Sheet - Horizontal Layout with Colors
    if (isWidgetVisible(widgetKeys.FINANCIAL_SUMMARY) && dashboardData.value?.summary) {
      const ws = XLSX.utils.aoa_to_sheet([])
      
      // Title Row
      XLSX.utils.sheet_add_aoa(ws, [['الملخص المالي']], { origin: 'A1' })
      ws['A1'] = applySectionHeaderStyle(ws['A1'])
      ws['!merges'] = [{ s: { r: 0, c: 0 }, e: { r: 0, c: 11 } }]
      
      // Headers Row
      const headers = ['إجمالي الرصيد', 'العملة', 'المبلغ', 'إجمالي المقبوضات', 'العملة', 'المبلغ', 'إجمالي المصروفات', 'العملة', 'المبلغ', 'إجمالي العهدة', 'العملة', 'المبلغ']
      XLSX.utils.sheet_add_aoa(ws, [headers], { origin: 'A3' })
      
      // Apply header styles with colors
      for (let col = 0; col < 12; col++) {
        const cellRef = XLSX.utils.encode_cell({ r: 2, c: col })
        ws[cellRef].s = {
          fill: { 
            fgColor: { 
              rgb: col < 3 ? "10B981" : // Green for Balance
                   col < 6 ? "3B82F6" : // Blue for Income
                   col < 9 ? "EF4444" : // Red for Expenses
                   "8B5CF6" // Purple for Custody
            } 
          },
          font: { bold: true, color: { rgb: "FFFFFF" }, sz: 11 },
          alignment: { horizontal: "center", vertical: "center" },
          border: {
            top: { style: "thin", color: { rgb: "000000" } },
            bottom: { style: "thin", color: { rgb: "000000" } },
            left: { style: "thin", color: { rgb: "000000" } },
            right: { style: "thin", color: { rgb: "000000" } }
          }
        }
      }
      
      // Data Row - collect all data
      const balanceData = dashboardData.value.summary.total_balance || []
      const incomeData = dashboardData.value.summary.total_income || []
      const expensesData = dashboardData.value.summary.total_expenses || []
      const custodyData = dashboardData.value.summary.total_custody || []
      
      // Create data rows (one row per currency if multiple)
      const maxRows = Math.max(balanceData.length, incomeData.length, expensesData.length, custodyData.length)
      
      for (let i = 0; i < maxRows; i++) {
        const row = [
          '', // Balance label (merged)
          balanceData[i]?.currency || '',
          balanceData[i]?.amount || 0,
          '', // Income label (merged)
          incomeData[i]?.currency || '',
          incomeData[i]?.amount || 0,
          '', // Expenses label (merged)
          expensesData[i]?.currency || '',
          expensesData[i]?.amount || 0,
          '', // Custody label (merged)
          custodyData[i]?.currency || '',
          custodyData[i]?.amount || 0
        ]
        
        XLSX.utils.sheet_add_aoa(ws, [row], { origin: `A${i + 4}` })
        
        // Apply data cell styles with light background colors
        for (let col = 0; col < 12; col++) {
          const cellRef = XLSX.utils.encode_cell({ r: i + 3, c: col })
          const isAmount = [2, 5, 8, 11].includes(col)
          
          ws[cellRef].s = {
            fill: { 
              fgColor: { 
                rgb: col < 3 ? "D1FAE5" : // Light green for Balance
                     col < 6 ? "DBEAFE" : // Light blue for Income
                     col < 9 ? "FEE2E2" : // Light red for Expenses
                     "EDE9FE" // Light purple for Custody
              } 
            },
            font: { bold: isAmount, sz: 10 },
            alignment: { horizontal: isAmount ? "right" : "center", vertical: "center" },
            border: {
              top: { style: "thin", color: { rgb: "D1D5DB" } },
              bottom: { style: "thin", color: { rgb: "D1D5DB" } },
              left: { style: "thin", color: { rgb: "D1D5DB" } },
              right: { style: "thin", color: { rgb: "D1D5DB" } }
            }
          }
        }
      }
      
      // Set column widths
      ws['!cols'] = [
        { wch: 18 }, { wch: 10 }, { wch: 15 }, // Balance
        { wch: 18 }, { wch: 10 }, { wch: 15 }, // Income
        { wch: 18 }, { wch: 10 }, { wch: 15 }, // Expenses
        { wch: 18 }, { wch: 10 }, { wch: 15 }  // Custody
      ]
      
      // Set RTL for Arabic
      setRTL(ws)
      
      XLSX.utils.book_append_sheet(wb, ws, 'الملخص المالي')
    }
    
    // Bank Accounts Sheet
    if (isWidgetVisible(widgetKeys.BANK_ACCOUNTS) && dashboardData.value?.bank_accounts?.length > 0) {
      const ws = XLSX.utils.aoa_to_sheet([])
      let rowIndex = 0
      
      // Title
      XLSX.utils.sheet_add_aoa(ws, [['الموقف المصرفي']], { origin: 'A1' })
      ws['A1'] = applySectionHeaderStyle(ws['A1'])
      ws['!merges'] = [{ s: { r: 0, c: 0 }, e: { r: 0, c: 3 } }]
      rowIndex += 2
      
      dashboardData.value.bank_accounts.forEach((bank, idx) => {
        // Bank header with separate columns
        XLSX.utils.sheet_add_aoa(ws, [[
          bank.container.bank_name_ar || bank.container.bank_name,
          bank.container.account_number,
          'الرصيد:',
          bank.container.balance,
          bank.container.currency
        ]], { origin: `A${rowIndex + 1}` })
        
        // Style bank header cells
        for (let col = 0; col < 5; col++) {
          const cellRef = XLSX.utils.encode_cell({ r: rowIndex, c: col })
          ws[cellRef] = applySectionHeaderStyle(ws[cellRef])
        }
        rowIndex++
        
        // Table headers
        XLSX.utils.sheet_add_aoa(ws, [['التاريخ', 'البيان', 'نوع العملية', 'القيمة', 'العملة']], { origin: `A${rowIndex + 1}` })
        for (let col = 0; col < 5; col++) {
          const cellRef = XLSX.utils.encode_cell({ r: rowIndex, c: col })
          applyHeaderStyle(ws[cellRef])
        }
        rowIndex++
        
        // Transactions
        bank.last_transactions?.forEach(tx => {
          const isIncome = ['income_grant', 'income_donation'].includes(tx.type)
          XLSX.utils.sheet_add_aoa(ws, [[
            formatDate(tx.transaction_date),
            tx.description_ar || tx.description,
            getTransactionLabel(tx.type),
            tx.amount,
            tx.currency
          ]], { origin: `A${rowIndex + 1}` })
          
          for (let col = 0; col < 5; col++) {
            const cellRef = XLSX.utils.encode_cell({ r: rowIndex, c: col })
            applyDataCellStyle(ws[cellRef], col === 3, col === 3 ? (isIncome ? 'green' : 'red') : null)
          }
          rowIndex++
        })
        rowIndex++
      })
      
      ws['!cols'] = [{ wch: 25 }, { wch: 20 }, { wch: 15 }, { wch: 12 }, { wch: 10 }]
      
      // Set RTL for Arabic
      setRTL(ws)
      
      XLSX.utils.book_append_sheet(wb, ws, 'الموقف المصرفي')
    }
    
    // Cash Accounts Sheet
    if (isWidgetVisible(widgetKeys.CASH_ACCOUNTS) && dashboardData.value?.cash_accounts?.length > 0) {
      const ws = XLSX.utils.aoa_to_sheet([])
      let rowIndex = 0
      
      XLSX.utils.sheet_add_aoa(ws, [['الموقف النقدي']], { origin: 'A1' })
      ws['A1'] = applySectionHeaderStyle(ws['A1'])
      ws['!merges'] = [{ s: { r: 0, c: 0 }, e: { r: 0, c: 3 } }]
      rowIndex += 2
      
      dashboardData.value.cash_accounts.forEach(cash => {
        // Cash header with separate columns
        XLSX.utils.sheet_add_aoa(ws, [[
          cash.container.name_ar || cash.container.name,
          cash.container.location_ar || cash.container.location || '',
          'الرصيد:',
          cash.container.balance,
          cash.container.currency
        ]], { origin: `A${rowIndex + 1}` })
        
        // Style cash header cells
        for (let col = 0; col < 5; col++) {
          const cellRef = XLSX.utils.encode_cell({ r: rowIndex, c: col })
          ws[cellRef] = applySectionHeaderStyle(ws[cellRef])
        }
        rowIndex++
        
        XLSX.utils.sheet_add_aoa(ws, [['التاريخ', 'البيان', 'نوع العملية', 'القيمة', 'العملة']], { origin: `A${rowIndex + 1}` })
        for (let col = 0; col < 5; col++) {
          applyHeaderStyle(ws[XLSX.utils.encode_cell({ r: rowIndex, c: col })])
        }
        rowIndex++
        
        cash.last_transactions?.forEach(tx => {
          const isIncome = ['income_grant', 'income_donation'].includes(tx.type)
          XLSX.utils.sheet_add_aoa(ws, [[
            formatDate(tx.transaction_date),
            tx.description_ar || tx.description,
            getTransactionLabel(tx.type),
            tx.amount,
            tx.currency
          ]], { origin: `A${rowIndex + 1}` })
          
          for (let col = 0; col < 5; col++) {
            applyDataCellStyle(ws[XLSX.utils.encode_cell({ r: rowIndex, c: col })], col === 3, col === 3 ? (isIncome ? 'green' : 'red') : null)
          }
          rowIndex++
        })
        rowIndex++
      })
      
      ws['!cols'] = [{ wch: 25 }, { wch: 20 }, { wch: 15 }, { wch: 12 }, { wch: 10 }]
      
      // Set RTL for Arabic
      setRTL(ws)
      
      XLSX.utils.book_append_sheet(wb, ws, 'الموقف النقدي')
    }
    
    // Projects Sheet
    if (isWidgetVisible(widgetKeys.PROJECTS) && dashboardData.value?.projects?.length > 0) {
      const ws = XLSX.utils.aoa_to_sheet([])
      let rowIndex = 0
      
      XLSX.utils.sheet_add_aoa(ws, [['المشاريع']], { origin: 'A1' })
      ws['A1'] = applySectionHeaderStyle(ws['A1'])
      ws['!merges'] = [{ s: { r: 0, c: 0 }, e: { r: 0, c: 3 } }]
      rowIndex += 2
      
      dashboardData.value.projects.forEach(project => {
        // Project header with separate columns
        XLSX.utils.sheet_add_aoa(ws, [[
          project.project.name_ar || project.project.name,
          project.project.code,
          'الميزانية:',
          project.project.budget || 0,
          project.project.currency
        ]], { origin: `A${rowIndex + 1}` })
        
        // Style project header cells
        for (let col = 0; col < 5; col++) {
          const cellRef = XLSX.utils.encode_cell({ r: rowIndex, c: col })
          ws[cellRef] = applySectionHeaderStyle(ws[cellRef])
        }
        rowIndex++
        
        XLSX.utils.sheet_add_aoa(ws, [['التاريخ', 'البيان', 'نوع العملية', 'القيمة', 'العملة']], { origin: `A${rowIndex + 1}` })
        for (let col = 0; col < 5; col++) {
          applyHeaderStyle(ws[XLSX.utils.encode_cell({ r: rowIndex, c: col })])
        }
        rowIndex++
        
        project.last_transactions?.forEach(tx => {
          const isIncome = ['income_grant', 'income_donation'].includes(tx.type)
          XLSX.utils.sheet_add_aoa(ws, [[
            formatDate(tx.transaction_date),
            tx.description_ar || tx.description,
            getTransactionLabel(tx.type),
            tx.amount,
            tx.currency
          ]], { origin: `A${rowIndex + 1}` })
          
          for (let col = 0; col < 5; col++) {
            applyDataCellStyle(ws[XLSX.utils.encode_cell({ r: rowIndex, c: col })], col === 3, col === 3 ? (isIncome ? 'green' : 'red') : null)
          }
          rowIndex++
        })
        rowIndex++
      })
      
      ws['!cols'] = [{ wch: 30 }, { wch: 15 }, { wch: 15 }, { wch: 12 }, { wch: 10 }]
      
      // Set RTL for Arabic
      setRTL(ws)
      
      XLSX.utils.book_append_sheet(wb, ws, 'المشاريع')
    }
    
    // Recent Income Sheet
    if (isWidgetVisible(widgetKeys.RECENT_INCOME) && dashboardData.value?.recent_income?.length > 0) {
      const ws = XLSX.utils.aoa_to_sheet([])
      
      XLSX.utils.sheet_add_aoa(ws, [['آخر المقبوضات']], { origin: 'A1' })
      ws['A1'] = applySectionHeaderStyle(ws['A1'])
      ws['!merges'] = [{ s: { r: 0, c: 0 }, e: { r: 0, c: 4 } }]
      
      XLSX.utils.sheet_add_aoa(ws, [['التاريخ', 'البيان', 'نوع العملية', 'الحساب', 'القيمة', 'العملة']], { origin: 'A3' })
      for (let col = 0; col < 6; col++) {
        applyHeaderStyle(ws[XLSX.utils.encode_cell({ r: 2, c: col })])
      }
      
      dashboardData.value.recent_income.forEach((tx, idx) => {
        XLSX.utils.sheet_add_aoa(ws, [[
          formatDate(tx.transaction_date),
          tx.description_ar || tx.description,
          getTransactionLabel(tx.type),
          tx.container_name,
          tx.amount,
          tx.currency
        ]], { origin: `A${idx + 4}` })
        
        for (let col = 0; col < 6; col++) {
          applyDataCellStyle(ws[XLSX.utils.encode_cell({ r: idx + 3, c: col })], col === 4, col === 4 ? 'green' : null)
        }
      })
      
      ws['!cols'] = [{ wch: 15 }, { wch: 35 }, { wch: 15 }, { wch: 20 }, { wch: 15 }, { wch: 10 }]
      
      // Set RTL for Arabic
      setRTL(ws)
      
      XLSX.utils.book_append_sheet(wb, ws, 'آخر المقبوضات')
    }
    
    // Recent Expenses Sheet
    if (isWidgetVisible(widgetKeys.RECENT_EXPENSES) && dashboardData.value?.recent_expenses?.length > 0) {
      const ws = XLSX.utils.aoa_to_sheet([])
      
      XLSX.utils.sheet_add_aoa(ws, [['آخر المصروفات']], { origin: 'A1' })
      ws['A1'] = applySectionHeaderStyle(ws['A1'])
      ws['!merges'] = [{ s: { r: 0, c: 0 }, e: { r: 0, c: 4 } }]
      
      XLSX.utils.sheet_add_aoa(ws, [['التاريخ', 'البيان', 'نوع العملية', 'الحساب', 'القيمة', 'العملة']], { origin: 'A3' })
      for (let col = 0; col < 6; col++) {
        applyHeaderStyle(ws[XLSX.utils.encode_cell({ r: 2, c: col })])
      }
      
      dashboardData.value.recent_expenses.forEach((tx, idx) => {
        XLSX.utils.sheet_add_aoa(ws, [[
          formatDate(tx.transaction_date),
          tx.description_ar || tx.description,
          getTransactionLabel(tx.type),
          tx.container_name,
          tx.amount,
          tx.currency
        ]], { origin: `A${idx + 4}` })
        
        for (let col = 0; col < 6; col++) {
          applyDataCellStyle(ws[XLSX.utils.encode_cell({ r: idx + 3, c: col })], col === 4, col === 4 ? 'red' : null)
        }
      })
      
      ws['!cols'] = [{ wch: 15 }, { wch: 35 }, { wch: 15 }, { wch: 20 }, { wch: 15 }, { wch: 10 }]
      
      // Set RTL for Arabic
      setRTL(ws)
      
      XLSX.utils.book_append_sheet(wb, ws, 'آخر المصروفات')
    }
    
    // Generate filename with current date
    const filename = `لوحة_التحكم_${format(new Date(), 'yyyy-MM-dd_HH-mm')}.xlsx`
    
    // Write file
    XLSX.writeFile(wb, filename)
    
  } catch (err) {
    console.error('Failed to export to Excel:', err)
    alert('فشل تصدير البيانات إلى Excel')
  } finally {
    exporting.value = false
  }
}
</script>

<template>
  <div class="p-6">
    <div class="flex justify-between items-center mb-8">
      <h1 class="text-3xl font-bold text-gray-800">لوحة التحكم</h1>
      <button 
        @click="exportToExcel" 
        :disabled="exporting || loading"
        class="flex items-center gap-2 bg-green-600 hover:bg-green-700 text-white px-4 py-2 rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
      >
        <span v-if="exporting" class="animate-spin h-5 w-5 border-2 border-white border-t-transparent rounded-full"></span>
        <span v-else>📊</span>
        <span>تصدير إلى Excel</span>
      </button>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="flex justify-center items-center h-64">
      <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600"></div>
      <span class="mr-3 text-gray-600">جاري تحميل البيانات...</span>
    </div>

    <!-- Error State -->
    <div v-else-if="error" class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded relative" role="alert">
      <strong class="font-bold">خطأ!</strong>
      <span class="block sm:inline mr-2">{{ error }}</span>
    </div>

    <!-- Dashboard Content -->
    <div v-else-if="dashboardData" class="space-y-8">
      
      <!-- Render widgets in configured order -->
      <template v-for="widgetKey in orderedWidgets" :key="widgetKey">
        
        <!-- Financial Summary Section -->
        <section v-if="widgetKey === widgetKeys.FINANCIAL_SUMMARY && dashboardData.summary && (dashboardData.summary.total_balance.length > 0 || dashboardData.summary.total_income.length > 0 || dashboardData.summary.total_expenses.length > 0 || dashboardData.summary.total_custody.length > 0)" class="bg-gradient-to-r from-primary-50 to-primary-100 rounded-lg shadow-lg p-6 border border-primary-200">
          <h2 class="text-2xl font-bold mb-6 text-gray-800 flex items-center">
            <span class="ml-2">📊</span> الملخص المالي
          </h2>
          <div class="grid grid-cols-1 md:grid-cols-4 gap-6">
            <!-- Total Balance -->
            <div class="bg-white rounded-lg p-4 shadow">
              <h3 class="text-sm font-semibold text-gray-600 mb-3">إجمالي الرصيد</h3>
              <div v-if="dashboardData.summary.total_balance && dashboardData.summary.total_balance.length > 0" class="space-y-2">
                <div v-for="amount in dashboardData.summary.total_balance" :key="amount.currency" class="text-lg font-bold text-green-600">
                  {{ formatCurrency(amount.amount, amount.currency) }}
                </div>
              </div>
              <div v-else class="text-gray-400">لا توجد بيانات</div>
            </div>

            <!-- Total Income -->
            <div class="bg-white rounded-lg p-4 shadow">
              <h3 class="text-sm font-semibold text-gray-600 mb-3">إجمالي المقبوضات</h3>
              <div v-if="dashboardData.summary.total_income && dashboardData.summary.total_income.length > 0" class="space-y-2">
                <div v-for="amount in dashboardData.summary.total_income" :key="amount.currency" class="text-lg font-bold text-blue-600">
                  {{ formatCurrency(amount.amount, amount.currency) }}
                </div>
              </div>
              <div v-else class="text-gray-400">لا توجد بيانات</div>
            </div>

            <!-- Total Expenses -->
            <div class="bg-white rounded-lg p-4 shadow">
              <h3 class="text-sm font-semibold text-gray-600 mb-3">إجمالي المصروفات</h3>
              <div v-if="dashboardData.summary.total_expenses && dashboardData.summary.total_expenses.length > 0" class="space-y-2">
                <div v-for="amount in dashboardData.summary.total_expenses" :key="amount.currency" class="text-lg font-bold text-orange-600">
                  {{ formatCurrency(amount.amount, amount.currency) }}
                </div>
              </div>
              <div v-else class="text-gray-400">لا توجد بيانات</div>
            </div>

            <!-- Total Custody -->
            <div class="bg-white rounded-lg p-4 shadow">
              <h3 class="text-sm font-semibold text-gray-600 mb-3">إجمالي العهدة</h3>
              <div v-if="dashboardData.summary.total_custody && dashboardData.summary.total_custody.length > 0" class="space-y-2">
                <div v-for="amount in dashboardData.summary.total_custody" :key="amount.currency" class="text-lg font-bold text-purple-600">
                  {{ formatCurrency(amount.amount, amount.currency) }}
                </div>
              </div>
              <div v-else class="text-gray-400">لا توجد بيانات</div>
            </div>
          </div>
        </section>
        
        <!-- Bank Accounts Section -->
        <section v-if="widgetKey === widgetKeys.BANK_ACCOUNTS && dashboardData.bank_accounts && dashboardData.bank_accounts.length > 0">
          <h2 class="text-xl font-bold mb-4 text-gray-700 flex items-center">
            <span class="ml-2">🏦</span> الموقف المصرفي (آخر الحركات)
          </h2>
          <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div v-for="bank in dashboardData.bank_accounts" :key="bank.container.id" class="bg-white rounded-lg shadow-md p-6 border border-gray-100">
              <div class="flex justify-between items-start mb-4 border-b pb-4">
                <div>
                  <h3 class="text-lg font-bold text-gray-900">{{ bank.container.bank_name_ar || bank.container.bank_name }}</h3>
                  <p class="text-sm text-gray-500">{{ bank.container.account_number }}</p>
                </div>
                <div class="text-left">
                  <p class="text-sm text-gray-500">الرصيد الحالي</p>
                  <p class="text-xl font-bold" :class="bank.container.balance >= 0 ? 'text-green-600' : 'text-red-600'">
                    {{ formatCurrency(bank.container.balance, bank.container.currency) }}
                  </p>
                </div>
              </div>
              
              <!-- Bank Transactions -->
              <div class="overflow-x-auto">
                <table class="min-w-full text-sm responsive-table">
                  <thead>
                    <tr class="bg-gray-50 text-gray-500">
                      <th class="px-3 py-2 text-right">التاريخ</th>
                      <th class="px-3 py-2 text-right">البيان</th>
                      <th class="px-3 py-2 text-right">نوع العملية</th>
                      <th class="px-3 py-2 text-right">القيمة</th>
                    </tr>
                  </thead>
                  <tbody class="divide-y divide-gray-100">
                    <tr v-for="tx in bank.last_transactions" :key="tx.id">
                      <td class="px-3 py-2 whitespace-nowrap" data-label="التاريخ">{{ formatDate(tx.transaction_date) }}</td>
                      <td class="px-3 py-2" data-label="البيان">{{ tx.description_ar || tx.description }}</td>
                      <td class="px-3 py-2 text-gray-600" data-label="نوع العملية">{{ getTransactionLabel(tx.type) }}</td>
                      <td class="px-3 py-2 font-medium" data-label="القيمة" :class="['income_grant', 'income_donation'].includes(tx.type) || (tx.type === 'transfer' && tx.balance_after > tx.balance_before) ? 'text-green-600' : 'text-red-600'">
                        {{ formatCurrency(tx.amount, tx.currency) }}
                      </td>
                    </tr>
                    <tr v-if="!bank.last_transactions || bank.last_transactions.length === 0">
                      <td colspan="4" class="px-3 py-4 text-center text-gray-400">لا توجد حركات حديثة</td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        </section>

        <!-- Cash Accounts Section -->
        <section v-if="widgetKey === widgetKeys.CASH_ACCOUNTS && dashboardData.cash_accounts && dashboardData.cash_accounts.length > 0">
          <h2 class="text-xl font-bold mb-4 text-gray-700 flex items-center">
            <span class="ml-2">💵</span> الموقف النقدي (آخر الحركات)
          </h2>
          <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div v-for="cash in dashboardData.cash_accounts" :key="cash.container.id" class="bg-white rounded-lg shadow-md p-6 border border-gray-100">
              <div class="flex justify-between items-start mb-4 border-b pb-4">
                <div>
                  <h3 class="text-lg font-bold text-gray-900">{{ cash.container.name_ar || cash.container.name }}</h3>
                  <p class="text-sm text-gray-500">{{ cash.container.location_ar || cash.container.location || 'صندوق نقدي' }}</p>
                </div>
                <div class="text-left">
                  <p class="text-sm text-gray-500">الرصيد الحالي</p>
                  <p class="text-xl font-bold" :class="cash.container.balance >= 0 ? 'text-green-600' : 'text-red-600'">
                    {{ formatCurrency(cash.container.balance, cash.container.currency) }}
                  </p>
                </div>
              </div>
              
              <!-- Cash Transactions -->
              <div class="overflow-x-auto">
                <table class="min-w-full text-sm responsive-table">
                  <thead>
                    <tr class="bg-gray-50 text-gray-500">
                      <th class="px-3 py-2 text-right">التاريخ</th>
                      <th class="px-3 py-2 text-right">البيان</th>
                      <th class="px-3 py-2 text-right">نوع العملية</th>
                      <th class="px-3 py-2 text-right">القيمة</th>
                    </tr>
                  </thead>
                  <tbody class="divide-y divide-gray-100">
                    <tr v-for="tx in cash.last_transactions" :key="tx.id">
                      <td class="px-3 py-2 whitespace-nowrap" data-label="التاريخ">{{ formatDate(tx.transaction_date) }}</td>
                      <td class="px-3 py-2" data-label="البيان">{{ tx.description_ar || tx.description }}</td>
                      <td class="px-3 py-2 text-gray-600" data-label="نوع العملية">{{ getTransactionLabel(tx.type) }}</td>
                      <td class="px-3 py-2 font-medium" data-label="القيمة" :class="['income_grant', 'income_donation'].includes(tx.type) || (tx.type === 'transfer' && tx.balance_after > tx.balance_before) ? 'text-green-600' : 'text-red-600'">
                        {{ formatCurrency(tx.amount, tx.currency) }}
                      </td>
                    </tr>
                    <tr v-if="!cash.last_transactions || cash.last_transactions.length === 0">
                      <td colspan="4" class="px-3 py-4 text-center text-gray-400">لا توجد حركات حديثة</td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        </section>

        <!-- Projects Section -->
        <section v-if="widgetKey === widgetKeys.PROJECTS && dashboardData.projects && dashboardData.projects.length > 0">
          <h2 class="text-xl font-bold mb-4 text-gray-700 flex items-center">
            <span class="ml-2">📁</span> المشاريع (آخر الحركات)
          </h2>
          <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div v-for="project in dashboardData.projects" :key="project.project.id" class="bg-white rounded-lg shadow-md p-6 border border-gray-100">
              <div class="flex justify-between items-start mb-4 border-b pb-4">
                <div>
                  <h3 class="text-lg font-bold text-gray-900">{{ project.project.name_ar || project.project.name }}</h3>
                  <p class="text-sm text-gray-500">{{ project.project.code }}</p>
                </div>
                <div class="text-left">
                  <p class="text-sm text-gray-500">الميزانية</p>
                  <p class="text-lg font-bold text-gray-700">
                    {{ project.project.budget ? formatCurrency(project.project.budget, project.project.currency) : '-' }}
                  </p>
                </div>
              </div>

              <!-- Project Transactions -->
              <div class="overflow-x-auto">
                <table class="min-w-full text-sm responsive-table">
                  <thead>
                    <tr class="bg-gray-50 text-gray-500">
                      <th class="px-3 py-2 text-right">التاريخ</th>
                      <th class="px-3 py-2 text-right">البيان</th>
                      <th class="px-3 py-2 text-right">نوع العملية</th>
                      <th class="px-3 py-2 text-right">القيمة</th>
                    </tr>
                  </thead>
                  <tbody class="divide-y divide-gray-100">
                    <tr v-for="tx in project.last_transactions" :key="tx.id">
                      <td class="px-3 py-2 whitespace-nowrap" data-label="التاريخ">{{ formatDate(tx.transaction_date) }}</td>
                      <td class="px-3 py-2" data-label="البيان">{{ tx.description_ar || tx.description }}</td>
                      <td class="px-3 py-2 text-gray-600" data-label="نوع العملية">{{ getTransactionLabel(tx.type) }}</td>
                      <td class="px-3 py-2 font-medium" data-label="القيمة" :class="['income_grant', 'income_donation'].includes(tx.type) ? 'text-green-600' : 'text-red-600'">
                        {{ formatCurrency(tx.amount, tx.currency) }}
                      </td>
                    </tr>
                    <tr v-if="!project.last_transactions || project.last_transactions.length === 0">
                      <td colspan="4" class="px-3 py-4 text-center text-gray-400">لا توجد حركات حديثة</td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        </section>

        <!-- Recent Income & Expenses (Combined Display) -->
        <!-- Only show once when we encounter either widget, and skip the second one -->
        <section 
          v-if="(widgetKey === widgetKeys.RECENT_INCOME && isWidgetVisible(widgetKeys.RECENT_INCOME)) || 
                (widgetKey === widgetKeys.RECENT_EXPENSES && isWidgetVisible(widgetKeys.RECENT_EXPENSES) && !isWidgetVisible(widgetKeys.RECENT_INCOME))" 
          class="grid grid-cols-1 lg:grid-cols-2 gap-6"
        >
          
          <!-- Recent Income -->
          <div v-if="isWidgetVisible(widgetKeys.RECENT_INCOME)" class="bg-white rounded-lg shadow-md p-6 border border-gray-100">
            <h2 class="text-lg font-bold mb-4 text-green-700 flex items-center">
              <span class="ml-2">📥</span> آخر المقبوضات
            </h2>
            <div class="overflow-x-auto">
              <table class="min-w-full text-sm responsive-table">
                <thead>
                  <tr class="bg-gray-50 text-gray-500">
                    <th class="px-3 py-2 text-right">التاريخ</th>
                    <th class="px-3 py-2 text-right">البيان</th>
                    <th class="px-3 py-2 text-right">نوع العملية</th>
                    <th class="px-3 py-2 text-right">القيمة</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-gray-100">
                  <tr v-for="tx in dashboardData.recent_income" :key="tx.id" class="hover:bg-gray-50">
                    <td class="px-3 py-2 whitespace-nowrap text-gray-600" data-label="التاريخ">{{ formatDate(tx.transaction_date) }}</td>
                    <td class="px-3 py-2" data-label="البيان">
                      <div class="font-medium text-gray-900">{{ tx.description_ar || tx.description }}</div>
                      <div class="text-xs text-gray-500">{{ tx.container_name }}</div>
                    </td>
                    <td class="px-3 py-2 text-gray-600" data-label="نوع العملية">{{ getTransactionLabel(tx.type) }}</td>
                    <td class="px-3 py-2 font-bold text-green-600" data-label="القيمة">
                      {{ formatCurrency(tx.amount, tx.currency) }}
                    </td>
                  </tr>
                  <tr v-if="!dashboardData.recent_income || dashboardData.recent_income.length === 0">
                    <td colspan="4" class="px-3 py-8 text-center text-gray-400">لا توجد مقبوضات حديثة</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <!-- Recent Expenses -->
          <div v-if="isWidgetVisible(widgetKeys.RECENT_EXPENSES)" class="bg-white rounded-lg shadow-md p-6 border border-gray-100">
            <h2 class="text-lg font-bold mb-4 text-red-700 flex items-center">
              <span class="ml-2">📤</span> آخر المصروفات
            </h2>
            <div class="overflow-x-auto">
              <table class="min-w-full text-sm responsive-table">
                <thead>
                  <tr class="bg-gray-50 text-gray-500">
                    <th class="px-3 py-2 text-right">التاريخ</th>
                    <th class="px-3 py-2 text-right">البيان</th>
                    <th class="px-3 py-2 text-right">نوع العملية</th>
                    <th class="px-3 py-2 text-right">القيمة</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-gray-100">
                  <tr v-for="tx in dashboardData.recent_expenses" :key="tx.id" class="hover:bg-gray-50">
                    <td class="px-3 py-2 whitespace-nowrap text-gray-600" data-label="التاريخ">{{ formatDate(tx.transaction_date) }}</td>
                    <td class="px-3 py-2" data-label="البيان">
                      <div class="font-medium text-gray-900">{{ tx.description_ar || tx.description }}</div>
                      <div class="text-xs text-gray-500">{{ tx.container_name }}</div>
                    </td>
                    <td class="px-3 py-2 text-gray-600" data-label="نوع العملية">{{ getTransactionLabel(tx.type) }}</td>
                    <td class="px-3 py-2 font-bold text-red-600" data-label="القيمة">
                      {{ formatCurrency(tx.amount, tx.currency) }}
                    </td>
                  </tr>
                  <tr v-if="!dashboardData.recent_expenses || dashboardData.recent_expenses.length === 0">
                    <td colspan="4" class="px-3 py-8 text-center text-gray-400">لا توجد مصروفات حديثة</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

        </section>

      </template>

    </div>

    <!-- Empty State (if data loaded but empty) -->
    <div v-else class="text-center py-12 bg-white rounded-lg shadow">
      <p class="text-gray-500 text-lg">لا توجد بيانات لعرضها حالياً</p>
    </div>
  </div>
</template>
