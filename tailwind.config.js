/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./client/components/*.templ", "./client/components/svg/*.templ"],
  theme: {
    extend: {
      colors: {
        cchoice: '#F6742F',
        searchbar: '#F7EFEA',
      }
    },
  },
  plugins: [],
}

