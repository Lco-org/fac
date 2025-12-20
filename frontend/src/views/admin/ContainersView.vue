<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex justify-between items-center">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">إدارة الحسابات المصرفية</h1>
        <p class="mt-1 text-sm text-gray-500">إدارة الصناديق النقدية والحسابات المصرفية</p>
      </div>
      <button
        type="button"
        @click="openCreateModal"
        class="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-primary-600 hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500"
      >
        <PlusIcon class="-mr-1 ml-2 h-5 w-5" aria-hidden="true" />
        حساب جديد
      </button>
    </div>

    <!-- Containers List -->
    <div class="bg-white shadow overflow-hidden sm:rounded-lg">
      <ul role="list" class="divide-y divide-gray-200">
        <li v-for="container in (containers || [])" :key="container.id" class="px-4 py-4 sm:px-6 hover:bg-gray-50">
          <div class="flex items-center justify-between">
            <div class="flex items-center min-w-0">
              <div class="flex-shrink-0">
                <span class="inline-flex items-center justify-center h-10 w-10 rounded-full bg-primary-100">
                  <BanknotesIcon v-if="container.type === 'cash'" class="h-6 w-6 text-primary-600" />
                  <BuildingLibraryIcon v-else class="h-6 w-6 text-primary-600" />
                </span>
              </div>
              <div class="mr-4 truncate">
                <div class="flex items-center">
                  <p class="text-sm font-medium text-primary-600 truncate">{{ container.name }}</p>
                  <span
                    :class="[
                      container.is_active ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800',
                      'mr-2 px-2 inline-flex text-xs leading-5 font-semibold rounded-full'
                    ]"
                  >
                    {{ container.is_active ? 'نشط' : 'غير نشط' }}
                  </span>
                </div>
                <div class="flex items-center text-sm text-gray-500 mt-1">
                  <span class="truncate">{{ container.type === 'cash' ? 'صندوق نقدي' : 'حساب مصرفي' }}</span>
                  <span class="mx-2">•</span>
                  <span class="truncate">{{ container.currency }}</span>
                  <span class="mx-2">•</span>
                  <span class="font-medium text-gray-900">{{ formatCurrency(container.balance, container.currency) }}</span>
                </div>
                <div v-if="container.type === 'bank'" class="text-xs text-gray-400 mt-1">
                  {{ container.bank_name }} - {{ container.account_number }}
                </div>
              </div>
            </div>
            <div class="flex items-center space-x-2 space-x-reverse">
              <button
                @click="openEditModal(container)"
                class="p-2 text-gray-400 hover:text-gray-500"
                title="تعديل"
              >
                <PencilIcon class="h-5 w-5" aria-hidden="true" />
              </button>
              <template v-if="authStore.canDelete">
                <button
                  @click="toggleActive(container)"
                  :class="[
                    container.is_active ? 'text-yellow-500 hover:text-yellow-600' : 'text-green-500 hover:text-green-600',
                    'p-2'
                  ]"
                  :title="container.is_active ? 'تعطيل' : 'تنشيط'"
                >
                  <NoSymbolIcon v-if="container.is_active" class="h-5 w-5" aria-hidden="true" />
                  <CheckCircleIcon v-else class="h-5 w-5" aria-hidden="true" />
                </button>
                <button
                  @click="openDeleteModal(container)"
                  class="p-2 text-red-400 hover:text-red-500"
                  title="حذف نهائي"
                >
                  <TrashIcon class="h-5 w-5" aria-hidden="true" />
                </button>
              </template>
            </div>
          </div>
        </li>
      </ul>
      <div v-if="!containers || containers.length === 0" class="text-center py-12">
        <BuildingLibraryIcon class="mx-auto h-12 w-12 text-gray-400" />
        <h3 class="mt-2 text-sm font-medium text-gray-900">لا يوجد حسابات</h3>
        <p class="mt-1 text-sm text-gray-500">ابدأ بإضافة حساب مصرفي أو صندوق نقدي جديد.</p>
      </div>
    </div>

    <!-- Create/Edit Modal (Standard HTML/CSS) -->
    <div v-if="isModalOpen" class="fixed inset-0 z-50 overflow-y-auto" aria-labelledby="modal-title" role="dialog" aria-modal="true">
      <div class="flex items-end justify-center min-h-screen pt-4 px-4 pb-20 text-center sm:block sm:p-0">
        <!-- Background overlay -->
        <div class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" aria-hidden="true" @click="closeModal"></div>

        <!-- Modal panel -->
        <span class="hidden sm:inline-block sm:align-middle sm:h-screen" aria-hidden="true">&#8203;</span>
        <div class="inline-block align-bottom bg-white rounded-lg px-4 pt-5 pb-4 text-right overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-lg sm:w-full sm:p-6">
          <div class="hidden sm:block absolute top-0 left-0 pt-4 pl-4">
            <button type="button" class="bg-white rounded-md text-gray-400 hover:text-gray-500 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500" @click="closeModal">
              <span class="sr-only">إغلاق</span>
              <XMarkIcon class="h-6 w-6" aria-hidden="true" />
            </button>
          </div>
          <div class="sm:flex sm:items-start">
            <div class="mt-3 text-center sm:mt-0 sm:mr-4 sm:text-right w-full">
              <h3 class="text-lg leading-6 font-medium text-gray-900" id="modal-title">
                {{ isEditing ? 'تعديل حساب' : 'حساب جديد' }}
              </h3>
              <div class="mt-4 space-y-4">
                <!-- Name -->
                <div>
                  <label for="name" class="block text-sm font-medium text-gray-700">اسم الحساب (English)</label>
                  <div class="mt-1">
                    <input
                      type="text"
                      name="name"
                      id="name"
                      v-model="form.name"
                      class="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
                    />
                  </div>
                </div>

                <!-- Name AR -->
                <div>
                  <label for="name_ar" class="block text-sm font-medium text-gray-700">اسم الحساب (عربي)</label>
                  <div class="mt-1">
                    <input
                      type="text"
                      name="name_ar"
                      id="name_ar"
                      v-model="form.name_ar"
                      class="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
                    />
                  </div>
                </div>

                <!-- Type -->
                <div>
                  <label for="type" class="block text-sm font-medium text-gray-700">النوع</label>
                  <div class="mt-1">
                    <select
                      id="type"
                      name="type"
                      v-model="form.type"
                      class="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
                    >
                      <option value="cash">صندوق نقدي</option>
                      <option value="bank">حساب مصرفي</option>
                    </select>
                  </div>
                </div>

                <!-- Currency -->
                <div>
                  <label for="currency" class="block text-sm font-medium text-gray-700">العملة</label>
                  <div class="mt-1">
                    <select
                      id="currency"
                      name="currency"
                      v-model="form.currency"
                      class="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
                    >
                      <option value="LYD">دينار ليبي (LYD)</option>
                      <option value="USD">دولار أمريكي (USD)</option>
                      <option value="EUR">يورو (EUR)</option>
                      <option value="GBP">جنيه إسترليني (GBP)</option>
                    </select>
                  </div>
                </div>

                <!-- Bank Details (Only if Type is Bank) -->
                <div v-if="form.type === 'bank'" class="space-y-4 border-t border-gray-200 pt-4 mt-4">
                  <h4 class="text-sm font-medium text-gray-900">بيانات البنك</h4>
                  
                  <div>
                    <label for="bank_name" class="block text-sm font-medium text-gray-700">اسم البنك</label>
                    <div class="mt-1">
                      <input
                        type="text"
                        name="bank_name"
                        id="bank_name"
                        v-model="form.bank_name"
                        class="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
                      />
                    </div>
                  </div>

                  <div>
                    <label for="account_number" class="block text-sm font-medium text-gray-700">رقم الحساب</label>
                    <div class="mt-1">
                      <input
                        type="text"
                        name="account_number"
                        id="account_number"
                        v-model="form.account_number"
                        class="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
                      />
                    </div>
                  </div>

                  <div>
                    <label for="iban" class="block text-sm font-medium text-gray-700">IBAN</label>
                    <div class="mt-1">
                      <input
                        type="text"
                        name="iban"
                        id="iban"
                        v-model="form.iban"
                        class="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
                      />
                    </div>
                  </div>

                  <div>
                    <label for="swift_code" class="block text-sm font-medium text-gray-700">Swift Code</label>
                    <div class="mt-1">
                      <input
                        type="text"
                        name="swift_code"
                        id="swift_code"
                        v-model="form.swift_code"
                        class="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
                      />
                    </div>
                  </div>
                </div>

                <!-- Active Status (Edit Only) -->
                <div v-if="isEditing" class="flex items-center">
                  <input
                    id="is_active"
                    name="is_active"
                    type="checkbox"
                    v-model="form.is_active"
                    class="h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500"
                  />
                  <label for="is_active" class="mr-2 block text-sm text-gray-900">حساب نشط</label>
                </div>
              </div>
            </div>
          </div>
          <div class="mt-5 sm:mt-4 sm:flex sm:flex-row-reverse">
            <button
              type="button"
              class="inline-flex w-full justify-center rounded-md border border-transparent bg-primary-600 px-4 py-2 text-base font-medium text-white shadow-sm hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500 sm:ml-3 sm:w-auto sm:text-sm"
              @click="saveContainer"
              :disabled="loading"
            >
              {{ loading ? 'جاري الحفظ...' : 'حفظ' }}
            </button>
            <button
              type="button"
              class="mt-3 inline-flex w-full justify-center rounded-md border border-gray-300 shadow-sm px-4 py-2 bg-white text-base font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500 sm:mt-0 sm:w-auto sm:text-sm"
              @click="closeModal"
            >
              إلغاء
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Delete Confirmation Modal -->
    <ConfirmModal
      v-model="showDeleteModal"
      :title="'حذف الحساب'"
      :message="deleteModalMessage"
      :danger="true"
      :requireConfirmation="true"
      :confirmationText="'أؤكد أنني أريد حذف هذا الحساب وجميع معاملاته نهائياً'"
      :confirmText="'حذف نهائي'"
      :loading="deleteLoading"
      @confirm="executeDelete"
    />

    <!-- Toast Notification -->
    <ToastNotification
      v-model="showToast"
      :type="toastType"
      :message="toastMessage"
      :duration="4000"
    />
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { PlusIcon, PencilIcon, TrashIcon, XMarkIcon, BanknotesIcon, BuildingLibraryIcon, NoSymbolIcon, CheckCircleIcon } from '@heroicons/vue/24/outline'
import { containers as containersApi } from '@/api'
import { useAuthStore } from '@/stores/auth'
import ConfirmModal from '@/components/ConfirmModal.vue'
import ToastNotification from '@/components/ToastNotification.vue'

