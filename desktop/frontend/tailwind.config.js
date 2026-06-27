/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{vue,js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        'bg-primary': '#000000',
        'bg-secondary': '#0d0e15',
        'border-default': '#2e2e2e',
        'accent-blue': '#006efe',
        'text-primary': '#ededed',
        'text-secondary': '#a0a0a0',
        'text-muted': '#666666',
        'status-success': '#00ac3a',
        'status-warning': '#ffae00',
        'status-error': '#e2162a',
      },
      borderRadius: {
        'sm': '6px',
        'md': '12px',
        'lg': '16px',
        'full': '9999px',
      },
    },
  },
  plugins: [],
}
