import { useEffect } from 'react'

// Detectar servicios globales necesarios
export function initGlobalServices() {
  useEffect(() => {
    // Google Analytics - Feature 0.8.0
    const script = document.createElement('script')
    script.async = true
    script.src = 'https://www.googletagmanager.com/gtag/js?id=GA_ID'
    document.head.appendChild(script)

    window.dataLayer = window.dataLayer || []
    function gtag() { dataLayer.push(arguments) }
    window.gtag = gtag
    gtag('js', new Date())
    gtag('config', 'GA_ID')

    // Service Worker para PWA - Feature 0.8.0
    if ('serviceWorker' in navigator) {
      navigator.serviceWorker.register('/sw.js').catch(() => {})
    }

    // Dark mode persistence
    const savedTheme = localStorage.getItem('theme')
    if (savedTheme) {
      document.documentElement.classList.toggle('dark', savedTheme === 'dark')
    }
  }, [])
}
