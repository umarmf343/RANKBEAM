import type { Config } from "tailwindcss";

export default {
  content: ["./index.html", "./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        night: "#070815",
        aurora: {
          50: "#f4f6ff",
          100: "#e5ebfe",
          200: "#c8d4fd",
          300: "#a0b4fa",
          400: "#6c8af6",
          500: "#4a66f1",
          600: "#3448e5",
          700: "#2a38c9",
          800: "#2732a2",
          900: "#232d80"
        }
      },
      boxShadow: {
        glow: "0 10px 50px rgba(76, 102, 241, 0.35)"
      },
      fontFamily: {
        display: ["'Poppins'", "system-ui", "sans-serif"],
        body: ["'Inter'", "system-ui", "sans-serif"]
      }
    }
  },
  plugins: []
} satisfies Config;
