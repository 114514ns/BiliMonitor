const { heroui } = require("@heroui/react");

/** @type {import('tailwindcss').Config} */
module.exports = {
    content: [
        "./index.html",
        "./*.{js,ts,jsx,tsx}",
        "./node_modules/@heroui/theme/dist/**/*.{js,ts,jsx,tsx}",
        "./src/pages/**/*.{js,ts,jsx,tsx}",
        "./src/components/**/*.{js,ts,jsx,tsx}",
    ],
    darkMode: "class",
    plugins: [
        heroui({
            themes: {
                light: {
                    colors: {
                        background: "#EFEFF5", // or DEFAULT
                        primary: {
                            //... 50 to 900
                            foreground: "#EFEFF5",
                            DEFAULT: "#17C964",
                        },
                    },

                },
                dark: {
                    colors: {
                        background: "#000000", // or DEFAULT
                        primary: {
                            //... 50 to 900
                            foreground: "#000000",
                            DEFAULT: "#17C964",
                        },
                    },

                }
            },
        }),
    ],
};