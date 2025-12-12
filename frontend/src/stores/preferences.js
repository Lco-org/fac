import { defineStore } from 'pinia'
import { ref, watch, computed } from 'vue'

export const usePreferencesStore = defineStore('preferences', () => {
    // Theme: 'light' | 'dark' | 'sepia' | 'system'
    const theme = ref(localStorage.getItem('theme') || 'light')

    // Primary color (hex)
    const primaryColor = ref(localStorage.getItem('primaryColor') || '#1e40af')

    // Surplus color (positive balance) - green by default
    const surplusColor = ref(localStorage.getItem('surplusColor') || '#22c55e')

    // Deficit color (negative balance) - red by default
    const deficitColor = ref(localStorage.getItem('deficitColor') || '#ef4444')

    // Font size: 'small' | 'medium' | 'large'
    const fontSize = ref(localStorage.getItem('fontSize') || 'medium')

    // Border radius: 'sharp' | 'rounded' | 'pill'
    const borderRadius = ref(localStorage.getItem('borderRadius') || 'rounded')

    // Avatar (base64 or URL)
    const avatar = ref(localStorage.getItem('userAvatar') || '')

    // Computed: actual theme based on system preference
    const effectiveTheme = computed(() => {
        if (theme.value === 'system') {
            return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
        }
        return theme.value
    })

    // Apply theme to document
    const applyTheme = () => {
        const root = document.documentElement

        // Remove all theme classes
        root.classList.remove('theme-light', 'theme-dark', 'theme-sepia')

        // Add current theme class
        root.classList.add(`theme-${effectiveTheme.value}`)

        // Apply primary color as CSS variable
        root.style.setProperty('--color-primary', primaryColor.value)

        // Generate lighter/darker variants
        const hex = primaryColor.value.replace('#', '')
        const r = parseInt(hex.substring(0, 2), 16)
        const g = parseInt(hex.substring(2, 4), 16)
        const b = parseInt(hex.substring(4, 6), 16)

        root.style.setProperty('--color-primary-50', `rgba(${r}, ${g}, ${b}, 0.1)`)
        root.style.setProperty('--color-primary-100', `rgba(${r}, ${g}, ${b}, 0.2)`)
        root.style.setProperty('--color-primary-200', `rgba(${r}, ${g}, ${b}, 0.3)`)
        root.style.setProperty('--color-primary-300', `rgba(${r}, ${g}, ${b}, 0.5)`)
        root.style.setProperty('--color-primary-400', adjustBrightness(primaryColor.value, 20))
        root.style.setProperty('--color-primary-500', adjustBrightness(primaryColor.value, 10))
        root.style.setProperty('--color-primary-600', primaryColor.value)
        root.style.setProperty('--color-primary-700', adjustBrightness(primaryColor.value, -15))
        root.style.setProperty('--color-primary-800', adjustBrightness(primaryColor.value, -30))
        root.style.setProperty('--color-primary-900', adjustBrightness(primaryColor.value, -45))

        // Apply surplus color (positive balance)
        root.style.setProperty('--color-surplus', surplusColor.value)
        root.style.setProperty('--color-surplus-light', adjustBrightness(surplusColor.value, 60))
        root.style.setProperty('--color-surplus-dark', adjustBrightness(surplusColor.value, -30))

        // Apply deficit color (negative balance)
        root.style.setProperty('--color-deficit', deficitColor.value)
        root.style.setProperty('--color-deficit-light', adjustBrightness(deficitColor.value, 60))
        root.style.setProperty('--color-deficit-dark', adjustBrightness(deficitColor.value, -30))

        // Apply font size
        const fontSizes = {
            small: '14px',
            medium: '16px',
            large: '18px'
        }
        root.style.setProperty('--base-font-size', fontSizes[fontSize.value] || '16px')

        // Apply border radius
        const radii = {
            sharp: '4px',
            rounded: '8px',
            pill: '20px'
        }
        root.style.setProperty('--border-radius', radii[borderRadius.value] || '8px')
    }

    // Helper function to adjust color brightness
    function adjustBrightness(hex, percent) {
        hex = hex.replace('#', '')
        const num = parseInt(hex, 16)
        const amt = Math.round(2.55 * percent)
        const R = Math.max(0, Math.min(255, (num >> 16) + amt))
        const G = Math.max(0, Math.min(255, ((num >> 8) & 0x00FF) + amt))
        const B = Math.max(0, Math.min(255, (num & 0x0000FF) + amt))
        return `#${(0x1000000 + R * 0x10000 + G * 0x100 + B).toString(16).slice(1)}`
    }

    // Setters with persistence
    const setTheme = (newTheme) => {
        theme.value = newTheme
        localStorage.setItem('theme', newTheme)
        applyTheme()
    }

    const setPrimaryColor = (color) => {
        primaryColor.value = color
        localStorage.setItem('primaryColor', color)
        applyTheme()
    }

    const setSurplusColor = (color) => {
        surplusColor.value = color
        localStorage.setItem('surplusColor', color)
        applyTheme()
    }

    const setDeficitColor = (color) => {
        deficitColor.value = color
        localStorage.setItem('deficitColor', color)
        applyTheme()
    }

    const setFontSize = (size) => {
        fontSize.value = size
        localStorage.setItem('fontSize', size)
        applyTheme()
    }

    const setBorderRadius = (radius) => {
        borderRadius.value = radius
        localStorage.setItem('borderRadius', radius)
        applyTheme()
    }

    const setAvatar = (avatarData) => {
        avatar.value = avatarData
        localStorage.setItem('userAvatar', avatarData)
    }

    const removeAvatar = () => {
        avatar.value = ''
        localStorage.removeItem('userAvatar')
    }

    // Reset to defaults
    const resetToDefaults = () => {
        setTheme('light')
        setPrimaryColor('#1e40af')
        setSurplusColor('#22c55e')
        setDeficitColor('#ef4444')
        setFontSize('medium')
        setBorderRadius('rounded')
    }

    // Listen for system theme changes
    if (typeof window !== 'undefined') {
        window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
            if (theme.value === 'system') {
                applyTheme()
            }
        })
    }

    // Initialize theme on store creation
    applyTheme()

    return {
        theme,
        primaryColor,
        surplusColor,
        deficitColor,
        fontSize,
        borderRadius,
        avatar,
        effectiveTheme,
        setTheme,
        setPrimaryColor,
        setSurplusColor,
        setDeficitColor,
        setFontSize,
        setBorderRadius,
        setAvatar,
        removeAvatar,
        resetToDefaults,
        applyTheme
    }
})
