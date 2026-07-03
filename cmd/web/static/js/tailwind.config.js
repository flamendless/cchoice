/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./cmd/web/**/*.html", "./cmd/web/**/*.templ",
  ],
  theme: {
    screens: {
      sm: '640px',
      md: '768px',
      lg: '1024px',
      xl: '1280px',
    },
    extend: {
      colors: {
        primary: {
          DEFAULT: '#d9480f',
          dark: '#EE531B',
          emphasis: '#F6742F',
          hover: '#F46133',
          muted: '#F6b08A',
        },
        surface: {
          DEFAULT: '#F7EFEA',
        },
        accent: {
          DEFAULT: '#2FB1F6',
        },
      }
    },
  },
  safelist: [
    "bg-primary-emphasis",
    "bg-primary-muted",
    "max-lg:hidden",
  ],
  plugins: [],
}
