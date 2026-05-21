// Service Worker - Feature 0.8.0 PWA Support
const CACHE_NAME = 'jabline-docs-v1'
const urlsToCache = [
  '/',
  '/index.html',
  '/favicon.svg',
  '/manifest.json',
  '/apple-touch-icon.png',
]

// Install event
self.addEventListener('install', (event) => {
  event.waitUntil(
    caches.open(CACHE_NAME)
      .then((cache) => cache.addAll(urlsToCache))
      .then(() => self.skipWaiting())
  )
})

// Activate event
self.addEventListener('activate', (event) => {
  event.waitUntil(
    caches.keys().then((cacheNames) => {
      return Promise.all(
        cacheNames.map((cacheName) => {
          if (cacheName !== CACHE_NAME) {
            return caches.delete(cacheName)
          }
        })
      )
    }).then(() => self.clients.claim())
  )
})

// Fetch event - Network first, fallback to cache
self.addEventListener('fetch', (event) => {
  if (event.request.method !== 'GET') {
    return
  }

  event.respondWith(
    fetch(event.request)
      .then((response) => {
        // Cache the response
        if (!response || response.status !== 200) {
          return response
        }

        const responseToCache = response.clone()
        caches.open(CACHE_NAME).then((cache) => {
          cache.put(event.request, responseToCache)
        })

        return response
      })
      .catch(() => {
        // Return from cache if network fails
        return caches.match(event.request)
          .then((response) => response || new Response('Offline - Page not cached'))
      })
  )
})

// Background sync for offline actions (future feature)
self.addEventListener('sync', (event) => {
  if (event.tag === 'sync-feedback') {
    event.waitUntil(syncFeedback())
  }
})

async function syncFeedback() {
  // Sync feedback submitted while offline
  const db = new Promise((resolve) => {
    const req = indexedDB.open('jabline-docs')
    req.onsuccess = () => resolve(req.result)
  })

  console.log('Syncing feedback...')
}
