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
        cchoice: "#F6742F",
        cchoice_c: "#2FB1F6",
        cchoice_border: "#F6b08A",
        searchbar: "#F7EFEA",
        cchoicesoft: "#F7EFEA",
        cchoice_dark: "#F46133",
      }
    },
  },
  plugins: [],
}

