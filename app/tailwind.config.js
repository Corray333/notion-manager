/** @type {import('tailwindcss').Config} */
export default {
  content: [],
  darkMode: false,
  purge: ['./index.html', './src/**/*.{vue,js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors:{
        accent: 'var(--accent)',
        dark: 'var(--dark)',
        success: 'var(--success)',
        error: 'var(--error)',
      },
    },
  },
  plugins: [],
}

