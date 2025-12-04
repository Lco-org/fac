<template>
  <div class="min-h-screen bg-gray-50 py-6">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
      <!-- Header -->
      <div class="mb-6 no-print">
        <h1 class="text-3xl font-bold text-gray-900">الجرد المفصل</h1>
        <p class="mt-2 text-sm text-gray-600">كشف حساب تفصيلي للحسابات المالية والمشاريع</p>
      </div>

      <!-- Selection Panel -->
      <div v-if="!statement" class="bg-white rounded-lg shadow-lg p-6 mb-6">
        <h2 class="text-xl font-bold text-gray-900 mb-6 pb-3 border-b-2 border-primary-500">إعداد الكشف</h2>
        
        <!-- Type Selection -->
        <div class="mb-6">
          <label class="block text-sm font-semibold text-gray-700 mb-3">نوع الكشف</label>
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <button @click="statementType = 'container'" :class="[statementType === 'container' ? 'bg-primary-50 border-primary-500 ring-2 ring-primary-500' : 'bg-white border-gray-300 hover:border-primary-300', 'relative rounded-lg border-2 p-5 flex items-center transition-all']">
              <div class="flex-shrink-0 text-4xl ml-4">🏦</div>
              <div class="flex-1 text-right">
                <span class="block text-lg font-bold text-gray-900">كشف حساب مالي</span>
                <span class="block text-sm text-gray-600 mt-1">الحسابات المصرفية والخزائن النقدية</span>
              </div>
            </button>

            <button @click="statementType = 'project'" :class="[statementType === 'project' ? 'bg-primary-50 border-primary-500 ring-2 ring-primary-500' : 'bg-white border-gray-300 hover:border-primary-300', 'relative rounded-lg border-2 p-5 flex items-center transition-all']">
              <div class="flex-shrink-0 text-4xl ml-4">📊</div>
              <div class="flex-1 text-right">
                <span class="block text-lg font-bold text-gray-900">كشف حساب مشروع</span>
                <span class="block text-sm text-gray-600 mt-1">المشاريع والبرامج التنموية</span>
              </div>
            </button>
          </div>
        </div>

        <!-- Container/Project Selection -->
        <div class="mb-6">
          <label class="block text-sm font-semibold text-gray-700 mb-3">
            {{ statementType === 'container' ? 'اختر الحساب المالي' : 'اختر المشروع' }}
          </label>
          
          <select v-if="statementType === 'container'" v-model="selectedContainer" class="block w-full text-lg rounded-lg border-2 border-gray-300 focus:border-primary-500 focus:ring-2 focus:ring-primary-500 py-3 px-4">
            <option value="">-- اختر الحساب --</option>
            <option v-for="container in containers" :key="container.id" :value="container.id">
              {{ container.name }} • {{ container.type === 'bank' ? '🏦 مصرفي' : '💰 نقدي' }} • {{ formatNumber(container.balance) }} {{ container.currency }}
            </option>
          </select>

          <select v-if="statementType === 'project'" v-model="selectedProject" class="block w-full text-lg rounded-lg border-2 border-gray-300 focus:border-primary-500 focus:ring-2 focus:ring-primary-500 py-3 px-4">
            <option value="">-- اختر المشروع --</option>
            <option v-for="project in projects" :key="project.id" :value="project.id">
              {{ project.code }} • {{ project.name }}
            </option>
          </select>
        </div>

        <!-- Date Range -->
        <div class="mb-6">
          <label class="block text-sm font-semibold text-gray-700 mb-3">الفترة الزمنية (اختياري)</label>
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label class="block text-xs text-gray-600 mb-1">من تاريخ</label>
              <input type="date" v-model="startDate" class="block w-full rounded-lg border-2 border-gray-300 focus:border-primary-500 focus:ring-2 focus:ring-primary-500 py-2 px-3">
            </div>
            <div>
              <label class="block text-xs text-gray-600 mb-1">إلى تاريخ</label>
              <input type="date" v-model="endDate" class="block w-full rounded-lg border-2 border-gray-300 focus:border-primary-500 focus:ring-2 focus:ring-primary-500 py-2 px-3">
            </div>
          </div>
        </div>

        <!-- Generate Button -->
        <button @click="generateStatement" :disabled="!canGenerate" class="w-full bg-primary-600 hover:bg-primary-700 disabled:bg-gray-400 text-white font-bold py-4 px-6 rounded-lg text-lg shadow-lg transition-all disabled:cursor-not-allowed">
          <span v-if="canGenerate">📄 عرض الكشف التفصيلي</span>
          <span v-else>يرجى اختيار {{ statementType === 'container' ? 'الحساب' : 'المشروع' }}</span>
        </button>
      </div>

      <!-- Statement Display -->
      <div v-if="statement" class="bg-white rounded-lg shadow-lg print:shadow-none print:rounded-none">
        <!-- Official Header -->
        <div class="bg-gradient-to-r from-primary-600 to-primary-700 text-white p-8 rounded-t-lg print:bg-white print:text-black print:border-b-4 print:border-black">
          <div class="text-center">
            <div class="text-sm mb-2 print:text-black">منظمات المجتمع المدني</div>
            <h1 class="text-3xl font-bold mb-2">نظام الإدارة المالية المتكامل</h1>
            <div class="text-xl font-semibold mt-4 print:text-black">
              {{ statementType === 'container' ? '📋 كشف حساب مالي تفصيلي' : '📊 كشف حساب مشروع تفصيلي' }}
            </div>
            <div class="text-sm mt-3 print:text-gray-700">
              تاريخ الإصدار: {{ formatDate(statement.generated_at) }} • {{ formatTime(statement.generated_at) }}
            </div>
          </div>
        </div>

        <!-- Account/Project Info -->
        <div class="p-8 bg-gray-50 print:bg-white border-b-2 border-gray-200 print:border-black">
          <h3 class="text-lg font-bold text-gray-900 mb-4">بيانات {{ statementType === 'container' ? 'الحساب' : 'المشروع' }}</h3>
          
          <div v-if="statementType === 'container'" class="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div class="bg-white p-4 rounded-lg border border-gray-200 print:border-gray-400">
              <div class="text-xs text-gray-500 mb-1">اسم الحساب</div>
              <div class="text-lg font-bold text-gray-900">{{ statement.container.name }}</div>
            </div>
            <div class="bg-white p-4 rounded-lg border border-gray-200 print:border-gray-400">
              <div class="text-xs text-gray-500 mb-1">نوع الحساب</div>
              <div class="text-lg font-bold text-gray-900">{{ statement.container.type === 'bank' ? '🏦 حساب مصرفي' : '💰 خزينة نقدية' }}</div>
            </div>
            <div v-if="statement.container.bank_name" class="bg-white p-4 rounded-lg border border-gray-200 print:border-gray-400">
              <div class="text-xs text-gray-500 mb-1">اسم البنك</div>
              <div class="text-lg font-bold text-gray-900">{{ statement.container.bank_name }}</div>
            </div>
            <div v-if="statement.container.account_number" class="bg-white p-4 rounded-lg border border-gray-200 print:border-gray-400">
              <div class="text-xs text-gray-500 mb-1">رقم الحساب</div>
              <div class="text-lg font-bold text-gray-900 font-mono">{{ statement.container.account_number }}</div>
            </div>
            <div class="bg-white p-4 rounded-lg border border-gray-200 print:border-gray-400">
              <div class="text-xs text-gray-500 mb-1">العملة</div>
              <div class="text-lg font-bold text-gray-900">{{ getCurrencyName(statement.container.currency) }}</div>
            </div>
          </div>

          <div v-if="statementType === 'project'" class="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div class="bg-white p-4 rounded-lg border border-gray-200 print:border-gray-400">
              <div class="text-xs text-gray-500 mb-1">كود المشروع</div>
              <div class="text-lg font-bold text-gray-900 font-mono">{{ statement.project.code }}</div>
            </div>
            <div class="bg-white p-4 rounded-lg border border-gray-200 print:border-gray-400">
              <div class="text-xs text-gray-500 mb-1">اسم المشروع</div>
              <div class="text-lg font-bold text-gray-900">{{ statement.project.name }}</div>
            </div>
            <div v-if="statement.project.donor_name" class="bg-white p-4 rounded-lg border border-gray-200 print:border-gray-400">
              <div class="text-xs text-gray-500 mb-1">الجهة المانحة</div>
              <div class="text-lg font-bold text-gray-900">{{ statement.project.donor_name }}</div>
            </div>
            <div v-if="statement.project.budget" class="bg-white p-4 rounded-lg border border-gray-200 print:border-gray-400">
              <div class="text-xs text-gray-500 mb-1">الميزانية المعتمدة</div>
              <div class="text-lg font-bold text-gray-900">{{ formatNumber(statement.project.budget) }} {{ statement.project.currency }}</div>
            </div>
          </div>
        </div>

        <!-- Summary -->
        <div class="p-8 bg-gradient-to-br from-blue-50 to-indigo-50 print:bg-gray-100 border-b-2 border-gray-200 print:border-black">
          <h3 class="text-xl font-bold text-gray-900 mb-6">📊 الملخص المالي</h3>
          <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
            <div v-if="statementType === 'container'" class="bg-white p-5 rounded-xl shadow-sm border-r-4 border-blue-500 print:border-2 print:border-gray-400">
              <div class="text-xs text-gray-600 mb-1">الرصيد الافتتاحي</div>
              <div class="text-2xl font-bold text-blue-600">{{ formatNumber(statement.summary.opening_balance) }}</div>
            </div>
            <div class="bg-white p-5 rounded-xl shadow-sm border-r-4 border-green-500 print:border-2 print:border-gray-400">
              <div class="text-xs text-gray-600 mb-1">{{ statementType === 'container' ? 'إجمالي المقبوضات' : 'إجمالي الإيرادات' }}</div>
              <div class="text-2xl font-bold text-green-600">{{ formatNumber(statement.summary.total_income) }}</div>
            </div>
            <div class="bg-white p-5 rounded-xl shadow-sm border-r-4 border-red-500 print:border-2 print:border-gray-400">
              <div class="text-xs text-gray-600 mb-1">{{ statementType === 'container' ? 'إجمالي المدفوعات' : 'إجمالي المصروفات' }}</div>
              <div class="text-2xl font-bold text-red-600">{{ formatNumber(statement.summary.total_expenses) }}</div>
            </div>
            <div class="bg-white p-5 rounded-xl shadow-sm border-r-4 border-primary-500 print:border-2 print:border-gray-400">
              <div class="text-xs text-gray-600 mb-1">{{ statementType === 'container' ? 'الرصيد الختامي' : 'صافي المركز' }}</div>
              <div class="text-2xl font-bold text-primary-600">{{ formatNumber(statementType === 'container' ? statement.summary.closing_balance : statement.summary.net_position) }}</div>
            </div>
          </div>
          <div class="mt-4 bg-white p-4 rounded-lg border border-gray-200 print:border-gray-400">
            <div class="text-sm text-gray-600">عدد الحركات المالية: <span class="font-bold text-gray-900">{{ statement.summary.transaction_count }} حركة</span></div>
          </div>
        </div>

        <!-- Transactions Table -->
        <div class="p-8">
          <h3 class="text-xl font-bold text-gray-900 mb-6">📝 تفاصيل الحركات المالية</h3>
          <div class="overflow-x-auto border-2 border-gray-200 rounded-lg print:border-black">
            <table class="min-w-full divide-y-2 divide-gray-200 print:divide-black">
              <thead class="bg-gray-800 text-white print:bg-gray-200 print:text-black">
                <tr>
                  <th class="px-4 py-4 text-right text-sm font-bold">التاريخ</th>
                  <th class="px-4 py-4 text-right text-sm font-bold">رقم المستند</th>
                  <th class="px-4 py-4 text-right text-sm font-bold">البيان</th>
                  <th v-if="statementType === 'container'" class="px-4 py-4 text-right text-sm font-bold">المشروع</th>
                  <th v-if="statementType === 'project'" class="px-4 py-4 text-right text-sm font-bold">الحساب</th>
                  <th class="px-4 py-4 text-center text-sm font-bold bg-green-700 print:bg-gray-300">مدين (إضافة)</th>
                  <th class="px-4 py-4 text-center text-sm font-bold bg-red-700 print:bg-gray-300">دائن (خصم)</th>
                  <th v-if="statementType === 'container'" class="px-4 py-4 text-center text-sm font-bold bg-blue-700 print:bg-gray-300">الرصيد</th>
                </tr>
              </thead>
              <tbody class="bg-white divide-y divide-gray-200 print:divide-gray-400">
                <tr v-for="(transaction, index) in statement.transactions" :key="transaction.id" :class="index % 2 === 0 ? 'bg-white' : 'bg-gray-50'">
                  <td class="px-4 py-3 whitespace-nowrap text-sm font-medium">{{ formatDate(transaction.transaction_date) }}</td>
                  <td class="px-4 py-3 whitespace-nowrap text-sm font-mono text-primary-600 font-bold">{{ transaction.transaction_number }}</td>
                  <td class="px-4 py-3 text-sm">
                    <div class="font-semibold">{{ getTransactionTypeName(transaction.type) }}</div>
                    <div class="text-xs text-gray-600">{{ transaction.description }}</div>
                    <div v-if="transaction.payee_payer" class="text-xs text-gray-500 mt-1">{{ transaction.payee_payer }}</div>
                  </td>
                  <td v-if="statementType === 'container'" class="px-4 py-3 text-sm font-medium">{{ transaction.project_name }}</td>
                  <td v-if="statementType === 'project'" class="px-4 py-3 text-sm">{{ transaction.container_name }}</td>
                  <td class="px-4 py-3 text-center">
                    <span v-if="isCredit(transaction)" class="text-lg font-bold text-green-600">{{ formatNumber(transaction.amount) }}</span>
                    <span v-else class="text-gray-300">-</span>
                  </td>
                  <td class="px-4 py-3 text-center">
                    <span v-if="!isCredit(transaction)" class="text-lg font-bold text-red-600">{{ formatNumber(transaction.amount) }}</span>
                    <span v-else class="text-gray-300">-</span>
                  </td>
                  <td v-if="statementType === 'container'" class="px-4 py-3 text-center">
                    <span class="text-lg font-bold text-blue-600">{{ formatNumber(transaction.balance_after) }}</span>
                  </td>
                </tr>
              </tbody>
              <tfoot class="bg-gray-100 print:bg-gray-200 border-t-4 border-gray-800 print:border-black">
                <tr>
                  <td colspan="{{ statementType === 'container' ? '4' : '3' }}" class="px-4 py-4 text-right font-bold text-lg">الإجمالي:</td>
                  <td class="px-4 py-4 text-center font-bold text-xl text-green-600">{{ formatNumber(statement.summary.total_income) }}</td>
                  <td class="px-4 py-4 text-center font-bold text-xl text-red-600">{{ formatNumber(statement.summary.total_expenses) }}</td>
                  <td v-if="statementType === 'container'" class="px-4 py-4 text-center font-bold text-xl text-blue-600">{{ formatNumber(statement.summary.closing_balance) }}</td>
                </tr>
              </tfoot>
            </table>
          </div>
        </div>

        <!-- Footer -->
        <div class="p-8 bg-gray-50 print:bg-white border-t-2 border-gray-200 print:border-black">
          <div class="grid grid-cols-3 gap-8 text-center print:mt-20">
            <div>
              <div class="border-t-2 border-gray-400 pt-2 mt-16">
                <div class="text-sm font-semibold">المحاسب</div>
              </div>
            </div>
            <div>
              <div class="border-t-2 border-gray-400 pt-2 mt-16">
                <div class="text-sm font-semibold">المدير المالي</div>
              </div>
            </div>
            <div>
              <div class="border-t-2 border-gray-400 pt-2 mt-16">
                <div class="text-sm font-semibold">المدير التنفيذي</div>
              </div>
            </div>
          </div>
        </div>

        <!-- Action Buttons -->
        <div class="p-6 bg-white border-t border-gray-200 no-print flex gap-4">
          <button @click="printStatement" class="flex-1 bg-primary-600 hover:bg-primary-700 text-white font-bold py-3 px-6 rounded-lg shadow-lg transition-all">
            🖨️ طباعة الكشف
          </button>
          <button @click="statement = null" class="flex-1 bg-gray-600 hover:bg-gray-700 text-white font-bold py-3 px-6 rounded-lg shadow-lg transition-all">
            ← عودة
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import api from '@/api'

