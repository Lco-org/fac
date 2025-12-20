<template>
  <div class="space-y-6">
    <div class="flex justify-between items-center">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">إدارة المشاريع</h1>
        <p class="mt-1 text-sm text-gray-500">إدارة المشاريع والميزانيات</p>
      </div>
      <button @click="openCreateModal" class="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-primary-600 hover:bg-primary-700">
        <PlusIcon class="-mr-1 ml-2 h-5 w-5" />
        مشروع جديد
      </button>
    </div>

    <div class="bg-white shadow overflow-hidden sm:rounded-lg">
      <ul role="list" class="divide-y divide-gray-200">
        <li v-for="project in projects" :key="project.id" class="px-4 py-4 sm:px-6 hover:bg-gray-50">
          <div class="flex items-center justify-between">
            <div class="flex-1">
              <div class="flex items-center">
                <p class="text-sm font-medium text-primary-600">{{ project.name_ar || project.name }}</p>
                <span :class="[project.is_active ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800', 'mr-2 px-2 inline-flex text-xs leading-5 font-semibold rounded-full']">
                  {{ project.is_active ? 'نشط' : 'غير نشط' }}
                </span>
                <span v-if="project.is_locked" class="mr-2 px-2 inline-flex text-xs leading-5 font-semibold rounded-full bg-yellow-100 text-yellow-800">
                  🔒 مقفل
                </span>
              </div>
              <div class="mt-1 flex items-center text-sm text-gray-500">
                <span>كود: {{ project.code }}</span>
                <span class="mx-2">•</span>
                <span v-if="project.budget">الميزانية: {{ project.budget }} {{ project.currency }}</span>
                <span v-if="project.start_date" class="mx-2">•</span>
                <span v-if="project.start_date">بدء: {{ formatDate(project.start_date) }}</span>
              </div>
              <div v-if="project.description" class="mt-1 text-xs text-gray-400 line-clamp-1">{{ project.description }}</div>
              <div v-if="project.is_locked && project.lock_reason" class="mt-1 text-xs text-yellow-600">سبب القفل: {{ project.lock_reason }}</div>
            </div>
            <div class="flex items-center space-x-2 space-x-reverse">
              <button v-if="!project.is_locked" @click="openLockModal(project)" class="p-2 text-yellow-500 hover:text-yellow-600" title="قفل وتحويل">
                <LockClosedIcon class="h-5 w-5" />
              </button>
              <button @click="openEditModal(project)" class="p-2 text-gray-400 hover:text-gray-500">
                <PencilIcon class="h-5 w-5" />
              </button>
              <button @click="openDeleteModal(project)" class="p-2 text-red-400 hover:text-red-500">
                <TrashIcon class="h-5 w-5" />
              </button>
            </div>
          </div>
        </li>
      </ul>
      <div v-if="projects.length === 0" class="text-center py-12">
        <h3 class="mt-2 text-sm font-medium text-gray-900">لا يوجد مشاريع</h3>
        <p class="mt-1 text-sm text-gray-500">ابدأ بإضافة مشروع جديد.</p>
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
                <DialogTitle as="h3" class="text-lg font-medium leading-6 text-gray-900">{{ isEditing ? 'تعديل مشروع' : 'مشروع جديد' }}</DialogTitle>
                <div class="mt-4 space-y-4">
                  <div>
                    <label class="block text-sm font-medium text-gray-700">كود المشروع</label>
                    <input type="text" v-model="form.code" :disabled="isEditing" class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm" />
                  </div>
                  <div>
                    <label class="block text-sm font-medium text-gray-700">اسم المشروع (English)</label>
                    <input type="text" v-model="form.name" class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm" />
                  </div>
                  <div>
                    <label class="block text-sm font-medium text-gray-700">اسم المشروع (عربي)</label>
                    <input type="text" v-model="form.name_ar" class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm" />
                  </div>
                  <div>
                    <label class="block text-sm font-medium text-gray-700">وصف المشروع</label>
                    <textarea v-model="form.description" rows="2" class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm" placeholder="وصف مختصر للمشروع..."></textarea>
                  </div>
                  <div class="grid grid-cols-2 gap-4">
                    <div>
                      <label class="block text-sm font-medium text-gray-700">الميزانية</label>
                      <input type="number" v-model="form.budget" class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm" />
                    </div>
                    <div>
                      <label class="block text-sm font-medium text-gray-700">العملة</label>
                      <select v-model="form.currency" class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm">
                        <option value="LYD">دينار ليبي (LYD)</option>
                        <option value="USD">دولار أمريكي (USD)</option>
                        <option value="EUR">يورو (EUR)</option>
                      </select>
                    </div>
                  </div>
                  <div>
                    <label class="block text-sm font-medium text-gray-700">تاريخ بدء المشروع</label>
                    <input type="date" v-model="form.start_date" class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm" />
                  </div>
                </div>
                <div class="mt-5 sm:mt-4 sm:flex sm:flex-row-reverse">
                  <button type="button" @click="saveProject" class="inline-flex w-full justify-center rounded-md border border-transparent bg-primary-600 px-4 py-2 text-base font-medium text-white shadow-sm hover:bg-primary-700 sm:ml-3 sm:w-auto sm:text-sm">حفظ</button>
                  <button type="button" @click="closeModal" class="mt-3 inline-flex w-full justify-center rounded-md border border-gray-300 bg-white px-4 py-2 text-base font-medium text-gray-700 shadow-sm hover:bg-gray-50 sm:mt-0 sm:w-auto sm:text-sm">إلغاء</button>
                </div>
              </DialogPanel>
            </TransitionChild>
          </div>
        </div>
      </Dialog>
    </TransitionRoot>

    <!-- Lock Modal -->
    <TransitionRoot as="template" :show="isLockModalOpen">
      <Dialog as="div" class="relative z-10" @close="isLockModalOpen = false">
        <TransitionChild as="template" enter="ease-out duration-300" enter-from="opacity-0" enter-to="opacity-100" leave="ease-in duration-200" leave-from="opacity-100" leave-to="opacity-0">
          <div class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" />
        </TransitionChild>
        <div class="fixed inset-0 z-10 overflow-y-auto">
          <div class="flex min-h-full items-end justify-center p-4 text-center sm:items-center sm:p-0">
            <TransitionChild as="template" enter="ease-out duration-300" enter-from="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95" enter-to="opacity-100 translate-y-0 sm:scale-100" leave="ease-in duration-200" leave-from="opacity-100 translate-y-0 sm:scale-100" leave-to="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95">
              <DialogPanel class="relative transform overflow-hidden rounded-lg bg-white px-4 pt-5 pb-4 text-right shadow-xl transition-all sm:my-8 sm:w-full sm:max-w-lg sm:p-6">
                <DialogTitle as="h3" class="text-lg font-medium leading-6 text-gray-900">🔒 قفل المشروع وتحويل الرصيد</DialogTitle>
                <div class="mt-4 space-y-4">
                  <div class="bg-yellow-50 border border-yellow-200 rounded-lg p-3">
                    <p class="text-sm text-yellow-800">☠️ <strong>تحذير:</strong> سيتم قفل المشروع نهائياً ولن يمكن إجراء أي عمليات عليه.</p>
                  </div>
                  <div>
                    <label class="block text-sm font-medium text-gray-700">سبب القفل <span class="text-red-500">*</span></label>
                    <textarea v-model="lockForm.reason" rows="2" required class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm" placeholder="مثال: انتهاء المشروع..."></textarea>
                  </div>
                  <div>
                    <label class="block text-sm font-medium text-gray-700">تحويل الرصيد المتبقي إلى</label>
                    <select v-model="lockForm.transfer_type" class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm">
                      <option value="none">لا تحويل (إبقاء الرصيد)</option>
                      <option value="project">مشروع آخر</option>
                      <option value="container">حساب مصرفي / نقدي</option>
                    </select>
                  </div>
                  <div v-if="lockForm.transfer_type === 'project'">
                    <label class="block text-sm font-medium text-gray-700">المشروع المستهدف</label>
                    <select v-model="lockForm.target_project_id" class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm">
                      <option value="">اختر المشروع...</option>
                      <option v-for="p in projects.filter(x => x.id !== lockForm.project_id && !x.is_locked)" :key="p.id" :value="p.id">{{ p.code }} - {{ p.name_ar || p.name }}</option>
                    </select>
                  </div>
                  <div v-if="lockForm.transfer_type === 'container'">
                    <label class="block text-sm font-medium text-gray-700">الحساب المستهدف</label>
                    <select v-model="lockForm.target_container_id" class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm">
                      <option value="">اختر الحساب...</option>
                      <option v-for="c in containers" :key="c.id" :value="c.id">{{ c.name }} ({{ c.currency }})</option>
                    </select>
                  </div>
                </div>
                <div class="mt-5 sm:mt-4 sm:flex sm:flex-row-reverse">
                  <button type="button" @click="lockProject" :disabled="!lockForm.reason" class="inline-flex w-full justify-center rounded-md border border-transparent bg-yellow-600 px-4 py-2 text-base font-medium text-white shadow-sm hover:bg-yellow-700 disabled:opacity-50 sm:ml-3 sm:w-auto sm:text-sm">🔒 قفل المشروع</button>
                  <button type="button" @click="isLockModalOpen = false" class="mt-3 inline-flex w-full justify-center rounded-md border border-gray-300 bg-white px-4 py-2 text-base font-medium text-gray-700 shadow-sm hover:bg-gray-50 sm:mt-0 sm:w-auto sm:text-sm">إلغاء</button>
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
      title="حذف مشروع"
      :message="deleteModalMessage"
      :danger="true"
      :requireConfirmation="true"
      confirmationText="أؤكد أنني أريد حذف هذا المشروع وجميع معاملاته"
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
import { PlusIcon, PencilIcon, TrashIcon, LockClosedIcon } from '@heroicons/vue/24/outline'
import { projects as projectsApi, containers as containersApi } from '@/api'
import ConfirmModal from '@/components/ConfirmModal.vue'
import ToastNotification from '@/components/ToastNotification.vue'

