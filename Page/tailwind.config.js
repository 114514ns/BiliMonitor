const { heroui } = require("@heroui/react");

/** @type {import('tailwindcss').Config} */
module.exports = {
    content: [
        "./index.html",
        "./*.{js,ts,jsx,tsx}",
        "./node_modules/@heroui/theme/dist/**/*.{js,ts,jsx,tsx}",
    ],
    theme: {
        color:{
            white: "#F2F2F8",
        }
    },
    darkMode: "class",
    plugins: [heroui({
        addCommonColors: true,
    })],
};