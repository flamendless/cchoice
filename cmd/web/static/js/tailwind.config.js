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
        cchoice: "#d9480f",
        cchoice_c: "#2FB1F6",
        cchoice_border: "#EE531B",
        searchbar: "#F7EFEA",
        cchoicesoft: "#F7EFEA",
        cchoice_dark: "#EE531B",
      }
    },
  },
  plugins: [],
}

