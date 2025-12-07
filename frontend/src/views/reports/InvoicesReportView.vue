<template>
  <div class="min-h-screen bg-gray-50 py-8" dir="rtl">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
      
      <!-- Header -->
      <div class="mb-6 no-print">
        <h1 class="text-3xl font-bold text-gray-900">تقرير الفواتير</h1>
        <p class="mt-2 text-sm text-gray-600">عرض تفصيلي للفواتير والإيرادات (منح وتبرعات)</p>
      </div>

      <!-- Filter Panel -->
      <div v-if="invoices.length === 0" class="bg-white rounded-lg shadow-lg p-6 mb-6 no-print">
        <h2 class="text-xl font-bold text-gray-900 mb-6 pb-3 border-b-2 border-primary-500">إعداد تقرير الفواتير</h2>
        
        <!-- Type Selection -->
        <div class="mb-6">
          <label class="block text-sm font-semibold text-gray-700 mb-3">نوع التقرير</label>
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <button @click="reportType = 'container'" :class="[reportType === 'container' ? 'bg-primary-50 border-primary-500 ring-2 ring-primary-500' : 'bg-white border-gray-300 hover:border-primary-300', 'relative rounded-lg border-2 p-5 flex items-center transition-all']">
              <div class="flex-shrink-0 text-4xl ml-4">🏦</div>
              <div class="flex-1 text-right">
                <span class="block text-lg font-bold text-gray-900">حسب الحساب المالي</span>
                <span class="block text-sm text-gray-600 mt-1">عرض الفواتير لحساب مصرفي أو خزينة</span>
              </div>
            </button>

            <button @click="reportType = 'project'" :class="[reportType === 'project' ? 'bg-primary-50 border-primary-500 ring-2 ring-primary-500' : 'bg-white border-gray-300 hover:border-primary-300', 'relative rounded-lg border-2 p-5 flex items-center transition-all']">
              <div class="flex-shrink-0 text-4xl ml-4">📊</div>
              <div class="flex-1 text-right">
                <span class="block text-lg font-bold text-gray-900">حسب المشروع</span>
                <span class="block text-sm text-gray-600 mt-1">عرض الفواتير الخاصة بمشروع معين</span>
              </div>
            </button>
          </div>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
          <!-- Container Selection -->
          <div v-if="reportType === 'container'">
            <label class="block text-sm font-semibold text-gray-700 mb-2">الحساب المصرفي / الخزينة</label>
            <select v-model="filters.containerID" class="block w-full rounded-lg border-2 border-gray-300 focus:border-primary-500 focus:ring-2 focus:ring-primary-500 py-3 px-4 transition-colors">
              <option value="">-- كل الحسابات --</option>
              <option v-for="container in containers" :key="container.id" :value="container.id">
                {{ container.name }}
              </option>
            </select>
          </div>

          <!-- Project Selection -->
          <div v-if="reportType === 'project'">
             <label class="block text-sm font-semibold text-gray-700 mb-2">المشروع</label>
            <select v-model="filters.projectID" class="block w-full rounded-lg border-2 border-gray-300 focus:border-primary-500 focus:ring-2 focus:ring-primary-500 py-3 px-4 transition-colors">
              <option value="">-- كل المشاريع --</option>
              <option v-for="project in projects" :key="project.id" :value="project.id">
                {{ project.name }} ({{ project.code }})
              </option>
            </select>
          </div>
          
          <!-- Transaction Number Search -->
          <div class="col-span-2">
             <label class="block text-sm font-semibold text-gray-700 mb-2">بحث برقم الفاتورة (اختياري)</label>
             <input type="text" v-model="filters.transactionNumber" placeholder="أدخل رقم الفاتورة..." class="block w-full rounded-lg border-2 border-gray-300 focus:border-primary-500 focus:ring-2 focus:ring-primary-500 py-2 px-3 transition-colors">
          </div>

          <!-- Date Range -->
          <div class="col-span-2 grid grid-cols-1 md:grid-cols-2 gap-6">
            <div>
                <label class="block text-sm font-semibold text-gray-700 mb-2">تاريخ البداية</label>
                <input type="date" v-model="filters.startDate" class="block w-full rounded-lg border-2 border-gray-300 focus:border-primary-500 focus:ring-2 focus:ring-primary-500 py-2 px-3 transition-colors">
            </div>
            <div>
                <label class="block text-sm font-semibold text-gray-700 mb-2">تاريخ النهاية</label>
                <input type="date" v-model="filters.endDate" class="block w-full rounded-lg border-2 border-gray-300 focus:border-primary-500 focus:ring-2 focus:ring-primary-500 py-2 px-3 transition-colors">
            </div>
          </div>
        </div>

        <div class="mt-8 flex justify-end">
          <button @click="fetchInvoices" :disabled="loading" class="w-full bg-primary-600 hover:bg-primary-700 disabled:bg-gray-400 text-white font-bold py-4 px-6 rounded-lg text-lg shadow-lg transition-all flex items-center justify-center">
            <MagnifyingGlassIcon v-if="!loading" class="-mr-1 ml-2 h-6 w-6" />
            <span v-if="loading">جاري إنشاء التقرير...</span>
            <span v-else>📄 عرض تقرير الفواتير</span>
          </button>
        </div>
      </div>

      <!-- Back Button when report is shown -->
      <div v-else class="mb-6 no-print">
         <button @click="resetSearch" class="bg-gray-600 hover:bg-gray-700 text-white font-bold py-2 px-4 rounded-lg shadow transition-all flex items-center">
            <span>← إعداد تقرير جديد</span>
         </button>
      </div>

      <!-- Report Display -->
      <div v-if="hasSearched && invoices.length > 0" id="printable-area" class="bg-white rounded-lg shadow-lg print:shadow-none print:rounded-none overflow-hidden max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        
        <!-- Official Print Header -->
        <div class="bg-gradient-to-r from-primary-600 to-primary-700 text-white p-8 rounded-t-lg print:bg-white print:text-black print:border-b-4 print:border-black">
          <div class="text-center">
            <div class="text-sm mb-2 print:text-black font-semibold">منظمات المجتمع المدني</div>
            <h1 class="text-3xl font-bold mb-2">نظام الإدارة المالية المتكامل</h1>
            <div class="text-xl font-semibold mt-4 print:text-black">
              📋 تقرير الفواتير
            </div>
            <div class="text-sm mt-3 print:text-gray-700">
               من: {{ formatDate(filters.startDate) || 'البداية' }} • إلى: {{ formatDate(filters.endDate) || 'النهاية' }}
            </div>
          </div>
        </div>

        <!-- Summary Box -->
        <div class="p-6 bg-gray-50 border-b border-gray-200 print:bg-white print:border-black">
             <div class="max-w-5xl mx-auto px-4 print:max-w-full print:px-0">
                <div class="grid grid-cols-1 md:grid-cols-3 gap-4 print:grid-cols-3 print:gap-2">
                    <div class="bg-white p-4 rounded-lg shadow-sm border-r-4 border-green-500 print:border-green-600 print:shadow-none print:p-2 print:text-right">
                        <div class="text-sm text-gray-500 mb-1 print:text-xs">إجمالي المقبوضات</div>
                        <div class="text-2xl font-bold text-green-700 print:text-sm">{{ formatNumber(totalIncome) }} <span class="text-sm font-normal text-gray-500 print:text-xs">د.ل</span></div>
                    </div>
                    <div class="bg-white p-4 rounded-lg shadow-sm border-r-4 border-red-500 print:border-red-600 print:shadow-none print:p-2 print:text-right">
                        <div class="text-sm text-gray-500 mb-1 print:text-xs">إجمالي المصروفات</div>
                        <div class="text-2xl font-bold text-red-700 print:text-sm">{{ formatNumber(totalExpenses) }} <span class="text-sm font-normal text-gray-500 print:text-xs">د.ل</span></div>
                    </div>
                    <div class="bg-white p-4 rounded-lg shadow-sm border-r-4 border-blue-500 print:border-blue-600 print:shadow-none print:p-2 print:text-right">
                        <div class="text-sm text-gray-500 mb-1 print:text-xs">إجمالي التحويلات / العهد</div>
                        <div class="text-2xl font-bold text-blue-700 print:text-sm">{{ formatNumber(totalTransfers) }} <span class="text-sm font-normal text-gray-500 print:text-xs">د.ل</span></div>
                    </div>
                </div>
             </div>
        </div>

        <!-- Transactions Table -->
        <div class="p-6">
          <div class="overflow-x-auto border-2 border-gray-200 rounded-lg print:border-black">
            <table class="min-w-full divide-y-2 divide-gray-200 print:divide-black">
              <thead class="bg-gray-800 text-white print:bg-gray-200 print:text-black">
                <tr>
                  <th scope="col" class="px-4 py-4 text-right text-sm font-bold whitespace-nowrap">رقم الفاتورة</th>
                  <th scope="col" class="px-4 py-4 text-right text-sm font-bold whitespace-nowrap">التاريخ</th>
                  <th scope="col" class="px-4 py-4 text-right text-sm font-bold">الجهة</th>
                  <th scope="col" class="px-4 py-4 text-right text-sm font-bold">النوع</th>
                  <th scope="col" class="px-4 py-4 text-right text-sm font-bold">المشروع</th>
                  <th scope="col" class="px-4 py-4 text-right text-sm font-bold">الحساب</th>
                  <th scope="col" class="px-4 py-4 text-right text-sm font-bold whitespace-nowrap">المبلغ</th>
                  <th scope="col" class="px-4 py-4 text-center text-sm font-bold no-print">إجراءات</th>
                </tr>
              </thead>
              <tbody class="bg-white divide-y divide-gray-200 print:divide-gray-400">
                <tr v-for="(invoice, index) in invoices" :key="invoice.id" :class="index % 2 === 0 ? 'bg-white' : 'bg-gray-50'">
                  <td class="px-4 py-3 whitespace-nowrap text-sm font-mono text-primary-600 font-bold">
                     <a :href="`/receipt/${invoice.id}`" target="_blank" class="hover:underline hover:text-primary-800">
                        {{ invoice.transaction_number }}
                    </a>
                  </td>
                  <td class="px-4 py-3 whitespace-nowrap text-sm font-medium">{{ formatDate(invoice.transaction_date) }}</td>
                  <td class="px-4 py-3 text-sm">{{ invoice.payee_payer || '-' }}</td>
                  <td class="px-4 py-3 text-sm">
                    <span 
                        class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium print:bg-transparent print:text-black print:p-0"
                        :class="isCredit(invoice.type) ? 'bg-green-100 text-green-800' : (invoice.type === 'transfer' ? 'bg-blue-100 text-blue-800' : 'bg-red-100 text-red-800')"
                    >
                        {{ getTransactionTypeName(invoice.type) }}
                    </span>
                  </td>
                  <td class="px-4 py-3 text-sm">{{ invoice.project_name || 'عام / غير محدد' }}</td>
                   <td class="px-4 py-3 text-sm">{{ invoice.container_name }}</td>
                  <td class="px-4 py-3 whitespace-nowrap text-sm font-bold print:text-black" :class="getTransactionColor(invoice.type)">
                    {{ formatNumber(invoice.amount) }} {{ invoice.currency }}
                  </td>
                  <td class="px-4 py-3 whitespace-nowrap text-center text-sm font-medium no-print">
                    <a 
                        :href="`/receipt/${invoice.id}`" 
                        target="_blank"
                        class="text-primary-600 hover:text-primary-900 hover:underline"
                    >
                      عرض
                    </a>
                  </td>
                </tr>
              </tbody>
               <tfoot class="bg-gray-100 print:bg-gray-200 border-t-4 border-gray-800 print:border-black">
                 <tr>
                    <td colspan="6" class="px-4 py-4 text-center font-bold text-lg">صافي الحركة:</td>
                    <td class="px-4 py-4 text-right font-bold text-xl print:text-black" :class="(totalIncome - totalExpenses) >= 0 ? 'text-green-600' : 'text-red-600'">{{ formatNumber(totalIncome - totalExpenses) }} د.ل</td>
                    <td class="no-print"></td>
                 </tr>
               </tfoot>
            </table>
          </div>
        </div>

        <!-- Action Buttons -->
        <div class="p-6 bg-gray-50 border-t border-gray-200 no-print flex justify-end">
            <button @click="printReport" class="flex items-center bg-gray-900 hover:bg-gray-800 text-white font-bold py-2 px-6 rounded-lg shadow-lg transition-all">
                <PrinterIcon class="h-5 w-5 ml-2" />
                طباعة التقرير
            </button>
        </div>

      </div>

      <!-- No Results State -->
      <div v-else-if="hasSearched" class="bg-white rounded-lg shadow p-12 text-center mt-6">
        <div class="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-gray-100 mb-4">
            <DocumentMagnifyingGlassIcon class="h-6 w-6 text-gray-400" />
        </div>
        <h3 class="text-lg font-medium text-gray-900">لا توجد فواتير</h3>
        <p class="mt-1 text-gray-500">جرب تغيير معايير البحث (التاريخ، المشروع، أو الحساب).</p>
      </div>

    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { PrinterIcon, MagnifyingGlassIcon, DocumentMagnifyingGlassIcon } from '@heroicons/vue/24/outline'