const authStore = useAuthStore()
const containers = ref([])
const isModalOpen = ref(false)
const isEditing = ref(false)
const loading = ref(false)

// Delete modal state
const showDeleteModal = ref(false)
const deleteLoading = ref(false)
const containerToDelete = ref(null)
const deleteModalMessage = computed(() => {
  if (!containerToDelete.value) return ''
  return `سيتم حذف الحساب "${containerToDelete.value.name}" مع جميع المعاملات المالية المرتبطة به بشكل نهائي.\n\n⚠️ لا يمكن التراجع عن هذا الإجراء!`
})

// Toast state
const showToast = ref(false)
const toastType = ref('success')
const toastMessage = ref('')

const showSuccessToast = (message) => {
  toastType.value = 'success'
  toastMessage.value = message
  showToast.value = true
}

const showErrorToast = (message) => {
  toastType.value = 'error'
  toastMessage.value = message
  showToast.value = true
}

const form = ref({
  id: null,
  name: '',
  name_ar: '',
  type: 'bank',
  currency: 'LYD',
  balance: 0,
  bank_name: '',
  account_number: '',
  iban: '',
  swift_code: '',
  is_active: true
})

const fetchContainers = async () => {
  try {
    const response = await containersApi.list()
    containers.value = response.data
  } catch (error) {
    console.error('Failed to fetch containers:', error)
    showErrorToast('فشل في تحميل الحسابات')
  }
}

