import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import api from '@/api/client'

export const useAuthStore = defineStore('auth', () => {
    const user = ref(null)
    const isAuthenticated = computed(() => !!user.value)
    const userRole = computed(() => user.value?.role || null)

    // Check if user has specific role
    const hasRole = (role) => {
        return userRole.value === role
    }

    // Check if user can write (Admin or Finance)
    const canWrite = computed(() => {
        return userRole.value === 'admin' || userRole.value === 'finance'
    })

    // Check if user can delete (Admin only)
    const canDelete = computed(() => {
        return userRole.value === 'admin'
    })

    // Check if user can read (All authenticated users)
    const canRead = computed(() => {
        return isAuthenticated.value
    })

    // Login
    const login = async (username, password) => {
        try {
            const response = await api.post('/auth/login', { username, password })
            user.value = response.data.user
            return { success: true }
        } catch (error) {
            return {
                success: false,
                error: error.response?.data?.error || 'فشل تسجيل الدخول',
            }
        }
    }

    // Logout
    const logout = async () => {
        try {
            await api.post('/auth/logout')
        } catch (error) {
            console.error('Logout error:', error)
        } finally {
            user.value = null
        }
    }

    // Get current user
    const fetchCurrentUser = async () => {
        try {
            const response = await api.get('/users/me')
            user.value = response.data
            return true
        } catch (error) {
            user.value = null
            return false
        }
    }

    return {
        user,
        isAuthenticated,
        userRole,
        hasRole,
        canWrite,
        canDelete,
        canRead,
        login,
        logout,
        fetchCurrentUser,
    }
})