import api from '@/api'

const loading = ref(false)
const hasSearched = ref(false)
const reportType = ref('container') // Default to container
const invoices = ref([])
const projects = ref([])
const containers = ref([])
const filters = ref({
  startDate: '',
  endDate: '',
  projectID: '',
  containerID: '',
  transactionNumber: ''
})

const totalIncome = computed(() => {
    return invoices.value
        .filter(inv => isCredit(inv.type) || inv.type === 'transfer')
        .filter(inv => ['income_grant', 'income_donation'].includes(inv.type))
        .reduce((sum, inv) => sum + inv.amount, 0)
})

const totalExpenses = computed(() => {
    return invoices.value
        .filter(inv => !['income_grant', 'income_donation', 'transfer'].includes(inv.type))
        .reduce((sum, inv) => sum + inv.amount, 0)
})

const totalTransfers = computed(() => {
    return invoices.value
        .filter(inv => inv.type === 'transfer')
        .reduce((sum, inv) => sum + inv.amount, 0)
})

const resetSearch = () => {
    invoices.value = []
    hasSearched.value = false
    filters.value.projectID = ''
    filters.value.containerID = ''
    filters.value.transactionNumber = ''
}

const fetchFilterData = async () => {
    try {
        const [projectsRes, containersRes] = await Promise.all([
            api.get('/reports/project-summary'),
            api.get('/containers')
        ])
        projects.value = projectsRes.data
        containers.value = containersRes.data
    } catch (error) {
        console.error('Failed to load filters:', error)
    }
}
fetchFilterData()

