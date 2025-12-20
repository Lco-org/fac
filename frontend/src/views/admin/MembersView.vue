<template>
  <div class="space-y-6">
    <div class="flex justify-between items-center">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">إدارة الأعضاء المنتسبين</h1>
        <p class="mt-1 text-sm text-gray-500">إدارة الأعضاء المنتسبين للمنظمة</p>
      </div>
      <button @click="openCreateModal" class="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-primary-600 hover:bg-primary-700">
        <PlusIcon class="-mr-1 ml-2 h-5 w-5" />
        عضو جديد
      </button>
    </div>

    <div class="bg-white shadow overflow-hidden sm:rounded-lg">
      <ul role="list" class="divide-y divide-gray-200">
        <li v-for="member in members" :key="member.id" class="px-4 py-4 sm:px-6 hover:bg-gray-50">
          <div class="flex items-center justify-between">
            <div class="flex-1">
              <div class="flex items-center">
                <p class="text-sm font-medium text-primary-600">{{ member.full_name }}</p>
                <span :class="[member.is_active ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800', 'mr-2 px-2 inline-flex text-xs leading-5 font-semibold rounded-full']">
                  {{ member.is_active ? 'نشط' : 'غير نشط' }}
                </span>
              </div>
              <div class="mt-1 flex items-center text-sm text-gray-500">
                <span>رقم العضوية: {{ member.membership_number }}</span>
                <span class="mx-2">•</span>
                <span>{{ member.member_type }}</span>
                <span v-if="member.phone" class="mx-2">•</span>
                <span v-if="member.phone">{{ member.phone }}</span>
              </div>
            </div>
            <div class="flex items-center space-x-2 space-x-reverse">
              <button @click="openEditModal(member)" class="p-2 text-gray-400 hover:text-gray-500">
                <PencilIcon class="h-5 w-5" />
              </button>
              <button @click="openDeleteModal(member)" class="p-2 text-red-400 hover:text-red-500">
                <TrashIcon class="h-5 w-5" />
              </button>
            </div>
          </div>
        </li>
      </ul>
      <div v-if="members.length === 0" class="text-center py-12">
        <h3 class="mt-2 text-sm font-medium text-gray-900">لا يوجد أعضاء</h3>
        <p class="mt-1 text-sm text-gray-500">ابدأ بإضافة عضو جديد.</p>
      </div>
    </div>

    <!-- Modal -->
    <TransitionRoot as="template" :show="isModalOpen">
      <Dialog as="div" class="relative z-10" @close="closeModal">
        <TransitionChild as="template" enter="ease-out duration-300" enter-from="opacity-0" enter-to="opacity-100" leave="ease-in duration-200" leave-from="opacity-100" leave-to="opacity-0">
          <div class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" />
        </TransitionChild>
        <div class="fixed inset-0 z-10 overflow-y-auto">
          <div class="flex min-h-full items-end justify-center p-4 text-center sm:items-center sm:p-0">
            <TransitionChild as="template" enter="ease-out duration-300" enter-from="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95" enter-to="opacity-100 translate-y-0 sm:scale-100" leave="ease-in duration-200" leave-from="opacity-100 translate-y-0 sm:scale-100" leave-to="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95">
              <DialogPanel class="relative transform overflow-hidden rounded-lg bg-white px-4 pt-5 pb-4 text-right shadow-xl transition-all sm:my-8 sm:w-full sm:max-w-lg sm:p-6">
                <DialogTitle as="h3" class="text-lg font-medium leading-6 text-gray-900">{{ isEditing ? 'تعديل عضو' : 'عضو جديد' }}</DialogTitle>
                <div class="mt-4 space-y-4">
                  <div>
                    <label class="block text-sm font-medium text-gray-700">رقم العضوية</label>
                    <input type="text" v-model="form.membership_number" :disabled="isEditing" class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm" />
                  </div>
                  <div>
                    <label class="block text-sm font-medium text-gray-700">الاسم الكامل</label>
                    <input type="text" v-model="form.full_name" class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm" />
                  </div>
                  <div>
                    <label class="block text-sm font-medium text-gray-700">نوع العضوية</label>
                    <select v-model="form.member_type" class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm">
                      <option value="regular">عادي</option>
                      <option value="honorary">فخري</option>
                      <option value="founding">مؤسس</option>
                    </select>
                  </div>
                  <div>
                    <label class="block text-sm font-medium text-gray-700">الهاتف</label>
                    <input type="text" v-model="form.phone" class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm" />
                  </div>
                  <div>
                    <label class="block text-sm font-medium text-gray-700">البريد الإلكتروني</label>
                    <input type="email" v-model="form.email" class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm" />
                  </div>
                </div>
                <div class="mt-5 sm:mt-4 sm:flex sm:flex-row-reverse">
                  <button type="button" @click="saveMember" class="inline-flex w-full justify-center rounded-md border border-transparent bg-primary-600 px-4 py-2 text-base font-medium text-white shadow-sm hover:bg-primary-700 sm:ml-3 sm:w-auto sm:text-sm">حفظ</button>
                  <button type="button" @click="closeModal" class="mt-3 inline-flex w-full justify-center rounded-md border border-gray-300 bg-white px-4 py-2 text-base font-medium text-gray-700 shadow-sm hover:bg-gray-50 sm:mt-0 sm:w-auto sm:text-sm">إلغاء</button>
                </div>
              </DialogPanel>
            </TransitionChild>
          </div>
        </div>
      </Dialog>
    </TransitionRoot>

    <!-- Delete Confirmation Modal -->
    <ConfirmModal
      v-model="showDeleteModal"
      title="حذف عضو"
      :message="deleteModalMessage"
      :danger="true"
      :requireConfirmation="true"
      confirmationText="أؤكد أنني أريد حذف هذا العضو"
      confirmText="حذف"
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
import { Dialog, DialogPanel, DialogTitle, TransitionChild, TransitionRoot } from '@headlessui/vue'
import { PlusIcon, PencilIcon, TrashIcon } from '@heroicons/vue/24/outline'
import { members as membersApi } from '@/api'
import ConfirmModal from '@/components/ConfirmModal.vue'
import ToastNotification from '@/components/ToastNotification.vue'

