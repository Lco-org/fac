<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { usePreferencesStore } from '@/stores/preferences'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const preferencesStore = usePreferencesStore() // Initialize preferences on app load

// Popover state
const showPopover = ref(false)
const currentMenu = ref('main') // 'main' or parent title
const popoverRef = ref(null)
const triggerRef = ref(null)

// Menu items definition
const menuItems = computed(() => {
  const items = [
    {
      title: 'لوحة التحكم',
      icon: '📊',
      path: '/',
      roles: ['admin', 'finance', 'auditor'],
      color: 'bg-blue-500',
    },
    {
      title: 'الإدارة',
      icon: '⚙️',
      roles: ['admin'],
      color: 'bg-gray-600',
      children: [
        { title: 'المستخدمين', path: '/admin/users', icon: '👤' },
        { title: 'الحسابات المصرفية', path: '/admin/containers', icon: '🏦' },
        { title: 'إدارة الفواتير', path: '/admin/invoices-management', icon: '🧾' },
        { title: 'إدارة الإيصالات', path: '/admin/receipts-management', icon: '🎫' },
        { title: 'المشاريع', path: '/admin/projects', icon: '📁' },
        { title: 'الموظفين', path: '/admin/employees', icon: '👨‍💼' },
        { title: 'الأعضاء المنتسبين', path: '/admin/members', icon: '👥' },
        { title: 'إعدادات', path: '/admin/dashboard-config', icon: '⚙️' },
        { title: 'المراسلات', path: '/admin/messaging', icon: '📧' },
      ],
    },
    {
      title: 'المقبوضات',
      icon: '💰',
      roles: ['admin', 'finance'],
      color: 'bg-green-500',
      children: [
        { title: 'منح ومشاريع', path: '/income/grants', icon: '🏛️' },
        { title: 'تبرعات واشتراكات', path: '/income/donations', icon: '❤️' },
      ],
    },
    {
      title: 'المدفوعات',
      icon: '💸',
      roles: ['admin', 'finance'],
      color: 'bg-red-500',
      children: [
        { title: 'العهدة المالية', path: '/disbursements/custody', icon: '💼' },
        { title: 'صرف مرتبات', path: '/disbursements/salaries', icon: '💵' },
        { title: 'صرف تشغيلية', path: '/disbursements/operational', icon: '🔧' },
      ],
    },
    {
      title: 'التحويلات',
      icon: '🔄',
      path: '/transfers',
      roles: ['admin', 'finance'],
      color: 'bg-purple-500',
    },
    {
      title: 'الموارد البشرية',
      icon: '👥',
      roles: ['admin', 'hr'],
      color: 'bg-indigo-500',
      children: [
        { title: 'الموظفين', path: '/hr/employees', icon: '👨‍💼' },
        { title: 'الخصومات', path: '/hr/deductions', icon: '➖' },
        { title: 'المكافآت', path: '/hr/bonuses', icon: '🎁' },
        { title: 'الرواتب', path: '/hr/salaries', icon: '💰' },
      ],
    },
    {
      title: 'التقارير',
      icon: '📈',
      roles: ['admin', 'finance', 'auditor', 'hr'],
      color: 'bg-teal-500',
      children: [
        { title: 'الجرد المفصل', path: '/reports/detailed-statement', icon: '📋' },
        { title: 'الموقف اليومي', path: '/reports/daily-position', icon: '📅' },
        ...(authStore.hasPermission('view_invoices') ? [{ title: 'تقرير الفواتير', path: '/reports/invoices', icon: '🧾' }] : []),
        { title: 'كشف حساب موظف', path: '/reports/employee-statement', icon: '👤' },
      ],
    },
  ]

  return items.filter(item => item.roles.includes(authStore.userRole))
})

// Get current submenu items
const currentChildren = computed(() => {
  if (currentMenu.value === 'main') return []
  const parent = menuItems.value.find(item => item.title === currentMenu.value)
  return parent?.children || []
})