const fetchInvoices = async () => {
  loading.value = true
  hasSearched.value = true
  try {
    const params = {}
    if (filters.value.startDate) params.start_date = filters.value.startDate
    if (filters.value.endDate) params.end_date = filters.value.endDate
    if (filters.value.projectID) params.project_id = filters.value.projectID
    if (filters.value.containerID) params.container_id = filters.value.containerID
    if (filters.value.transactionNumber) params.transaction_number = filters.value.transactionNumber
    
    const response = await api.get('/reports/invoices', { params })
    invoices.value = response.data
  } catch (error) {
    console.error('Failed to fetch invoices:', error)
    alert('حدث خطأ أثناء جلب الفواتير')
  } finally {
    loading.value = false
  }
}

const getTransactionTypeName = (type) => {
  const types = {
    'income_grant': 'منحة واردة',
    'income_donation': 'تبرع وارد',
    'disbursement_custody': 'عهدة مالية',
    'disbursement_salary': 'صرف مرتبات',
    'disbursement_operational': 'مصروفات تشغيلية',
    'disbursement_donation': 'تبرع صادر',
    'disbursement_bonus': 'مكافآت',
    'procurement': 'مشتريات',
    'transfer': 'تحويل'
  }
  return types[type] || type
}

const isCredit = (type) => {
  return ['income_grant', 'income_donation'].includes(type)
}

const getTransactionColor = (type) => {
    if (isCredit(type)) return 'text-green-600'
    if (type === 'transfer') return 'text-blue-600'
    return 'text-red-600'
}

const formatDate = (dateString) => {
  if (!dateString) return ''
  return new Date(dateString).toLocaleDateString('en-GB')
}

const formatNumber = (val) => {
  if (val === null || val === undefined) return '0.00'
  return Number(val).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

const printReport = () => {
  window.print()
}

onMounted(() => {
    fetchFilterData()
    // Optional: Auto fetch if you want
})
</script>

<style scoped>
@media print {
  .no-print { display: none !important; }
  
  body, .min-h-screen { 
    background-color: white !important; 
    height: auto !important;
  }

  body * {
    visibility: hidden;
  }

  #printable-area, #printable-area * {
    visibility: visible;
  }

  #printable-area {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    margin: 0;
    padding: 0;
  }
  
  /* Ensure background colors print */
  * {
    -webkit-print-color-adjust: exact !important; 
    print-color-adjust: exact !important; 
  }

  @page { margin: 1cm; size: A4 landscape; }
}
</style>