const members = ref([])
const isModalOpen = ref(false)
const isEditing = ref(false)

// Delete modal state
const showDeleteModal = ref(false)
const deleteLoading = ref(false)
const memberToDelete = ref(null)
const deleteModalMessage = computed(() => {
  if (!memberToDelete.value) return ''
  return `سيتم حذف العضو "${memberToDelete.value.full_name}" بشكل نهائي.`
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
  membership_number: '',
  full_name: '',
  member_type: 'regular',
  phone: '',
  email: ''
})

const fetchMembers = async () => {
  try {
    const response = await membersApi.list()
    members.value = response.data
  } catch (error) {
    console.error('Failed to fetch members:', error)
    showErrorToast('فشل في تحميل الأعضاء')
  }
}

const openCreateModal = () => {
  isEditing.value = false
  form.value = { id: null, membership_number: '', full_name: '', member_type: 'regular', phone: '', email: '' }
  isModalOpen.value = true
}

const openEditModal = (member) => {
  isEditing.value = true
  form.value = { ...member }
  isModalOpen.value = true
}

const closeModal = () => {
  isModalOpen.value = false
}

const saveMember = async () => {
  try {
    if (isEditing.value) {
      await membersApi.update(form.value.id, form.value)
    } else {
      await membersApi.create(form.value)
    }
    await fetchMembers()
    closeModal()
    showSuccessToast(isEditing.value ? 'تم تحديث بيانات العضو بنجاح' : 'تم إضافة العضو بنجاح')
  } catch (error) {
    console.error('Failed to save member:', error)
    showErrorToast(error.response?.data?.error || 'حدث خطأ أثناء حفظ البيانات')
  }
}

const openDeleteModal = (member) => {
  memberToDelete.value = member
  showDeleteModal.value = true
}

const executeDelete = async () => {
  if (!memberToDelete.value) return
  
  deleteLoading.value = true
  try {
    await membersApi.delete(memberToDelete.value.id)
    await fetchMembers()
    showDeleteModal.value = false
    showSuccessToast('تم حذف العضو بنجاح')
  } catch (error) {
    console.error('Failed to delete member:', error)
    showErrorToast(error.response?.data?.error || 'حدث خطأ أثناء حذف العضو')
  } finally {
    deleteLoading.value = false
    memberToDelete.value = null
  }
}

onMounted(() => {
  fetchMembers()
})
</script>