const projects = ref([])
const containers = ref([])
const isModalOpen = ref(false)
const isLockModalOpen = ref(false)
const isEditing = ref(false)

// Delete modal state
const showDeleteModal = ref(false)
const deleteLoading = ref(false)
const projectToDelete = ref(null)
const deleteModalMessage = computed(() => {
  if (!projectToDelete.value) return ''
  return `سيتم حذف المشروع "${projectToDelete.value.name_ar || projectToDelete.value.name}" وجميع المعاملات المرتبطة به بشكل نهائي.`
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
  code: '',
  name: '',
  name_ar: '',
  description: '',
  start_date: '',
  budget: null,
  currency: 'LYD'
})

const formatDate = (dateString) => {
  if (!dateString) return ''
  return new Date(dateString).toLocaleDateString('en-GB')
}

const fetchProjects = async () => {
  try {
    const response = await projectsApi.list()
    projects.value = response.data
  } catch (error) {
    console.error('Failed to fetch projects:', error)
    showErrorToast('فشل في تحميل المشاريع')
  }
}

const openCreateModal = () => {
  isEditing.value = false
  form.value = { id: null, code: '', name: '', name_ar: '', description: '', start_date: '', budget: null, currency: 'LYD' }
  isModalOpen.value = true
}

const openEditModal = (project) => {
  isEditing.value = true
  form.value = { ...project }
  isModalOpen.value = true
}

