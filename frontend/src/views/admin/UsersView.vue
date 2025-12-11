<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex justify-between items-center">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">إدارة المستخدمين</h1>
        <p class="mt-1 text-sm text-gray-500">إدارة حسابات المستخدمين وصلاحياتهم</p>
      </div>
      <button
        @click="openCreateModal"
        class="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-primary-600 hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500"
      >
        <PlusIcon class="-mr-1 ml-2 h-5 w-5" aria-hidden="true" />
        مستخدم جديد
      </button>
    </div>

    <!-- Users List -->
    <div class="bg-white shadow overflow-hidden sm:rounded-lg">
      <ul role="list" class="divide-y divide-gray-200">
        <li v-for="user in users" :key="user.id" class="px-4 py-4 sm:px-6 hover:bg-gray-50">
          <div class="flex items-center justify-between">
            <div class="flex items-center min-w-0">
              <div class="flex-shrink-0">
                <span class="inline-flex items-center justify-center h-10 w-10 rounded-full bg-gray-500">
                  <span class="text-sm font-medium leading-none text-white">{{ getInitials(user.full_name) }}</span>
                </span>
              </div>
              <div class="mr-4 truncate">
                <div class="flex items-center">
                  <p class="text-sm font-medium text-primary-600 truncate">{{ user.full_name }}</p>
                  <span
                    :class="[
                      user.is_active ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800',
                      'mr-2 px-2 inline-flex text-xs leading-5 font-semibold rounded-full'
                    ]"
                  >
                    {{ user.is_active ? 'نشط' : 'غير نشط' }}
                  </span>
                </div>
                <div class="flex items-center text-sm text-gray-500 mt-1">
                  <UserIcon class="flex-shrink-0 mr-1.5 h-4 w-4 text-gray-400" aria-hidden="true" />
                  <span class="truncate">{{ user.username }}</span>
                  <span class="mx-2">•</span>
                  <span class="truncate">{{ user.email }}</span>
                  <span class="mx-2">•</span>
                  <span class="truncate">{{ getRoleName(user.role) }}</span>
                </div>
              </div>
            </div>
            <div class="flex items-center space-x-2 space-x-reverse">
              <button
                @click="openEditModal(user)"
                class="p-2 text-gray-400 hover:text-gray-500"
                title="تعديل"
              >
                <PencilIcon class="h-5 w-5" aria-hidden="true" />
              </button>
              <template v-if="authStore.canDelete">
                <button
                  @click="toggleActive(user)"
                  :class="[
                    user.is_active ? 'text-yellow-500 hover:text-yellow-600' : 'text-green-500 hover:text-green-600',
                    'p-2'
                  ]"
                  :title="user.is_active ? 'تعطيل' : 'تنشيط'"
                >
                  <NoSymbolIcon v-if="user.is_active" class="h-5 w-5" aria-hidden="true" />
                  <CheckCircleIcon v-else class="h-5 w-5" aria-hidden="true" />
                </button>
                <button
                  @click="confirmDelete(user)"
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
      <div v-if="users.length === 0" class="text-center py-12">
        <UserGroupIcon class="mx-auto h-12 w-12 text-gray-400" />
        <h3 class="mt-2 text-sm font-medium text-gray-900">لا يوجد مستخدمين</h3>
        <p class="mt-1 text-sm text-gray-500">ابدأ بإضافة مستخدم جديد للنظام.</p>
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
                {{ isEditing ? 'تعديل مستخدم' : 'مستخدم جديد' }}
              </h3>
              <div class="mt-4 space-y-4">
                <!-- Username -->
                <div>
                  <label for="username" class="block text-sm font-medium text-gray-700">اسم المستخدم</label>
                  <div class="mt-1">
                    <input
                      type="text"
                      name="username"
                      id="username"
                      v-model="form.username"
                      :disabled="isEditing"
                      class="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm disabled:bg-gray-100"
                    />
                  </div>
                </div>

                <!-- Full Name -->
                <div>
                  <label for="full_name" class="block text-sm font-medium text-gray-700">الاسم الكامل (English)</label>
                  <div class="mt-1">
                    <input
                      type="text"
                      name="full_name"
                      id="full_name"
                      v-model="form.full_name"
                      class="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
                    />
                  </div>
                </div>

                <!-- Full Name AR -->
                <div>
                  <label for="full_name_ar" class="block text-sm font-medium text-gray-700">الاسم الكامل (عربي)</label>
                  <div class="mt-1">
                    <input
                      type="text"
                      name="full_name_ar"
                      id="full_name_ar"
                      v-model="form.full_name_ar"
                      class="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
                    />
                  </div>
                </div>

                <!-- Email -->
                <div>
                  <label for="email" class="block text-sm font-medium text-gray-700">البريد الإلكتروني</label>
                  <div class="mt-1">
                    <input
                      type="email"
                      name="email"
                      id="email"
                      v-model="form.email"
                      class="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
                    />
                  </div>
                </div>

                <!-- Password (Create Only) -->
                <div v-if="!isEditing">
                  <label for="password" class="block text-sm font-medium text-gray-700">كلمة المرور</label>
                  <div class="mt-1">
                    <input
                      type="password"
                      name="password"
                      id="password"
                      v-model="form.password"
                      class="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
                    />
                  </div>
                </div>

                <!-- Role -->
                <div>
                  <label for="role" class="block text-sm font-medium text-gray-700">الدور</label>
                  <div class="mt-1">
                    <select
                      id="role"
                      name="role"
                      v-model="form.role"
                      class="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
                    >
                      <option value="admin">مدير نظام</option>
                      <option value="finance">موظف مالية</option>
                      <option value="auditor">مدقق حسابات</option>
                    </select>
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

                <!-- Granular Permissions (Removed) -->
                <!-- <div class="border-t border-gray-200 pt-4 mt-4"> ... </div> -->
              </div>
            </div>
          </div>
          <div class="mt-5 sm:mt-4 sm:flex sm:flex-row-reverse">
            <button
              type="button"
              class="inline-flex w-full justify-center rounded-md border border-transparent bg-primary-600 px-4 py-2 text-base font-medium text-white shadow-sm hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500 sm:ml-3 sm:w-auto sm:text-sm"
              @click="saveUser"
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
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { Dialog, DialogPanel, DialogTitle, TransitionChild, TransitionRoot } from '@headlessui/vue'
import { PlusIcon, PencilIcon, TrashIcon, XMarkIcon, UserIcon, UserGroupIcon, NoSymbolIcon, CheckCircleIcon } from '@heroicons/vue/24/outline'

