// Service Worker per QR Menu PWA
const CACHE_VERSION = 'v1.0.0';
const CACHE_NAME = 'qr-menu-' + CACHE_VERSION;
const urlsToCache = [
  '/',
  '/login',
  '/admin',
  '/static/style.css',
  '/static/script.js',
  '/static/offline.html',
  '/manifest.json'
];

// Install event
self.addEventListener('install', event => {
  event.waitUntil(
    caches.open(CACHE_NAME)
      .then(cache => {
        console.log('Cache aperta:', CACHE_NAME);
        return cache.addAll(urlsToCache);
      })
      .then(() => {
        console.log('Service Worker installato');
        return self.skipWaiting();
      })
  );
});

// Activate event
self.addEventListener('activate', event => {
  event.waitUntil(
    caches.keys()
      .then(cacheNames => {
        return Promise.all(
          cacheNames.map(cacheName => {
            if (cacheName !== CACHE_NAME) {
              console.log('Elimino cache vecchia:', cacheName);
              return caches.delete(cacheName);
            }
          })
        );
      })
      .then(() => {
        console.log('Service Worker attivato');
        return self.clients.claim();
      })
  );
});

// Fetch event - Network First with Cache Fallback
self.addEventListener('fetch', event => {
  const { request } = event;
  
  // Skip non-GET requests
  if (request.method !== 'GET') {
    return;
  }
  
  // Skip chrome extensions
  if (request.url.startsWith('chrome-extension://')) {
    return;
  }

  event.respondWith(
    fetch(request)
      .then(response => {
        // Clone the response per caching
        const cloned = response.clone();
        
        // Cache successful responses
        if (response.status === 200 && request.method === 'GET') {
          caches.open(CACHE_NAME).then(cache => {
            cache.put(request, cloned);
          });
        }
        
        return response;
      })
      .catch(() => {
        // Network failed, try cache
        return caches.match(request)
          .then(response => {
            if (response) {
              return response;
            }
            
            // Cache miss
            if (request.destination === 'document') {
              return caches.match('/offline.html');
            }
            
            return new Response('Offline - Resource not available', {
              status: 503,
              statusText: 'Service Unavailable',
              headers: new Headers({
                'Content-Type': 'text/plain'
              })
            });
          });
      })
  );
});

// Background sync
self.addEventListener('sync', event => {
  if (event.tag === 'sync-orders') {
    event.waitUntil(syncOrders());
  }
  if (event.tag === 'sync-analytics') {
    event.waitUntil(syncAnalytics());
  }
});

async function syncOrders() {
  try {
    // Sincronizza ordini pendenti
    const db = await openIndexedDB();
    const tx = db.transaction('pending_orders', 'readonly');
    const store = tx.objectStore('pending_orders');
    const orders = await store.getAll();
    
    for (const order of orders) {
      await fetch('/api/orders', {
        method: 'POST',
        body: JSON.stringify(order),
        headers: {
          'Content-Type': 'application/json'
        }
      });
      
      // Rimuovi dall'IndexedDB
      const delTx = db.transaction('pending_orders', 'readwrite');
      delTx.objectStore('pending_orders').delete(order.id);
    }
  } catch (error) {
    console.error('Errore sync ordini:', error);
  }
}

async function syncAnalytics() {
  try {
    // Sincronizza eventi analytics pendenti
    const db = await openIndexedDB();
    const tx = db.transaction('pending_analytics', 'readonly');
    const store = tx.objectStore('pending_analytics');
    const events = await store.getAll();
    
    for (const event of events) {
      await fetch('/api/analytics/track', {
        method: 'POST',
        body: JSON.stringify(event),
        headers: {
          'Content-Type': 'application/json'
        }
      });
      
      // Rimuovi dall'IndexedDB
      const delTx = db.transaction('pending_analytics', 'readwrite');
      delTx.objectStore('pending_analytics').delete(event.id);
    }
  } catch (error) {
    console.error('Errore sync analytics:', error);
  }
}

function openIndexedDB() {
  return new Promise((resolve, reject) => {
    const request = indexedDB.open('qr-menu-db', 1);
    
    request.onerror = () => reject(request.error);
    request.onsuccess = () => resolve(request.result);
    
    request.onupgradeneeded = (event) => {
      const db = event.target.result;
      
      if (!db.objectStoreNames.contains('pending_orders')) {
        db.createObjectStore('pending_orders', { keyPath: 'id' });
      }
      
      if (!db.objectStoreNames.contains('pending_analytics')) {
        db.createObjectStore('pending_analytics', { keyPath: 'id' });
      }
      
      if (!db.objectStoreNames.contains('cached_data')) {
        db.createObjectStore('cached_data', { keyPath: 'key' });
      }
    };
  });
}

// Message from client
self.addEventListener('message', event => {
  if (event.data && event.data.type === 'SKIP_WAITING') {
    self.skipWaiting();
  }
});