const closeModal = () => {
  isModalOpen.value = false
}

const saveProject = async () => {
  try {
    // Prepare payload - convert empty strings to null for optional fields
    const payload = {
      ...form.value,
      name_ar: form.value.name_ar || null,
      description: form.value.description || null,
      start_date: form.value.start_date ? new Date(form.value.start_date).toISOString() : null,
      budget: form.value.budget || null
    }
    
    if (isEditing.value) {
      await projectsApi.update(form.value.id, payload)
    } else {
      await projectsApi.create(payload)
    }
    await fetchProjects()
    closeModal()
    showSuccessToast(isEditing.value ? 'تم تحديث المشروع بنجاح' : 'تم إنشاء المشروع بنجاح')
  } catch (error) {
    console.error('Failed to save project:', error)
    showErrorToast(error.response?.data?.error || 'حدث خطأ أثناء حفظ البيانات')
  }
}

const openDeleteModal = (project) => {
  projectToDelete.value = project
  showDeleteModal.value = true
}

const executeDelete = async () => {
  if (!projectToDelete.value) return
  
  deleteLoading.value = true
  try {
    await projectsApi.delete(projectToDelete.value.id)
    await fetchProjects()
    showDeleteModal.value = false
    showSuccessToast('تم حذف المشروع بنجاح')
  } catch (error) {
    console.error('Failed to delete project:', error)
    showErrorToast(error.response?.data?.error || 'حدث خطأ أثناء حذف المشروع')
  } finally {
    deleteLoading.value = false
    projectToDelete.value = null
  }
}

// Lock functionality
const lockForm = ref({
  project_id: '',
  reason: '',
  transfer_type: 'none',
  target_project_id: '',
  target_container_id: ''
})

const fetchContainers = async () => {
  try {
    const response = await containersApi.list()
    containers.value = response.data
  } catch (error) {
    console.error('Failed to fetch containers:', error)
  }
}

const openLockModal = (project) => {
  lockForm.value = {
    project_id: project.id,
    reason: '',
    transfer_type: 'none',
    target_project_id: '',
    target_container_id: ''
  }
  isLockModalOpen.value = true
}

const lockProject = async () => {
  try {
    // For now, just update the project as locked (Backend API to be implemented)
    await projectsApi.update(lockForm.value.project_id, {
      is_locked: true,
      lock_reason: lockForm.value.reason
    })
    
    // If transfer is selected, create a transfer transaction
    if (lockForm.value.transfer_type !== 'none') {
      showErrorToast('ملاحظة: ميزة تحويل الرصيد التلقائي قيد التطوير. يرجى إنشاء عملية تحويل يدوياً.')
    }
    
    await fetchProjects()
    isLockModalOpen.value = false
    showSuccessToast('تم قفل المشروع بنجاح')
  } catch (error) {
    console.error('Failed to lock project:', error)
    showErrorToast(error.response?.data?.error || 'حدث خطأ أثناء قفل المشروع')
  }
}

onMounted(() => {
  fetchProjects()
  fetchContainers()
})
</script>
