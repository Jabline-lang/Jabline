/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,jsx,ts,tsx}",
  ],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        // Jabline dark palette
        jb: {
          950: '#030306',
          900: '#07080f',
          850: '#0b0d16',
          800: '#0f111a',
          750: '#13151f',
          700: '#181b28',
          650: '#1d2133',
          600: '#232740',
          550: '#2c3150',
          500: '#374060',
          400: '#505878',
          300: '#7a8296',
          200: '#a8b0c4',
          150: '#c8cedd',
          100: '#e0e4ef',
          50:  '#f2f4fb',
        },
        // Jabline accent — electric blue
        accent: {
          DEFAULT: '#3d7eff',
          light:   '#6aa3ff',
          lighter: '#99c0ff',
          dark:    '#2b5fd9',
          darker:  '#1a3fa8',
          glow:    'rgba(61,126,255,0.25)',
        },
        // Secondary accent — violet
        violet: {
          jb: '#7c6fff',
          light: '#9f95ff',
          dark: '#5448e0',
        },
        // Semantic
        success: '#22d3a5',
        warning: '#f7a41d',
        danger:  '#f05252',
        info:    '#60a5fa',
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', '-apple-system', 'sans-serif'],
        mono: ['"JetBrains Mono"', '"IBM Plex Mono"', '"Fira Code"', 'monospace'],
      },
      animation: {
        'fade-in':    'fadeIn 0.5s ease-out both',
        'slide-up':   'slideUp 0.6s ease-out both',
        'slide-in':   'slideIn 0.4s ease-out both',
        'glow-pulse': 'glowPulse 3s ease-in-out infinite',
        'shimmer':    'shimmer 2.5s infinite',
        'blink':      'blink 1s step-end infinite',
        'float':      'float 4s ease-in-out infinite',
      },
      keyframes: {
        fadeIn: {
          'from': { opacity: '0', transform: 'translateY(10px)' },
          'to':   { opacity: '1', transform: 'translateY(0)' },
        },
        slideUp: {
          'from': { opacity: '0', transform: 'translateY(28px)' },
          'to':   { opacity: '1', transform: 'translateY(0)' },
        },
        slideIn: {
          'from': { opacity: '0', transform: 'translateX(-16px)' },
          'to':   { opacity: '1', transform: 'translateX(0)' },
        },
        glowPulse: {
          '0%, 100%': { boxShadow: '0 0 0 0 rgba(247,164,29,0)' },
          '50%':      { boxShadow: '0 0 24px 6px rgba(247,164,29,0.2)' },
        },
        shimmer: {
          '0%':   { backgroundPosition: '-200% 0' },
          '100%': { backgroundPosition: '200% 0' },
        },
        blink: {
          '0%, 100%': { opacity: '1' },
          '50%':      { opacity: '0' },
        },
        float: {
          '0%, 100%': { transform: 'translateY(0)' },
          '50%':      { transform: 'translateY(-8px)' },
        },
      },
      boxShadow: {
        'jb':      '0 0 0 1px rgba(247,164,29,0.3), 0 4px 24px rgba(247,164,29,0.08)',
        'jb-lg':   '0 0 0 1px rgba(247,164,29,0.5), 0 8px 40px rgba(247,164,29,0.18)',
        'jb-inner':'inset 0 0 0 1px rgba(247,164,29,0.2)',
        'card':    '0 1px 3px rgba(0,0,0,0.5), 0 1px 2px rgba(0,0,0,0.6)',
        'card-hover': '0 6px 20px rgba(0,0,0,0.6), 0 2px 6px rgba(0,0,0,0.4)',
        'glow':    '0 0 30px rgba(247,164,29,0.15)',
      },
      backgroundImage: {
        'gradient-jb':       'linear-gradient(135deg, #f7a41d 0%, #7c6fff 100%)',
        'gradient-dark':     'linear-gradient(180deg, #0b0d16 0%, #07080f 100%)',
        'gradient-hero':     'radial-gradient(ellipse 80% 50% at 50% -10%, rgba(247,164,29,0.12) 0%, transparent 60%)',
        'gradient-radial':   'radial-gradient(var(--tw-gradient-stops))',
        'shimmer-gradient':  'linear-gradient(90deg, transparent 0%, rgba(247,164,29,0.08) 50%, transparent 100%)',
      },
    },
  },
  plugins: [],
}
