package pwa

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"qr-menu/logger"
)

// PWAManager gestisce le funzionalità Progressive Web App
type PWAManager struct {
	mu                sync.RWMutex
	appName           string
	appShortName      string
	appDescription    string
	appStartURL       string
	appScope          string
	appThemeColor     string
	appBackgroundColor string
	appIcon           string
	staticPath        string
	serviceWorkerPath string
	manifestPath      string
}

// PWAConfig contiene la configurazione PWA
type PWAConfig struct {
	AppName           string
	AppShortName      string
	AppDescription    string
	AppStartURL       string
	AppScope          string
	AppThemeColor     string
	AppBackgroundColor string
	AppIcon           string
	StaticPath        string
}

var (
	defaultManager *PWAManager
	once           sync.Once
)

// GetPWAManager restituisce il singleton PWAManager
func GetPWAManager() *PWAManager {
	once.Do(func() {
		defaultManager = &PWAManager{
			appName:            "QR Menu System",
			appShortName:       "QR Menu",
			appDescription:     "Digital QR Code Menu System for Restaurants",
			appStartURL:        "/",
			appScope:           "/",
			appThemeColor:      "#2E7D32",
			appBackgroundColor: "#FFFFFF",
			appIcon:            "/static/icon-192x192.png",
			staticPath:         "static",
			serviceWorkerPath:  "static/service-worker.js",
			manifestPath:       "static/manifest.json",
		}
	})
	return defaultManager
}

