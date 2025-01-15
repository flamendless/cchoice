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
        cchoice_border: '#f6b08a',
        searchbar: '#F7EFEA',
        cchoicesoft: '#F7EFEA',
      }
    },
  },
  plugins: [],
}

