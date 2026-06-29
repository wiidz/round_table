import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { RouterProvider } from 'react-router-dom'
import { Toaster } from 'sonner'

import { router } from '@/router'

import './index.css'

let scrollHideTimer: ReturnType<typeof setTimeout> | undefined

document.addEventListener(
  'scroll',
  () => {
    document.documentElement.dataset.scrolling = 'true'
    if (scrollHideTimer) clearTimeout(scrollHideTimer)
    scrollHideTimer = setTimeout(() => {
      delete document.documentElement.dataset.scrolling
    }, 900)
  },
  true,
)

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <RouterProvider router={router} />
    <Toaster richColors position="top-center" />
  </StrictMode>,
)
