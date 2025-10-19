import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    host: 'https://fifo-system-iq29.onrender.com',
    port: 5173,
  },
})