import api from '@/api'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const users = ref([])
const isModalOpen = ref(false)
const isEditing = ref(false)
const loading = ref(false)
const currentUserId = computed(() => authStore.user?.id)

const form = ref({
  id: null,
  username: '',
  full_name: '',
  full_name_ar: '',
  email: '',
  password: '',
  role: 'finance',
  is_active: true,
  role: 'finance',
  is_active: true
})


const getInitials = (name) => {
  if (!name) return ''
  const names = name.trim().split(' ')
  if (names.length === 1) return names[0].substring(0, 2).toUpperCase()
  return (names[0][0] + names[names.length - 1][0]).toUpperCase()
}

const fetchUsers = async () => {
  try {
    const response = await api.get('/users')
    users.value = response.data
  } catch (error) {
    console.error('Failed to fetch users:', error)
  }
}

const openCreateModal = () => {
  isEditing.value = false
  form.value = {
    id: null,
    username: '',
    full_name: '',
    full_name_ar: '',
    email: '',
    password: '',
    role: 'finance',
    role: 'finance',
    is_active: true
  }
  isModalOpen.value = true
}

const openEditModal = (user) => {
  isEditing.value = true
  form.value = {
    id: user.id,
    username: user.username,
    full_name: user.full_name,
    full_name_ar: user.full_name_ar,
    email: user.email,
    password: '', // Password not shown
    role: user.role,
    role: user.role,
    is_active: user.is_active
  }
  isModalOpen.value = true
}

const closeModal = () => {
  isModalOpen.value = false
}

const saveUser = async () => {
  loading.value = true
  try {
    // Validation
    const isNewUser = !form.value.id
    if (!form.value.username || !form.value.full_name || !form.value.email || (isNewUser && !form.value.password)) {
      alert('يرجى ملء جميع الحقول المطلوبة')
      return
    }

    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
    if (!emailRegex.test(form.value.email)) {
      alert('البريد الإلكتروني غير صالح. يجب أن يكون بصيغة: example@domain.com')
      return
    }

    if (isEditing.value) {
      await api.put(`/users/${form.value.id}`, {
        full_name: form.value.full_name,
        full_name_ar: form.value.full_name_ar,
        email: form.value.email,
        is_active: form.value.is_active,
        permissions: [] // Force empty permissions
      })
      // Update role separately if changed
      const originalUser = users.value.find(u => u.id === form.value.id)
      if (originalUser.role !== form.value.role) {
        await api.patch(`/users/${form.value.id}/role`, { role: form.value.role })
      }
    } else {
      await api.post('/users', form.value)
    }
    await fetchUsers()
    closeModal()
  } catch (error) {
    console.error('Failed to save user:', error)
    alert('حدث خطأ أثناء حفظ البيانات')
  } finally {
    loading.value = false
  }
}

const toggleActive = async (user) => {
  const action = user.is_active ? 'تعطيل' : 'تنشيط'
  if (confirm(`هل أنت متأكد من ${action} المستخدم ${user.full_name}؟`)) {
    try {
      await api.put(`/users/${user.id}`, {
        ...user,
        is_active: !user.is_active
      })
      await fetchUsers()
    } catch (error) {
      console.error('Failed to toggle user status:', error)
      alert('حدث خطأ أثناء تغيير حالة المستخدم')
    }
  }
}

const confirmDelete = async (user) => {
  if (confirm(`هل أنت متأكد من حذف المستخدم ${user.full_name} نهائياً؟ لا يمكن التراجع عن هذا الإجراء.`)) {
    try {
      await api.delete(`/users/${user.id}`)
      await fetchUsers()
    } catch (error) {
      console.error('Failed to delete user:', error)
      alert(error.response?.data?.error || 'حدث خطأ أثناء حذف المستخدم')
    }
  }
}

const getRoleName = (role) => {
  const roles = {
    admin: 'مدير نظام',
    finance: 'موظف مالية',
    auditor: 'مدقق حسابات'
  }
  return roles[role] || role
}

onMounted(() => {
  fetchUsers()
})
</script>