const statementType = ref('container')
const selectedContainer = ref('')
const selectedProject = ref('')
const startDate = ref('')
const endDate = ref('')
const statement = ref(null)
const containers = ref([])
const projects = ref([])

const canGenerate = computed(() => {
  if (statementType.value === 'container') return selectedContainer.value !== ''
  else return selectedProject.value !== ''
})

const fetchContainers = async () => {
  try {
    const response = await api.get('/containers')
    containers.value = response.data
  } catch (error) {
    console.error('Failed to fetch containers:', error)
  }
}

const fetchProjects = async () => {
  try {
    const response = await api.get('/projects')
    projects.value = response.data
  } catch (error) {
    console.error('Failed to fetch projects:', error)
  }
}

const generateStatement = async () => {
  try {
    let url = statementType.value === 'container' 
      ? `/reports/containers/${selectedContainer.value}/statement`
      : `/reports/projects/${selectedProject.value}/statement`

    const params = new URLSearchParams()
    if (startDate.value) params.append('start_date', startDate.value)
    if (endDate.value) params.append('end_date', endDate.value)
    const queryString = params.toString()
    if (queryString) url += `?${queryString}`

    const response = await api.get(url)
    statement.value = response.data
  } catch (error) {
    console.error('Failed to generate statement:', error)
    alert('حدث خطأ أثناء إنشاء الكشف')
  }
}