const openCreateModal = () => {
  isEditing.value = false
  form.value = {
    id: null,
    name: '',
    name_ar: '',
    type: 'bank',
    currency: 'LYD',
    balance: 0,
    bank_name: '',
    account_number: '',
    iban: '',
    swift_code: '',
    is_active: true
  }
  isModalOpen.value = true
}

const openEditModal = (container) => {
  isEditing.value = true
  form.value = {
    id: container.id,
    name: container.name,
    name_ar: container.name_ar,
    type: container.type,
    currency: container.currency,
    balance: container.balance,
    bank_name: container.bank_name || '',
    account_number: container.account_number || '',
    iban: container.iban || '',
    swift_code: container.swift_code || '',
    is_active: container.is_active
  }
  isModalOpen.value = true
}

const closeModal = () => {
  isModalOpen.value = false
}

const saveContainer = async () => {
  loading.value = true
  try {
    if (!form.value.name || !form.value.currency || !form.value.type) {
      showErrorToast('يرجى إدخال اسم الحساب والعملة والنوع')
      return
    }
    if (form.value.type === 'bank') {
      if (!form.value.bank_name || !form.value.account_number) {
        showErrorToast('يرجى إدخال اسم البنك ورقم الحساب للحسابات المصرفية')
        return
      }
    }

    const payload = { ...form.value }

    if (isEditing.value) {
      await containersApi.update(form.value.id, payload)
    } else {
      delete payload.id
      await containersApi.create(payload)
    }
    await fetchContainers()
    closeModal()
    showSuccessToast(isEditing.value ? 'تم تحديث الحساب بنجاح' : 'تم إنشاء الحساب بنجاح')
  } catch (error) {
    console.error('Failed to save container:', error)
    const errorMessage = error.response?.data?.error || 'حدث خطأ أثناء حفظ البيانات'
    showErrorToast(errorMessage)
  } finally {
    loading.value = false
  }
}

