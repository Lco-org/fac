<script setup>
import { ref, computed } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const sidebarOpen = ref(true)
const openMenus = ref([])

// Explicit navigation handler for debugging
const handleMenuClick = (path) => {
  console.log('Menu clicked:', path)
  try {
    router.push(path)
  } catch (e) {
    console.error('Navigation error:', e)
    alert('Navigation failed: ' + e.message)
  }
}

const toggleMenu = (title) => {
  if (!sidebarOpen.value) sidebarOpen.value = true
  
  const index = openMenus.value.indexOf(title)
  if (index === -1) {
    openMenus.value.push(title)
  } else {
    openMenus.value.splice(index, 1)
  }
}

const menuItems = computed(() => {
  const items = [
    {
      title: 'لوحة التحكم',
      icon: '📊',
      path: '/',
      roles: ['admin', 'finance', 'auditor'],
    },
    {
      title: 'الإدارة',
      icon: '⚙️',
      roles: ['admin'],
      children: [
        { title: 'المستخدمين', path: '/admin/users' },
        { title: 'الحسابات المصرفية', path: '/admin/containers' },
        { title: 'المشاريع', path: '/admin/projects' },
        { title: 'الموظفين', path: '/admin/employees' },
        { title: 'الأعضاء المنتسبين', path: '/admin/members' },
        { title: 'إعدادات لوحة التحكم', path: '/admin/dashboard-config' },
      ],
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
      title: 'التحويلات',
      icon: '🔄',
      path: '/transfers',
      roles: ['admin', 'finance'],
    },
    {
      title: 'الموارد البشرية',
      icon: '👥',
      roles: ['admin', 'hr'],
      children: [
        { title: 'الموظفين', path: '/hr/employees' },
        { title: 'الخصومات', path: '/hr/deductions' },
        { title: 'المكافآت', path: '/hr/bonuses' },
      ],
    },
    {
      title: 'التقارير',
      icon: '📈',
      roles: ['admin', 'finance', 'auditor', 'hr'],
      children: [
        { title: 'الجرد المفصل', path: '/reports/detailed-statement' },
        { title: 'الموقف اليومي', path: '/reports/daily-position' },
        ...(authStore.hasPermission('view_invoices') ? [{ title: 'تقرير الفواتير', path: '/reports/invoices' }] : []),
        { title: 'كشف حساب موظف', path: '/reports/employee-statement' },
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
        'bg-primary-800 text-white transition-all duration-300 flex flex-col h-screen sticky top-0 z-50',
        sidebarOpen ? 'w-64' : 'w-20'
      ]"
    >
      <div class="p-4 flex-shrink-0">
        <h1 v-if="sidebarOpen" class="text-xl font-bold">نظام ERP المالي</h1>
        <button
          @click="sidebarOpen = !sidebarOpen"
          class="mt-4 p-2 bg-primary-700 rounded-lg hover:bg-primary-600 w-full"
        >
          {{ sidebarOpen ? '◀' : '▶' }}
        </button>
      </div>

      <!-- Navigation Menu -->
      <nav class="mt-8 flex-1 overflow-y-auto custom-scrollbar">
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
            <div 
              class="flex items-center px-4 py-3 font-semibold cursor-pointer hover:bg-primary-700 transition-colors justify-between"
              @click="toggleMenu(item.title)"
            >
              <div class="flex items-center">
                <span class="text-2xl ml-3">{{ item.icon }}</span>
                <span v-if="sidebarOpen">{{ item.title }}</span>
              </div>
              <span v-if="sidebarOpen" class="text-sm">
                {{ openMenus.includes(item.title) ? '▼' : '◀' }}
              </span>
            </div>
            <div v-if="sidebarOpen && item.children && openMenus.includes(item.title)" class="pr-6 space-y-1 mt-1 bg-primary-800">
              <div
                v-for="child in item.children"
                :key="child.path"
                @click.stop="handleMenuClick(child.path)"
                class="block px-4 py-2 hover:bg-primary-700 rounded transition-colors cursor-pointer select-none text-sm"
                :class="{ 'bg-primary-900': $route.path === child.path, 'font-bold': $route.path === child.path }"
              >
                {{ child.title }}
              </div>
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
                {{ authStore.userRole === 'admin' ? 'مدير' : authStore.userRole === 'finance' ? 'مالي' : authStore.userRole === 'hr' ? 'موارد بشرية' : 'مراجع' }}
              </span>
            </span>
            <RouterLink
              to="/account/security"
              class="text-gray-600 hover:text-primary-600 transition-colors"
              title="إعدادات الأمان"
            >
              🔐
            </RouterLink>
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
