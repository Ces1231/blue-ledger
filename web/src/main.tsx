import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ToastProvider } from './components/Toast'
import { App } from './App'
import './styles/base.css'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      // Don't refetch on window focus in dev — too noisy
      refetchOnWindowFocus: import.meta.env.PROD,
      // Retry once on failure (network hiccups), not infinitely
      retry: 1,
      // Stale after 1 minute by default; individual queries can override
      staleTime: 1000 * 60,
    },
    mutations: {
      // Don't retry mutations — side effects make retries dangerous
      retry: 0,
    },
  },
})

const container = document.getElementById('root')!
const root = createRoot(container)

root.render(
  <StrictMode>
    <BrowserRouter>
      <QueryClientProvider client={queryClient}>
        <ToastProvider>
          <App />
        </ToastProvider>
      </QueryClientProvider>
    </BrowserRouter>
  </StrictMode>
)
