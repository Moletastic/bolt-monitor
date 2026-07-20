import type { Config } from 'tailwindcss'

const config: Config = {
  darkMode: ['class'],
  content: ['./app/**/*.{ts,tsx}', './components/**/*.{ts,tsx}', './lib/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        background: 'hsl(var(--background))',
        foreground: 'hsl(var(--foreground))',
        card: 'hsl(var(--card))',
        'card-foreground': 'hsl(var(--card-foreground))',
        popover: 'hsl(var(--popover))',
        'popover-foreground': 'hsl(var(--popover-foreground))',
        primary: 'hsl(var(--primary))',
        'primary-foreground': 'hsl(var(--primary-foreground))',
        secondary: 'hsl(var(--secondary))',
        'secondary-foreground': 'hsl(var(--secondary-foreground))',
        muted: 'hsl(var(--muted))',
        'muted-foreground': 'hsl(var(--muted-foreground))',
        accent: 'hsl(var(--accent))',
        'accent-foreground': 'hsl(var(--accent-foreground))',
        destructive: 'hsl(var(--destructive))',
        'destructive-foreground': 'hsl(var(--destructive-foreground))',
        border: 'hsl(var(--border))',
        input: 'hsl(var(--input))',
        ring: 'hsl(var(--ring))',
        surface: {
          lowest: 'hsl(var(--surface-lowest))',
          low: 'hsl(var(--surface-low))',
          DEFAULT: 'hsl(var(--surface))',
          high: 'hsl(var(--surface-high))',
          highest: 'hsl(var(--surface-highest))',
          bright: 'hsl(var(--surface-bright))',
        },
        status: {
          up: 'hsl(var(--status-up))',
          warn: 'hsl(var(--status-warn))',
          down: 'hsl(var(--status-down))',
          unknown: 'hsl(var(--status-unknown))',
        },
      },
      borderRadius: {
        sm: '0.125rem',
        DEFAULT: '0.25rem',
        md: '0.375rem',
        lg: '0.5rem',
        xl: '0.75rem',
      },
      fontFamily: {
        sans: ['var(--font-inter)'],
        mono: ['var(--font-jetbrains-mono)'],
      },
      fontSize: {
        'display-lg': ['2rem', { fontWeight: '700', lineHeight: '1.2', letterSpacing: '-0.02em' }],
        'headline-md': [
          '1.5rem',
          { fontWeight: '600', lineHeight: '1.3', letterSpacing: '-0.01em' },
        ],
        'title-sm': ['1.125rem', { fontWeight: '600', lineHeight: '1.4' }],
        'body-sm': ['0.8125rem', { fontWeight: '400', lineHeight: '1.5' }],
        'data-lg': ['1.125rem', { fontWeight: '500', lineHeight: '1.2' }],
        'data-md': ['0.875rem', { fontWeight: '500', lineHeight: '1.2' }],
        'label-caps': [
          '0.6875rem',
          { fontWeight: '700', lineHeight: '1', letterSpacing: '0.05em' },
        ],
      },
      spacing: {
        'dashboard-xs': '0.25rem',
        'dashboard-sm': '0.5rem',
        'dashboard-md': '1rem',
        'dashboard-lg': '1.5rem',
        'dashboard-xl': '3rem',
        'dashboard-gutter': '1rem',
        'dashboard-desktop-gutter': '2rem',
      },
      boxShadow: {
        panel: '0 12px 32px rgba(1, 15, 31, 0.28)',
      },
    },
  },
}

export default config