// Toggle popover
const togglePopover = () => {
  showPopover.value = !showPopover.value
  if (showPopover.value) {
    currentMenu.value = 'main'
  }
}

// Handle menu item click
const handleItemClick = (item) => {
  if (item.path) {
    router.push(item.path)
    showPopover.value = false
    currentMenu.value = 'main'
  } else if (item.children) {
    currentMenu.value = item.title
  }
}

// Go back to main menu
const goBack = () => {
  currentMenu.value = 'main'
}

// Close popover when clicking outside
const handleClickOutside = (event) => {
  if (popoverRef.value && !popoverRef.value.contains(event.target) && 
      triggerRef.value && !triggerRef.value.contains(event.target)) {
    showPopover.value = false
    currentMenu.value = 'main'
  }
}

// Check if path is active
const isActive = (path) => {
  return route.path === path
}

// Logout handler
const handleLogout = async () => {
  await authStore.logout()
  router.push('/login')
}

// Close on escape
const handleEscape = (event) => {
  if (event.key === 'Escape' && showPopover.value) {
    showPopover.value = false
    currentMenu.value = 'main'
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
  document.addEventListener('keydown', handleEscape)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
  document.removeEventListener('keydown', handleEscape)
})
</script>

<template>
  <div class="min-h-screen bg-gray-50">
    <!-- Header -->
    <header class="bg-white shadow-md sticky top-0 z-50">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 py-3 flex items-center justify-between">
        
        <!-- Right Side: App Launcher (First in RTL) -->
        <div class="flex items-center gap-3">
          <!-- App Launcher Button -->
          <div class="relative">
            <button
              ref="triggerRef"
              @click.stop="togglePopover"
              class="app-launcher-btn"
              :class="{ 'active': showPopover }"
              title="القائمة"
            >
              <div class="grid grid-cols-3 gap-1">
                <span v-for="i in 9" :key="i" class="dot"></span>
              </div>
            </button>

            <!-- Popover Menu -->
            <Transition name="popover">
              <div
                v-if="showPopover"
                ref="popoverRef"
                class="popover-menu"
                @click.stop
              >
                <!-- Header -->
                <div class="popover-header">
                  <button
                    v-if="currentMenu !== 'main'"
                    @click="goBack"
                    class="back-btn"
                  >
                    →
                  </button>
                  <span class="popover-title">
                    {{ currentMenu === 'main' ? 'القائمة الرئيسية' : currentMenu }}
                  </span>
                </div>

                <!-- Main Menu Grid -->
                <div v-if="currentMenu === 'main'" class="menu-grid">
                  <div
                    v-for="item in menuItems"
                    :key="item.title"
                    @click="handleItemClick(item)"
                    class="menu-item"
                    :class="{ 'active': item.path && isActive(item.path) }"
                  >
                    <div class="item-icon" :class="item.color">
                      {{ item.icon }}
                    </div>
                    <span class="item-title">{{ item.title }}</span>
                    <span v-if="item.children" class="has-children">▼</span>
                  </div>
                </div>

                <!-- Submenu Grid -->
                <div v-else class="menu-grid submenu">
                  <div
                    v-for="child in currentChildren"
                    :key="child.path"
                    @click="handleItemClick(child)"
                    class="menu-item"
                    :class="{ 'active': isActive(child.path) }"
                  >
                    <div class="item-icon bg-gray-200">
                      {{ child.icon }}
                    </div>
                    <span class="item-title">{{ child.title }}</span>
                  </div>
                </div>
              </div>
            </Transition>
          </div>

          <!-- Logo & Title -->
          <h1 class="text-lg sm:text-xl font-bold text-primary-700 hidden sm:block">📊 نظام ERP المالي</h1>
        </div>

        <!-- Left Side: User Info & Actions -->
        <div class="flex items-center gap-2 sm:gap-4">
          <!-- User greeting -->
          <span class="text-sm text-gray-600 hidden lg:block">
            مرحباً، <span class="font-semibold">{{ authStore.user?.full_name }}</span>
          </span>

          <!-- Role badge -->
          <span class="text-xs px-2 py-1 bg-primary-100 text-primary-700 rounded-full hidden md:block">
            {{ authStore.userRole === 'admin' ? 'مدير' : authStore.userRole === 'finance' ? 'مالي' : authStore.userRole === 'hr' ? 'موارد بشرية' : 'مراجع' }}
          </span>

          <!-- Account Settings link -->
          <RouterLink
            to="/account"
            class="text-gray-600 hover:text-primary-600 transition-colors text-xl"
            title="إعدادات الحساب"
          >
            ⚙️
          </RouterLink>

          <!-- Logout button -->
          <button
            @click="handleLogout"
            class="btn-secondary text-xs sm:text-sm"
          >
            <span class="hidden sm:inline">تسجيل الخروج</span>
            <span class="sm:hidden">خروج</span>
          </button>
        </div>
      </div>
    </header>

    <!-- Main Content - Centered -->
    <main class="max-w-7xl mx-auto px-4 sm:px-6 py-6">
      <RouterView />
    </main>
  </div>
