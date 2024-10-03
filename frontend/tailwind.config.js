/** @type {import('tailwindcss').Config} */
import colors from 'tailwindcss/colors'

export default {
  content: [],
  theme: {
    extend: {
      colors: {
        'primary': colors.amber,
        secondary: colors.slate,
        accent: colors.amber,
        neutral: colors.slate,
        'base-content': colors.slate,
        info: colors.cyan,
        success: colors.green,
        warning: colors.amber,
        error: colors.red,
        background: 'rgb(15, 23, 42)',
      },
    },
  },
  fontFamily: {
    sans: ['Graphik', 'ui-sans-serif', 'system-ui', 'sans-serif', "Apple Color Emoji", "Segoe UI Emoji", "Segoe UI Symbol", "Noto Color Emoji"],
  },
  plugins: [],
}

