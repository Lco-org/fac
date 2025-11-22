<script setup>
import { ref, computed } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const sidebarOpen = ref(true)

const menuItems = computed(() => {
  const items = [
    {
      title: 'لوحة التحكم',
      icon: '📊',
      path: '/',
      roles: ['admin', 'finance', 'auditor'],
    },
    {
      title: 'المقبوضات',
      icon: '💰',
      roles: ['admin', 'finance'],
      children: [
        { title: 'منح ومشاريع', path: '/income/grants' },
        { title: 'تبرعات واشتراكات', path: '/income/donations' },
      ],
    },
    {
      title: 'المدفوعات',
      icon: '💸',
      roles: ['admin', 'finance'],
      children: [
        { title: 'العهدة المالية', path: '/disbursements/custody' },
        { title: 'صرف مرتبات', path: '/disbursements/salaries' },
        { title: 'صرف تشغيلية', path: '/disbursements/operational' },
      ],
    },
    {
      title: 'التقارير',
      icon: '📈',
      roles: ['admin', 'finance', 'auditor'],
      children: [
        { title: 'الجرد اليومي', path: '/reports/daily-position' },
        { title: 'حركة المشاريع', path: '/reports/project-summary' },
        { title: 'تقرير العجز', path: '/reports/deficit' },
      ],
    },
  ]

  // Filter based on user role
  return items.filter(item => 
    item.roles.includes(authStore.userRole)
  )
})

const handleLogout = async () => {
  await authStore.logout()
  router.push('/login')
}
</script>

<template>
  <div class="min-h-screen bg-gray-50 flex">
    <!-- Sidebar -->
    <aside
      :class="[
        'bg-primary-800 text-white transition-all duration-300',
        sidebarOpen ? 'w-64' : 'w-20'
      ]"
    >
      <div class="p-4">
        <h1 v-if="sidebarOpen" class="text-xl font-bold">نظام ERP المالي</h1>
        <button
          @click="sidebarOpen = !sidebarOpen"
          class="mt-4 p-2 bg-primary-700 rounded-lg hover:bg-primary-600 w-full"
        >
          {{ sidebarOpen ? '◀' : '▶' }}
        </button>
      </div>

      <!-- Navigation Menu -->
      <nav class="mt-8">
        <div v-for="item in menuItems" :key="item.title" class="mb-2">
          <!-- Single Item -->
          <RouterLink
            v-if="item.path"
            :to="item.path"
            class="flex items-center px-4 py-3 hover:bg-primary-700 transition-colors"
            active-class="bg-primary-900"
          >
            <span class="text-2xl ml-3">{{ item.icon }}</span>
            <span v-if="sidebarOpen">{{ item.title }}</span>
          </RouterLink>

          <!-- Parent with Children -->
          <div v-else>
            <div class="flex items-center px-4 py-3 font-semibold">
              <span class="text-2xl ml-3">{{ item.icon }}</span>
              <span v-if="sidebarOpen">{{ item.title }}</span>
            </div>
            <div v-if="sidebarOpen && item.children" class="mr-8">
              <RouterLink
                v-for="child in item.children"
                :key="child.path"
                :to="child.path"
                class="block px-4 py-2 hover:bg-primary-700 rounded transition-colors"
                active-class="bg-primary-900"
              >
                {{ child.title }}
              </RouterLink>
            </div>
          </div>
        </div>
      </nav>
    </aside>

    <!-- Main Content -->
    <div class="flex-1 flex flex-col">
      <!-- Header -->
      <header class="bg-white shadow-sm">
        <div class="px-6 py-4 flex justify-between items-center">
          <h2 class="text-xl font-semibold text-gray-800">
            مرحباً، {{ authStore.user?.full_name }}
          </h2>
          <div class="flex items-center gap-4">
            <span class="text-sm text-gray-600">
              الصلاحية: 
              <span class="font-semibold">
                {{ authStore.userRole === 'admin' ? 'مدير' : authStore.userRole === 'finance' ? 'مالي' : 'مراجع' }}
              </span>
            </span>
            <button
              @click="handleLogout"
              class="btn-secondary text-sm"
            >
              تسجيل الخروج
            </button>
          </div>
        </div>
      </header>

      <!-- Page Content -->
      <main class="flex-1 p-6 overflow-auto">
        <RouterView />
      </main>
    </div>
  </div>
</template>
