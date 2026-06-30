/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{vue,js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        'bg-primary': 'var(--color-bg-primary)',
        'bg-secondary': 'var(--color-bg-secondary)',
        'border-default': 'var(--color-border-default)',
        'accent-blue': 'var(--color-accent-blue)',
        'text-primary': 'var(--color-primary-text)',
        'text-secondary': 'var(--color-secondary-text)',
        'text-muted': 'var(--color-muted-text)',
        'primary-text': 'var(--color-primary-text)',
        'secondary-text': 'var(--color-secondary-text)',
        'muted-text': 'var(--color-muted-text)',
        'status-success': 'var(--color-status-success)',
        'status-warning': 'var(--color-status-warning)',
        'status-error': 'var(--color-status-error)',
        'status-neutral': 'var(--color-status-neutral)',
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
