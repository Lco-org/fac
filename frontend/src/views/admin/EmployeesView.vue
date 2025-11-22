<template>
  <div class="space-y-6">
    <div class="flex justify-between items-center">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">إدارة الموظفين</h1>
        <p class="mt-1 text-sm text-gray-500">إدارة موظفي المنظمة</p>
      </div>
      <button @click="openCreateModal" class="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-primary-600 hover:bg-primary-700">
        <PlusIcon class="-mr-1 ml-2 h-5 w-5" />
        موظف جديد
      </button>
    </div>

    <div class="bg-white shadow overflow-hidden sm:rounded-lg">
      <ul role="list" class="divide-y divide-gray-200">
        <li v-for="employee in employees" :key="employee.id" class="px-4 py-4 sm:px-6 hover:bg-gray-50">
          <div class="flex items-center justify-between">
            <div class="flex-1">
              <div class="flex items-center">
                <p class="text-sm font-medium text-primary-600">{{ employee.full_name }}</p>
                <span :class="[employee.is_active ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800', 'mr-2 px-2 inline-flex text-xs leading-5 font-semibold rounded-full']">
                  {{ employee.is_active ? 'نشط' : 'غير نشط' }}
                </span>
              </div>
              <div class="mt-1 flex items-center text-sm text-gray-500">
                <span>رقم: {{ employee.employee_number }}</span>
                <span class="mx-2">•</span>
                <span v-if="employee.position">{{ employee.position }}</span>
                <span v-if="employee.department" class="mx-2">•</span>
                <span v-if="employee.department">{{ employee.department }}</span>
              </div>
            </div>
            <div class="flex items-center space-x-2 space-x-reverse">
              <button @click="openEditModal(employee)" class="p-2 text-gray-400 hover:text-gray-500">
                <PencilIcon class="h-5 w-5" />
              </button>
              <button @click="confirmDelete(employee)" class="p-2 text-red-400 hover:text-red-500">
                <TrashIcon class="h-5 w-5" />
              </button>
            </div>
          </div>
        </li>
      </ul>
      <div v-if="employees.length === 0" class="text-center py-12">
        <h3 class="mt-2 text-sm font-medium text-gray-900">لا يوجد موظفين</h3>
        <p class="mt-1 text-sm text-gray-500">ابدأ بإضافة موظف جديد.</p>
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
                <DialogTitle as="h3" class="text-lg font-medium leading-6 text-gray-900">{{ isEditing ? 'تعديل موظف' : 'موظف جديد' }}</DialogTitle>
                <div class="mt-4 space-y-4">
                  <div>
                    <label class="block text-sm font-medium text-gray-700">رقم الموظف</label>
                    <input type="text" v-model="form.employee_number" :disabled="isEditing" class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm" />
                  </div>
                  <div>
                    <label class="block text-sm font-medium text-gray-700">الاسم الكامل</label>
                    <input type="text" v-model="form.full_name" class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm" />
                  </div>
                  <div>
                    <label class="block text-sm font-medium text-gray-700">المنصب</label>
                    <input type="text" v-model="form.position" class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm" />
                  </div>
                  <div>
                    <label class="block text-sm font-medium text-gray-700">القسم</label>
                    <input type="text" v-model="form.department" class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm" />
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
                  <button type="button" @click="saveEmployee" class="inline-flex w-full justify-center rounded-md border border-transparent bg-primary-600 px-4 py-2 text-base font-medium text-white shadow-sm hover:bg-primary-700 sm:ml-3 sm:w-auto sm:text-sm">حفظ</button>
                  <button type="button" @click="closeModal" class="mt-3 inline-flex w-full justify-center rounded-md border border-gray-300 bg-white px-4 py-2 text-base font-medium text-gray-700 shadow-sm hover:bg-gray-50 sm:mt-0 sm:w-auto sm:text-sm">إلغاء</button>
                </div>
              </DialogPanel>
            </TransitionChild>
          </div>
        </div>
      </Dialog>
    </TransitionRoot>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { Dialog, DialogPanel, DialogTitle, TransitionChild, TransitionRoot } from '@headlessui/vue'
import { PlusIcon, PencilIcon, TrashIcon } from '@heroicons/vue/24/outline'
import api from '@/api'

const employees = ref([])
const isModalOpen = ref(false)
const isEditing = ref(false)

const form = ref({
  id: null,
  employee_number: '',
  full_name: '',
  position: '',
  department: '',
  phone: '',
  email: ''
})

const fetchEmployees = async () => {
  try {
    const response = await api.get('/employees')
    employees.value = response.data
  } catch (error) {
    console.error('Failed to fetch employees:', error)
  }
}

const openCreateModal = () => {
  isEditing.value = false
  form.value = { id: null, employee_number: '', full_name: '', position: '', department: '', phone: '', email: '' }
  isModalOpen.value = true
}

const openEditModal = (employee) => {
  isEditing.value = true
  form.value = { ...employee }
  isModalOpen.value = true
}

const closeModal = () => {
  isModalOpen.value = false
}

const saveEmployee = async () => {
  try {
    if (isEditing.value) {
      await api.put(`/employees/${form.value.id}`, form.value)
    } else {
      await api.post('/employees', form.value)
    }
    await fetchEmployees()
    closeModal()
  } catch (error) {
    console.error('Failed to save employee:', error)
    alert('حدث خطأ أثناء حفظ البيانات')
  }
}

const confirmDelete = async (employee) => {
  if (confirm(`هل أنت متأكد من حذف الموظف ${employee.full_name}؟`)) {
    try {
      await api.delete(`/employees/${employee.id}`)
      await fetchEmployees()
    } catch (error) {
      console.error('Failed to delete employee:', error)
      alert(error.response?.data?.error || 'حدث خطأ أثناء حذف الموظف')
    }
  }
}

onMounted(() => {
  fetchEmployees()
})
</script>
