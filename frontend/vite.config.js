import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
server: {
    // Remova as configurações de host e port para usar os padrões do Docker
    // host: 'localhost',
    // port: 80,
  },
})
