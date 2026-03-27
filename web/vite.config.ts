import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { VitePWA } from 'vite-plugin-pwa'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    react(),
    VitePWA({
      registerType: 'autoUpdate',
      includeAssets: ['favicon.ico', 'icons/*.png'],
      manifest: {
        name: 'Blue Ledger — Phi Beta Sigma',
        short_name: 'Blue Ledger',
        description: 'The gamified chapter management platform for Phi Beta Sigma',
        theme_color: '#001A4D',
        background_color: '#060D1A',
        display: 'standalone',
        orientation: 'portrait',
        start_url: '/',
        scope: '/',
        icons: [
          {
            src: '/icons/icon-192.png',
            sizes: '192x192',
            type: 'image/png',
          },
          {
            src: '/icons/icon-512.png',
            sizes: '512x512',
            type: 'image/png',
          },
          {
            src: '/icons/icon-512.png',
            sizes: '512x512',
            type: 'image/png',
            purpose: 'maskable',
          },
        ],
        categories: ['productivity', 'social'],
        shortcuts: [
          {
            name: 'Dashboard',
            url: '/dashboard',
            description: 'Go to your dashboard',
          },
          {
            name: 'Challenges',
            url: '/challenges',
            description: 'View active challenges',
          },
          {
            name: 'Leaderboard',
            url: '/leaderboard',
            description: 'See chapter XP rankings',
          },
        ],
      },
      workbox: {
        globPatterns: ['**/*.{js,css,html,ico,png,svg,woff2}'],
        runtimeCaching: [
          {
            // Cache API calls for 5 minutes
            urlPattern: /^https?:\/\/.*\/v1\/(members|events|announcements|leaderboard)/,
            handler: 'NetworkFirst',
            options: {
              cacheName: 'api-cache',
              expiration: {
                maxEntries: 100,
                maxAgeSeconds: 5 * 60,
              },
            },
          },
          {
            // Cache avatar images for 7 days
            urlPattern: /\.(png|jpg|jpeg|webp|svg)$/,
            handler: 'CacheFirst',
            options: {
              cacheName: 'image-cache',
              expiration: {
                maxEntries: 200,
                maxAgeSeconds: 7 * 24 * 60 * 60,
              },
            },
          },
        ],
        // Don't cache WebSocket or auth endpoints
        navigateFallbackDenylist: [/^\/v1\/ws/, /^\/v1\/auth/],
      },
    }),
  ],
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: ['./src/test/setup.ts'],
    exclude: ['node_modules', 'dist'],
  },
  server: {
    port: 3001,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, '/v1'),
      },
    },
  },
  build: {
    outDir: 'dist',
    sourcemap: true,
  },
})

