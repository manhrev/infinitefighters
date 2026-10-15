import { defineConfig } from 'vite'

export default defineConfig({
  server: {
    fs: { allow: ['..'] }, // art lives in ../assets
    proxy: {
      '/ws': { target: 'ws://localhost:8080', ws: true },
      '/levels': 'http://localhost:8080',
    },
  },
  // The Go binary embeds this folder: go:embed cannot reach outside its module.
  build: { outDir: '../infinitefighters-backend/web', emptyOutDir: true },
})
