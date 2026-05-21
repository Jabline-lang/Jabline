// SEO Helper - Feature 0.8.0
export function usePageMeta(title, description, keywords = '', image = '') {
  React.useEffect(() => {
    // Title
    document.title = `${title} | Jabline`

    // Meta tags
    updateMetaTag('name', 'description', description)
    if (keywords) updateMetaTag('name', 'keywords', keywords)
    updateMetaTag('property', 'og:title', title)
    updateMetaTag('property', 'og:description', description)
    if (image) updateMetaTag('property', 'og:image', image)
    updateMetaTag('property', 'og:type', 'website')
    updateMetaTag('property', 'og:url', window.location.href)
    updateMetaTag('name', 'twitter:title', title)
    updateMetaTag('name', 'twitter:description', description)
  }, [title, description, keywords, image])
}

function updateMetaTag(type, name, content) {
  let tag = document.querySelector(`meta[${type}="${name}"]`)
  if (!tag) {
    tag = document.createElement('meta')
    tag.setAttribute(type, name)
    document.head.appendChild(tag)
  }
  tag.setAttribute('content', content)
}

// Analytics Helper - Feature 0.8.0
export function trackEvent(eventName, eventData = {}) {
  if (typeof window !== 'undefined' && window.gtag) {
    window.gtag('event', eventName, eventData)
  }
  console.log(`📊 Event tracked: ${eventName}`, eventData)
}

export function trackPageView(pageName) {
  if (typeof window !== 'undefined' && window.gtag) {
    window.gtag('config', 'GA_MEASUREMENT_ID', {
      page_path: window.location.pathname,
      page_title: pageName,
    })
  }
}

// PWA Helper - Feature 0.8.0
export function registerServiceWorker() {
  if ('serviceWorker' in navigator) {
    navigator.serviceWorker.register('/sw.js').then(registration => {
      console.log('✓ Service Worker registered')
    }).catch(error => {
      console.log('✗ Service Worker registration failed:', error)
    })
  }
}

export function checkAppInstallPromotion() {
  let deferredPrompt
  const installButton = document.getElementById('install-btn')

  window.addEventListener('beforeinstallprompt', (e) => {
    e.preventDefault()
    deferredPrompt = e
    if (installButton) installButton.style.display = 'block'
  })

  if (installButton) {
    installButton.addEventListener('click', async () => {
      if (deferredPrompt) {
        deferredPrompt.prompt()
        const { outcome } = await deferredPrompt.userChoice
        console.log(`App installation: ${outcome}`)
        deferredPrompt = null
      }
    })
  }
}
