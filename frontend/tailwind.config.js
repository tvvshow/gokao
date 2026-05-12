/** @type {import('tailwindcss').Config} */
// UI-01：调色板对齐 frontend/src/styles/theme.css 的语义 token。
//   primary   = 墨绿 ink   （主色，替代旧 sky-blue）
//   secondary = 石墨 graphite（中性，替代旧 fuchsia）
//   accent    = 琥珀 amber  （风险/重点）
//   error     = 朱砂 cinnabar
//   success   = 林木绿 moss
//   gray      = 暖灰 graphite（取代冷调 slate）
// 旧组件中 `bg-primary-600` 等 utility 不需要改名 —— 同名键自动换色。
export default {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        // 墨绿主色 ink —— 决策稳重感
        primary: {
          50:  '#eef5f0',
          100: '#d8e8de',
          200: '#b8d4c2',
          300: '#8eb89c',
          400: '#5e9476',
          500: '#3c7457',
          600: '#2a5b43',
          700: '#1f4533',
          800: '#163224',
          900: '#0e2018',
          950: '#071510',
        },
        // 石墨灰阶 secondary —— 替代旧 fuchsia 紫
        secondary: {
          50:  '#faf9f7',
          100: '#f1efeb',
          200: '#e3e0d9',
          300: '#c9c5bc',
          400: '#a39e93',
          500: '#797368',
          600: '#58534a',
          700: '#3d3934',
          800: '#2a2723',
          900: '#1a1814',
          950: '#0c0b09',
        },
        // 琥珀 accent —— 风险/重点提示
        accent: {
          50:  '#fff8eb',
          100: '#ffeec6',
          200: '#ffd97f',
          300: '#fbbf24',
          400: '#f59e0b',
          500: '#d97706',
          600: '#b45309',
          700: '#92400e',
          800: '#78350f',
          900: '#5b2a0a',
          950: '#3a1a05',
        },
        // 林木绿 success —— 与主墨绿区分（更亮、偏黄）
        success: {
          50:  '#f0f7ee',
          100: '#d8ebd1',
          200: '#b6d9aa',
          300: '#8cc480',
          400: '#6cb15f',
          500: '#4e9a44',
          600: '#3b7c33',
          700: '#2c5d27',
          800: '#1f4220',
          900: '#13291a',
          950: '#0a1610',
        },
        // 琥珀色 warning（与 accent 同源，留独立 ramp 给 Element Plus warning）
        warning: {
          50:  '#fff8eb',
          100: '#ffeec6',
          200: '#ffd97f',
          300: '#fbbf24',
          400: '#f59e0b',
          500: '#d97706',
          600: '#b45309',
          700: '#92400e',
          800: '#78350f',
          900: '#5b2a0a',
          950: '#3a1a05',
        },
        // 朱砂 error —— 警示/危险
        error: {
          50:  '#fdf2ef',
          100: '#fbe0d8',
          200: '#f5b7a4',
          300: '#ec8769',
          400: '#d8593a',
          500: '#b73a1f',
          600: '#962d17',
          700: '#72200f',
          800: '#52170a',
          900: '#330d05',
          950: '#1c0703',
        },
        // 暖灰 gray —— 取代冷调 slate（与纸感背景搭配）
        gray: {
          50:  '#faf9f7',
          100: '#f1efeb',
          200: '#e3e0d9',
          300: '#c9c5bc',
          400: '#a39e93',
          500: '#797368',
          600: '#58534a',
          700: '#3d3934',
          800: '#2a2723',
          900: '#1a1814',
          950: '#0c0b09',
        },
        // Paper —— 暖纸感页面背景（仅 bg-paper-* utility 用）
        paper: {
          50:  '#fdfaf3',
          100: '#f8f3e6',
          200: '#f1eadb',
          300: '#e6dec9',
          400: '#d9cfb6',
        },
      },
      fontFamily: {
        sans: [
          'Inter',
          'Manrope',
          '-apple-system',
          'BlinkMacSystemFont',
          'Segoe UI',
          'PingFang SC',
          'Hiragino Sans GB',
          'Microsoft YaHei',
          'sans-serif',
        ],
        serif: [
          'Source Han Serif SC',
          'Noto Serif CJK SC',
          'Songti SC',
          'Source Han Serif',
          'PingFang SC',
          'serif',
        ],
        mono: [
          'JetBrains Mono',
          'Fira Code',
          'Cascadia Code',
          'Menlo',
          'Consolas',
          'monospace',
        ],
      },
      fontSize: {
        'xs': ['0.75rem', { lineHeight: '1rem' }],
        'sm': ['0.875rem', { lineHeight: '1.25rem' }],
        'base': ['1rem', { lineHeight: '1.55rem' }],
        'lg': ['1.125rem', { lineHeight: '1.75rem' }],
        'xl': ['1.25rem', { lineHeight: '1.75rem' }],
        '2xl': ['1.5rem', { lineHeight: '2rem' }],
        '3xl': ['1.875rem', { lineHeight: '2.25rem' }],
        '4xl': ['2.25rem', { lineHeight: '2.5rem' }],
        '5xl': ['3rem', { lineHeight: '1' }],
        '6xl': ['3.75rem', { lineHeight: '1' }],
        '7xl': ['4.5rem', { lineHeight: '1' }],
        '8xl': ['6rem', { lineHeight: '1' }],
        '9xl': ['8rem', { lineHeight: '1' }],
      },
      spacing: {
        '18': '4.5rem',
        '88': '22rem',
        '128': '32rem',
        '144': '36rem',
      },
      borderRadius: {
        // 报告感偏方正 —— xl/2xl 收紧
        DEFAULT: '6px',
        'sm':   '4px',
        'md':   '6px',
        'lg':   '10px',
        'xl':   '14px',
        '2xl':  '18px',
        '3xl':  '24px',
        '4xl':  '2rem',
        '5xl':  '2.5rem',
      },
      boxShadow: {
        // 低饱和、薄阴影 —— 取代旧的浮夸大阴影
        'soft':   '0 1px 2px rgba(26, 24, 20, 0.06)',
        'medium': '0 2px 6px rgba(26, 24, 20, 0.07), 0 1px 2px rgba(26, 24, 20, 0.04)',
        'large':  '0 8px 24px rgba(26, 24, 20, 0.08), 0 2px 6px rgba(26, 24, 20, 0.04)',
        'glow':   '0 0 0 3px rgba(31, 69, 51, 0.18)',         // 主色聚焦光晕
        'glow-lg': '0 0 24px rgba(31, 69, 51, 0.12)',
      },
      animation: {
        'fade-in': 'fadeIn 0.4s ease-out',
        'slide-up': 'slideUp 0.3s ease-out',
        'slide-down': 'slideDown 0.3s ease-out',
        'scale-in': 'scaleIn 0.2s ease-out',
        'bounce-soft': 'bounceSoft 0.6s ease-out',
        'pulse-soft': 'pulseSoft 2s infinite',
        'float': 'float 3s ease-in-out infinite',
      },
      keyframes: {
        fadeIn: {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' },
        },
        slideUp: {
          '0%': { transform: 'translateY(10px)', opacity: '0' },
          '100%': { transform: 'translateY(0)', opacity: '1' },
        },
        slideDown: {
          '0%': { transform: 'translateY(-10px)', opacity: '0' },
          '100%': { transform: 'translateY(0)', opacity: '1' },
        },
        scaleIn: {
          '0%': { transform: 'scale(0.95)', opacity: '0' },
          '100%': { transform: 'scale(1)', opacity: '1' },
        },
        bounceSoft: {
          '0%, 20%, 53%, 80%, 100%': { transform: 'translate3d(0,0,0)' },
          '40%, 43%': { transform: 'translate3d(0, -8px, 0)' },
          '70%': { transform: 'translate3d(0, -4px, 0)' },
          '90%': { transform: 'translate3d(0, -2px, 0)' },
        },
        pulseSoft: {
          '0%, 100%': { opacity: '1' },
          '50%': { opacity: '0.8' },
        },
        float: {
          '0%, 100%': { transform: 'translateY(0px)' },
          '50%': { transform: 'translateY(-10px)' },
        },
      },
      backdropBlur: {
        xs: '2px',
      },
    },
  },
  plugins: [
    require('@tailwindcss/forms'),
    require('@tailwindcss/typography'),
  ],
}
