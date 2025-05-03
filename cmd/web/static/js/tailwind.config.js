/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./cmd/web/**/*.html", "./cmd/web/**/*.templ",
  ],
  theme: {
    extend: {
      colors: {
        cchoice: '#F6742F',
        cchoice_c: '#2FB1F6',
        cchoice_border: '#F6b08A',
        searchbar: '#F7EFEA',
        cchoicesoft: '#F7EFEA',
        cchoice_dark: '#F46133',
      }
    },
  },
  plugins: [],
}