</template>

<style scoped>
/* App Launcher Button */
.app-launcher-btn {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  border: none;
  cursor: pointer;
  transition: all 0.2s;
}

.app-launcher-btn:hover,
.app-launcher-btn.active {
  background: #f3f4f6;
}

.app-launcher-btn .dot {
  width: 5px;
  height: 5px;
  background: #6b7280;
  border-radius: 50%;
}

.app-launcher-btn:hover .dot,
.app-launcher-btn.active .dot {
  background: #1e40af;
}

/* Popover Menu */
.popover-menu {
  position: absolute;
  top: calc(100% + 10px);
  right: 0;
  width: 320px;
  background: white;
  border-radius: 16px;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.15), 0 0 0 1px rgba(0, 0, 0, 0.05);
  overflow: hidden;
  z-index: 1000;
}

/* Mobile popover adjustments */
@media (max-width: 400px) {
  .popover-menu {
    width: calc(100vw - 32px);
    right: -8px;
  }
}

.popover-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px 20px;
  background: linear-gradient(135deg, #1e40af 0%, #3b82f6 100%);
  color: white;
}

.back-btn {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.2);
  border: none;
  color: white;
  font-size: 18px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.2s;
}

.back-btn:hover {
  background: rgba(255, 255, 255, 0.3);
}

.popover-title {
  font-size: 16px;
  font-weight: 600;
}

/* Menu Grid */
.menu-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;
  padding: 16px;
  max-height: 400px;
  overflow-y: auto;
}

.menu-grid.submenu {
  grid-template-columns: repeat(2, 1fr);
}

.menu-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 16px 8px;
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.2s;
  position: relative;
  text-align: center;
}

.menu-item:hover {
  background: #f3f4f6;
  transform: translateY(-2px);
}

.menu-item.active {
  background: #dbeafe;
}

.item-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  margin-bottom: 8px;
  transition: transform 0.2s;
}

.menu-item:hover .item-icon {
  transform: scale(1.1);
}

.item-title {
  font-size: 12px;
  color: #374151;
  font-weight: 500;
  line-height: 1.3;
}

.has-children {
  position: absolute;
  top: 8px;
  left: 8px;
  font-size: 8px;
  color: #9ca3af;
}

/* Popover Animation */
.popover-enter-active,
.popover-leave-active {
  transition: all 0.2s ease;
}

.popover-enter-from,
.popover-leave-to {
  opacity: 0;
  transform: translateY(-10px) scale(0.95);
}

/* Custom scrollbar for popover */
.menu-grid::-webkit-scrollbar {
  width: 6px;
}

.menu-grid::-webkit-scrollbar-track {
  background: #f1f1f1;
  border-radius: 3px;
}

.menu-grid::-webkit-scrollbar-thumb {
  background: #c1c1c1;
  border-radius: 3px;
}

.menu-grid::-webkit-scrollbar-thumb:hover {
  background: #a1a1a1;
}
</style>
