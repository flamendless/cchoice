/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./client/components/*.templ", "./client/components/svg/*.templ"],
  theme: {
    extend: {
      colors: {
        cchoice: '#F6742F',
        cchoice_c: '#2FB1F6',
        searchbar: '#F7EFEA',
        cchoicesoft: '#F7EFEA',
      }
    },
  },
  plugins: [],
}

