(function () {
  if (!('serviceWorker' in navigator)) {
    return;
  }

  window.addEventListener('load', function () {
    navigator.serviceWorker.register('/service-worker.js').catch(function (err) {
      console.warn('Service worker registration failed:', err);
    });
  });

  var deferredPrompt = null;

  window.addEventListener('beforeinstallprompt', function (event) {
    event.preventDefault();
    deferredPrompt = event;
    showInstallButton();
  });

  function showInstallButton() {
    var existing = document.getElementById('pwa-install-btn');
    if (existing) {
      existing.style.display = 'inline-flex';
      return;
    }

    var btn = document.createElement('button');
    btn.id = 'pwa-install-btn';
    btn.type = 'button';
    btn.textContent = 'Installa app';
    btn.setAttribute('aria-label', 'Installa QR Menu come app');
    btn.style.cssText = [
      'position:fixed',
      'right:16px',
      'bottom:16px',
      'z-index:10000',
      'min-height:44px',
      'padding:10px 14px',
      'border-radius:999px',
      'border:1px solid #cbd5e1',
      'background:#0f172a',
      'color:#fff',
      'font-weight:700',
      'cursor:pointer',
      'box-shadow:0 10px 22px rgba(15,23,42,.25)'
    ].join(';');

    btn.addEventListener('click', async function () {
      if (!deferredPrompt) {
        return;
      }
      deferredPrompt.prompt();
      try {
        await deferredPrompt.userChoice;
      } catch (_) {
      }
      deferredPrompt = null;
      btn.remove();
    });

    document.body.appendChild(btn);
  }

  window.addEventListener('appinstalled', function () {
    var btn = document.getElementById('pwa-install-btn');
    if (btn) {
      btn.remove();
    }
    deferredPrompt = null;
  });
})();