const toggleActive = async (container) => {
  try {
    await containersApi.update(container.id, {
      ...container,
      is_active: !container.is_active
    })
    await fetchContainers()
    const action = container.is_active ? 'تم تعطيل' : 'تم تنشيط'
    showSuccessToast(`${action} الحساب بنجاح`)
  } catch (error) {
    console.error('Failed to toggle container status:', error)
    showErrorToast('حدث خطأ أثناء تغيير حالة الحساب')
  }
}

const openDeleteModal = (container) => {
  containerToDelete.value = container
  showDeleteModal.value = true
}

const executeDelete = async () => {
  if (!containerToDelete.value) return
  
  deleteLoading.value = true
  try {
    await containersApi.delete(containerToDelete.value.id)
    await fetchContainers()
    showDeleteModal.value = false
    showSuccessToast('تم حذف الحساب وجميع معاملاته بنجاح')
  } catch (error) {
    console.error('Failed to delete container:', error)
    const errorMessage = error.response?.data?.error || 'حدث خطأ أثناء حذف الحساب'
    let arabicMessage = errorMessage
    if (errorMessage.includes('Container not found')) {
      arabicMessage = 'الحساب غير موجود'
    }
    showErrorToast(arabicMessage)
  } finally {
    deleteLoading.value = false
    containerToDelete.value = null
  }
}

const formatCurrency = (amount, currency) => {
  const code = currency || 'LYD'
  let locale = 'ar-LY'
  
  // Use specific locales for foreign currencies to ensure correct symbols
  if (code === 'USD') locale = 'en-US'
  if (code === 'EUR') locale = 'en-IE'
  if (code === 'GBP') locale = 'en-GB'

  return new Intl.NumberFormat(locale, {
    style: 'currency',
    currency: code
  }).format(amount)
}

onMounted(() => {
  fetchContainers()
})
</script>
