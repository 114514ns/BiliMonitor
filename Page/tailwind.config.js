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
    theme: {
        light: {
            color:{
                white: "#F2F2F8",
            }
        }

    },
    darkMode: "class",
    plugins: [heroui({
        /*
        "themes": {
            "light": {
                "colors": {
                    "default": {
                        "foreground": "#f2f2f8",
                        "DEFAULT": "#f2f2f8"
                    },
                    "primary": {
                        "foreground": "#f2f2f8",
                        "DEFAULT": "#66cc8a"
                    },
                    "background": "#f2f2f8",
                    "foreground": {
                        "foreground": "#f2f2f8",
                        "DEFAULT": "#000"
                    },
                }
            }
        }
        
         */
    })],
};