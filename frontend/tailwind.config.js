/** @type {import('tailwindcss').Config} */
export default {
    content: [
        "./index.html",
        "./src/**/*.{vue,js,ts,jsx,tsx}",
    ],
    theme: {
        extend: {
            colors: {
                primary: {
                    50: 'var(--color-primary-50, #f0f9ff)',
                    100: 'var(--color-primary-100, #e0f2fe)',
                    200: 'var(--color-primary-200, #bae6fd)',
                    300: 'var(--color-primary-300, #7dd3fc)',
                    400: 'var(--color-primary-400, #38bdf8)',
                    500: 'var(--color-primary-500, #0ea5e9)',
                    600: 'var(--color-primary-600, #0284c7)',
                    700: 'var(--color-primary-700, #0369a1)',
                    800: 'var(--color-primary-800, #075985)',
                    900: 'var(--color-primary-900, #0c4a6e)',
                },
                deficit: {
                    light: 'var(--color-deficit-light, #fef2f2)',
                    DEFAULT: 'var(--color-deficit, #ef4444)',
                    dark: 'var(--color-deficit-dark, #991b1b)',
                },
                surplus: {
                    light: 'var(--color-surplus-light, #f0fdf4)',
                    DEFAULT: 'var(--color-surplus, #22c55e)',
                    dark: 'var(--color-surplus-dark, #166534)',
                },
            },
        },
    },
    plugins: [],
}
