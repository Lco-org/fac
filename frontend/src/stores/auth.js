import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import api from '@/api/client'

export const useAuthStore = defineStore('auth', () => {
    const user = ref(null)
    const isAuthenticated = computed(() => !!user.value)
    const userRole = computed(() => user.value?.role || null)

    // 2FA State
    const requires2FA = ref(false)
    const pendingUserId = ref(null)

    // Check if user has specific role
    const hasRole = (role) => {
        return userRole.value === role
    }

    // User permissions
    const permissions = computed(() => user.value?.permissions || [])

    // Check if user has specific permission
    const hasPermission = (perm) => {
        // Admins automatically have all permissions (optional, but good practice)
        if (userRole.value === 'admin') return true
        return permissions.value.includes(perm)
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

    // Login Step 1 - Credentials
    const login = async (username, password) => {
        try {
            const response = await api.post('/auth/login', { username, password })

            // Check if 2FA is required
            if (response.data.requires_2fa) {
                requires2FA.value = true
                pendingUserId.value = response.data.pending_user_id
                return {
                    success: true,
                    requires2FA: true,
                    message: response.data.message || 'تم إرسال رمز التحقق إلى بريدك الإلكتروني'
                }
            }

            // No 2FA - login complete
            user.value = response.data.user
            requires2FA.value = false
            pendingUserId.value = null
            return { success: true, requires2FA: false }
        } catch (error) {
            return {
                success: false,
                error: error.response?.data?.error || 'فشل تسجيل الدخول',
            }
        }
    }

    // Login Step 2 - Verify 2FA OTP
    const verify2FA = async (code) => {
        if (!pendingUserId.value) {
            return { success: false, error: 'جلسة التحقق منتهية، يرجى تسجيل الدخول مجدداً' }
        }

        try {
            const response = await api.post('/auth/verify-2fa', {
                user_id: pendingUserId.value,
                code: code
            })

            user.value = response.data.user
            requires2FA.value = false
            pendingUserId.value = null
            return { success: true }
        } catch (error) {
            return {
                success: false,
                error: error.response?.data?.error || 'رمز التحقق غير صحيح',
            }
        }
    }

    // Resend OTP Code
    const resendOTP = async () => {
        if (!pendingUserId.value) {
            return { success: false, error: 'جلسة التحقق منتهية' }
        }

        try {
            const response = await api.post('/auth/resend-otp', {
                user_id: pendingUserId.value
            })
            return {
                success: true,
                message: response.data.message || 'تم إرسال رمز جديد'
            }
        } catch (error) {
            return {
                success: false,
                error: error.response?.data?.error || 'فشل إرسال الرمز',
            }
        }
    }

    // Cancel 2FA and reset state
    const cancel2FA = () => {
        requires2FA.value = false
        pendingUserId.value = null
    }

    // Logout
    const logout = async () => {
        try {
            await api.post('/auth/logout')
        } catch (error) {
            console.error('Logout error:', error)
        } finally {
            user.value = null
            requires2FA.value = false
            pendingUserId.value = null
        }
    }

    // Revoke all sessions (Kill Switch)
    const revokeAllSessions = async () => {
        try {
            await api.post('/auth/revoke-all-sessions')
            user.value = null
            return { success: true, message: 'تم إلغاء جميع الجلسات' }
        } catch (error) {
            return {
                success: false,
                error: error.response?.data?.error || 'فشل إلغاء الجلسات',
            }
        }
    }

    // Change password
    const changePassword = async (oldPassword, newPassword) => {
        try {
            await api.post('/auth/change-password', {
                old_password: oldPassword,
                new_password: newPassword
            })
            user.value = null
            return { success: true, message: 'تم تغيير كلمة المرور، يرجى تسجيل الدخول مجدداً' }
        } catch (error) {
            return {
                success: false,
                error: error.response?.data?.error || 'فشل تغيير كلمة المرور',
            }
        }
    }

    // Toggle 2FA
    const toggle2FA = async (enabled) => {
        try {
            const response = await api.post('/auth/toggle-2fa', { enabled })
            if (user.value) {
                user.value.two_factor_enabled = enabled
            }
            return { success: true, message: response.data.message }
        } catch (error) {
            return {
                success: false,
                error: error.response?.data?.error || 'فشل تحديث إعدادات التحقق الثنائي',
            }
        }
    }

    // Get current user
    const fetchCurrentUser = async () => {
        try {
            const response = await api.get('/users/me')
            user.value = response.data
            return true
            user.value = response.data
            return true
        } catch (error) {
            // If failed, try SSA
            return await fetchSSAUser()
        }
    }

    // Fetch SSA User
    const fetchSSAUser = async () => {
        try {
            // Use specific SSA endpoint to avoid 401 loop
            const response = await api.get('/ssa/me')
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
        permissions,
        requires2FA,
        pendingUserId,
        hasRole,
        hasPermission,
        canWrite,
        canDelete,
        canRead,
        login,
        verify2FA,
        resendOTP,
        cancel2FA,
        logout,
        revokeAllSessions,
        changePassword,
        toggle2FA,
        toggle2FA,
        fetchCurrentUser,
        fetchSSAUser,
    }
})