// Init inizializza il PWA manager
func (pm *PWAManager) Init(config PWAConfig) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if config.AppName != "" {
		pm.appName = config.AppName
	}
	if config.AppShortName != "" {
		pm.appShortName = config.AppShortName
	}
	if config.AppDescription != "" {
		pm.appDescription = config.AppDescription
	}
	if config.AppStartURL != "" {
		pm.appStartURL = config.AppStartURL
	}
	if config.AppScope != "" {
		pm.appScope = config.AppScope
	}
	if config.AppThemeColor != "" {
		pm.appThemeColor = config.AppThemeColor
	}
	if config.AppBackgroundColor != "" {
		pm.appBackgroundColor = config.AppBackgroundColor
	}
	if config.AppIcon != "" {
		pm.appIcon = config.AppIcon
	}
	if config.StaticPath != "" {
		pm.staticPath = config.StaticPath
	}

	// Crea le directory necessarie
	if err := os.MkdirAll(pm.staticPath, 0755); err != nil {
		return fmt.Errorf("errore creazione directory static: %w", err)
	}

	// Genera il manifest.json
	if err := pm.generateManifest(); err != nil {
		logger.Warn("Errore generazione manifest.json", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Genera il service worker
	if err := pm.generateServiceWorker(); err != nil {
		logger.Warn("Errore generazione service worker", map[string]interface{}{
			"error": err.Error(),
		})
	}

	logger.Info("PWA manager inizializzato", map[string]interface{}{
		"app_name":       pm.appName,
		"app_short_name": pm.appShortName,
		"start_url":      pm.appStartURL,
		"scope":          pm.appScope,
	})

	return nil
}

// generateManifest genera il file manifest.json
func (pm *PWAManager) generateManifest() error {
	manifestContent := fmt.Sprintf(`{
  "name": "%s",
  "short_name": "%s",
  "description": "%s",
  "start_url": "%s",
  "scope": "%s",
  "display": "standalone",
  "orientation": "portrait-primary",
  "theme_color": "%s",
  "background_color": "%s",
  "icons": [
    {
      "src": "/static/icon-192x192.png",
      "sizes": "192x192",
      "type": "image/png",
      "purpose": "any"
    },
    {
      "src": "/static/icon-512x512.png",
      "sizes": "512x512",
      "type": "image/png",
      "purpose": "any"
    },
    {
      "src": "/static/icon-masked-192x192.png",
      "sizes": "192x192",
      "type": "image/png",
      "purpose": "maskable"
    },
    {
      "src": "/static/icon-masked-512x512.png",
      "sizes": "512x512",
      "type": "image/png",
      "purpose": "maskable"
    }
  ],
  "categories": ["business", "food"],
  "screenshots": [
    {
      "src": "/static/screenshot-540x720.png",
      "sizes": "540x720",
      "type": "image/png"
    },
    {
      "src": "/static/screenshot-1080x1440.png",
      "sizes": "1080x1440",
      "type": "image/png"
    }
  ],
  "shortcuts": [
    {
      "name": "View Orders",
      "short_name": "Orders",
      "description": "View your restaurant orders",
      "url": "/admin?tab=orders",
      "icons": [
        {
          "src": "/static/icon-96x96.png",
          "sizes": "96x96"
        }
      ]
    },
    {
      "name": "View Menu",
      "short_name": "Menu",
      "description": "View active menu",
      "url": "/admin?tab=menu",
      "icons": [
        {
          "src": "/static/icon-96x96.png",
          "sizes": "96x96"
        }
      ]
    }
  ],
  "share_target": {
    "action": "/share",
    "method": "POST",
    "enctype": "application/x-www-form-urlencoded",
    "params": {
      "title": "title",
      "text": "text",
      "url": "url"
    }
  }
}
`, pm.appName, pm.appShortName, pm.appDescription, pm.appStartURL, pm.appScope,
		pm.appThemeColor, pm.appBackgroundColor)

	filePath := filepath.Join(pm.staticPath, "manifest.json")
	err := os.WriteFile(filePath, []byte(manifestContent), 0644)
	if err != nil {
		return fmt.Errorf("errore scrittura manifest.json: %w", err)
	}

	logger.Info("Manifest.json generato", map[string]interface{}{
		"path": filePath,
	})

	return nil
}

// generateServiceWorker genera il file service-worker.js
func (pm *PWAManager) generateServiceWorker() error {
	swContent := `// Service Worker per QR Menu PWA
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
`

	filePath := filepath.Join(pm.staticPath, "service-worker.js")
	err := os.WriteFile(filePath, []byte(swContent), 0644)
	if err != nil {
		return fmt.Errorf("errore scrittura service-worker.js: %w", err)
	}

	logger.Info("Service Worker generato", map[string]interface{}{
		"path": filePath,
	})

	return nil
}

// GetManifest restituisce il contenuto del manifest.json
func (pm *PWAManager) GetManifest() (string, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	manifestPath := filepath.Join(pm.staticPath, "manifest.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return "", fmt.Errorf("errore lettura manifest: %w", err)
	}

	return string(data), nil
}

// GetServiceWorker restituisce il contenuto del service worker
func (pm *PWAManager) GetServiceWorker() (string, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	swPath := filepath.Join(pm.staticPath, "service-worker.js")
	data, err := os.ReadFile(swPath)
	if err != nil {
		return "", fmt.Errorf("errore lettura service worker: %w", err)
	}

	return string(data), nil
}

// GetOfflinePageHTML restituisce la pagina offline
func (pm *PWAManager) GetOfflinePageHTML() string {
	return `<!DOCTYPE html>
<html lang="it">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>QR Menu - Offline</title>
  <style>
    * {
      margin: 0;
      padding: 0;
      box-sizing: border-box;
    }
    
    body {
      font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
      background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
      min-height: 100vh;
      display: flex;
      align-items: center;
      justify-content: center;
      padding: 20px;
    }
    
    .offline-container {
      background: white;
      border-radius: 16px;
      padding: 40px;
      text-align: center;
      max-width: 500px;
      box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
    }
    
    .offline-icon {
      font-size: 64px;
      margin-bottom: 20px;
    }
    
    h1 {
      color: #333;
      margin-bottom: 10px;
      font-size: 28px;
    }
    
    p {
      color: #666;
      margin-bottom: 30px;
      font-size: 16px;
      line-height: 1.6;
    }
    
    .status {
      background: #f5f5f5;
      padding: 15px;
      border-radius: 8px;
      margin-bottom: 30px;
      font-size: 14px;
      color: #666;
    }
    
    .status.online {
      background: #e8f5e9;
      color: #2e7d32;
    }
    
    button {
      background: #667eea;
      color: white;
      border: none;
      padding: 12px 30px;
      border-radius: 8px;
      font-size: 16px;
      cursor: pointer;
      transition: background 0.3s;
    }
    
    button:hover {
      background: #764ba2;
    }
    
    .cached-data {
      background: #f9f9f9;
      border-radius: 8px;
      padding: 15px;
      margin-top: 20px;
      text-align: left;
    }
    
    .cached-item {
      padding: 8px;
      border-bottom: 1px solid #eee;
      font-size: 14px;
    }
  </style>
</head>
<body>
  <div class="offline-container">
    <div class="offline-icon">📵</div>
    <h1>You're Offline</h1>
    <p>Non sei connesso a internet in questo momento. Alcune funzionalità potrebbero non essere disponibili.</p>
    
    <div class="status" id="status">
      Controllando la connessione...
    </div>
    
    <button onclick="location.reload()">Riprova</button>
    
    <div class="cached-data" id="cached-data" style="display: none;">
      <h3 style="margin-bottom: 10px;">Dati disponibili offline:</h3>
      <div id="cached-list"></div>
    </div>
  </div>
  
  <script>
    // Controlla la connessione
    function checkConnection() {
      fetch('/ping', { method: 'HEAD', no-cors: true })
        .then(() => {
          document.getElementById('status').textContent = 'Online - La pagina si ricaricherà automaticamente';
          document.getElementById('status').className = 'status online';
          setTimeout(() => location.reload(), 2000);
        })
        .catch(() => {
          document.getElementById('status').textContent = 'Offline - I dati potrebbero non essere aggiornati';
        });
    }
    
    // Carica i dati in cache
    async function loadCachedData() {
      try {
        const db = await openIndexedDB();
        
        // Leggi ordini in cache
        const ordersTx = db.transaction('pending_orders', 'readonly');
        const orders = await new Promise(resolve => {
          const req = ordersTx.objectStore('pending_orders').getAll();
          req.onsuccess = () => resolve(req.result);
        });
        
        if (orders.length > 0) {
          document.getElementById('cached-data').style.display = 'block';
          const list = document.getElementById('cached-list');
          
          orders.forEach(order => {
            const item = document.createElement('div');
            item.className = 'cached-item';
            // Use string concatenation instead of template literals to avoid escaping issues
            item.innerHTML = '<strong>Order:</strong> ' + order.id + ' - ' + new Date(order.date).toLocaleDateString();
            list.appendChild(item);
          });
        }
      } catch (error) {
        console.error('Errore caricamento cache:', error);
      }
    }
    
    function openIndexedDB() {
      return new Promise((resolve, reject) => {
        const request = indexedDB.open('qr-menu-db', 1);
        request.onerror = () => reject(request.error);
        request.onsuccess = () => resolve(request.result);
      });
    }
    
    // Controlla connessione ogni 5 secondi
    checkConnection();
    setInterval(checkConnection, 5000);
    loadCachedData();
  </script>
</body>
</html>`
}