const printStatement = () => window.print()

const formatDate = (dateString) => {
  if (!dateString) return ''
  return new Date(dateString).toLocaleDateString('ar-EG', { year: 'numeric', month: '2-digit', day: '2-digit' })
}

const formatTime = (dateString) => {
  if (!dateString) return ''
  return new Date(dateString).toLocaleTimeString('ar-EG', { hour: '2-digit', minute: '2-digit' })
}

const formatNumber = (number) => {
  if (number === null || number === undefined) return '0.00'
  // Use 'en-US' for Western numerals (123456) with comma separators
  return Number(number).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

const getCurrencyName = (code) => {
  const currencies = { 'LYD': 'دينار ليبي', 'USD': 'دولار أمريكي', 'EUR': 'يورو' }
  return currencies[code] || code
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

const isCredit = (transaction) => {
  if (['income_grant', 'income_donation'].includes(transaction.type)) return true
  if (transaction.type === 'transfer' && transaction.balance_after > transaction.balance_before) return true
  return false
}

onMounted(() => {
  fetchContainers()
  fetchProjects()
})
</script>

<style scoped>
@media print {
  .no-print { display: none !important; }
  body { print-color-adjust: exact; -webkit-print-color-adjust: exact; }
  @page { margin: 1cm; size: A4; }
}
</style>
