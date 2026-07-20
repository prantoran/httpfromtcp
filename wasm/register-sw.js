// ====================================================================
// Service Worker Registration Helper
// ====================================================================
//
// This script registers the Service Worker (sw.js) and provides a
// promise-based API for the boot sequence to wait until the Service
// Worker is active and controlling this page.
//
// Why a separate file:
// - Keeps the registration logic clean and reusable
// - The index.html boot sequence awaits registerServiceWorker() before
//   enabling the terminal input
//
// Service Worker lifecycle:
// 1. register() — browser downloads sw.js and starts installation
// 2. install event — Service Worker is installed (sw.js handles this)
// 3. activate event — Service Worker is ready to intercept fetch events
// 4. clients.claim() — Service Worker takes control of this page
//
// On first visit, the Service Worker may not control the page until
// clients.claim() is called in the activate event. The skipWaiting()
// call in sw.js ensures this happens immediately.
// ====================================================================

/**
 * Register the Service Worker and wait until it's controlling this page.
 *
 * @returns {Promise<ServiceWorkerRegistration>} The active registration
 * @throws {Error} If Service Workers are not supported or registration fails
 */
async function registerServiceWorker() {
  if (!('serviceWorker' in navigator)) {
    throw new Error('Service Workers are not supported in this browser');
  }

  try {
    // Register sw.js — the browser will download and install it
    const registration = await navigator.serviceWorker.register('sw.js', {
      // Scope controls which fetch() requests the Service Worker intercepts.
      // '/' means it intercepts all requests from this origin.
      scope: '/',
    });

    console.log('[register-sw] Service Worker registered:', registration.scope);

    // Wait for the Service Worker to become active
    if (registration.active) {
      console.log('[register-sw] Service Worker already active');
      return registration;
    }

    // The Service Worker might be installing or waiting
    const sw = registration.installing || registration.waiting;
    if (sw) {
      await new Promise((resolve) => {
        sw.addEventListener('statechange', () => {
          if (sw.state === 'activated') {
            resolve();
          }
        });
      });
    }

    // Wait for the Service Worker to claim this page.
    // This ensures fetch() requests from this page are intercepted.
    if (!navigator.serviceWorker.controller) {
      await new Promise((resolve) => {
        navigator.serviceWorker.addEventListener('controllerchange', resolve, {
          once: true,
        });
      });
    }

    console.log('[register-sw] Service Worker controlling page');
    return registration;

  } catch (err) {
    console.error('[register-sw] Registration failed:', err);
    throw err;
  }
}
