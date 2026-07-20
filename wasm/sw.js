// ====================================================================
// Service Worker — HTTP Request Interceptor for Go WASM Server
// ====================================================================

self.addEventListener('install', (event) => {
  console.log('[SW] Installing Service Worker...');
  self.skipWaiting();
});

self.addEventListener('activate', (event) => {
  console.log('[SW] Service Worker activated');
  event.waitUntil(self.clients.claim());
});

self.addEventListener('fetch', (event) => {
  const url = new URL(event.request.url);
  if (!url.pathname.startsWith('/api/')) {
    return;
  }
  event.respondWith(handleGoRequest(event.request, url));
});

async function handleGoRequest(request, url) {
  try {
    const path = url.pathname.replace(/^\/api/, '') || '/';
    const headers = [];
    request.headers.forEach((value, key) => {
      headers.push(`${key}: ${value}`);
    });
    const headersStr = headers.join('\n');

    let body = '';
    if (request.method !== 'GET' && request.method !== 'HEAD') {
      try {
        body = await request.text();
      } catch (e) {}
    }

    if (typeof globalThis.handleHTTPRequest !== 'function') {
      return new Response('Go WASM server not ready', { status: 503 });
    }

    let streamController;
    const stream = new ReadableStream({
      start(controller) {
        streamController = controller;
      }
    });

    const enqueueChunk = (chunk) => {
        if (streamController) streamController.enqueue(chunk);
    };
    const closeStream = () => {
        if (streamController) streamController.close();
    };

    const rawHeaders = await globalThis.handleHTTPRequest(
      request.method,
      path,
      headersStr,
      body,
      enqueueChunk,
      closeStream
    );

    return parseRawResponseToStream(rawHeaders, stream);
  } catch (err) {
    console.error('[SW] Error handling request:', err);
    return new Response(`Internal SW error: ${err.message}`, { status: 500 });
  }
}

function parseRawResponseToStream(rawHeaders, bodyStream) {
  const lines = rawHeaders.trim().split('\r\n');
  const statusLine = lines[0] || '';
  const statusMatch = statusLine.match(/^HTTP\/\d+\.\d+\s+(\d+)\s+(.*)/);

  let status = 200;
  let statusText = 'OK';
  if (statusMatch) {
    status = parseInt(statusMatch[1], 10);
    statusText = statusMatch[2];
  }

  const responseHeaders = new Headers();
  for (let i = 1; i < lines.length; i++) {
    const colonIdx = lines[i].indexOf(':');
    if (colonIdx > 0) {
      const key = lines[i].substring(0, colonIdx).trim();
      const value = lines[i].substring(colonIdx + 1).trim();
      responseHeaders.set(key, value);
    }
  }

  return new Response(bodyStream, {
    status: status,
    statusText: statusText,
    headers: responseHeaders,
  });
}
