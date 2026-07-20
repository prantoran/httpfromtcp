/**
     * Initialize the Go WASM server:
     * 1. Load wasm_exec.js (already loaded via <script> tag)
     * 2. Fetch and instantiate main.wasm
     * 3. Register the Service Worker
     * 4. Enable the terminal input
     *
     * Each step is displayed in the terminal as a boot log entry.
     */
    async function boot() {
      try {
        // Step 1: wasm_exec.js is loaded via <script> tag
        appendLine('$ Loading wasm_exec.js...', 'boot-cmd');
        if (typeof Go === 'undefined') {
          appendLine('✗ Failed to load wasm_exec.js — Go class not found', 'boot-err');
          appendLine('  Make sure wasm_exec.js is in the same directory as index.html', 'boot-err');
          appendLine('  Run: make wasm-setup', 'boot-err');
          return;
        }
        appendLine('✓ wasm_exec.js loaded', 'boot-ok');

        // Step 2: Fetch and instantiate main.wasm
        appendLine('$ Fetching main.wasm...', 'boot-cmd');
        const go = new Go();

        let wasm;
        try {
          const result = await WebAssembly.instantiateStreaming(
            fetch('main.wasm'),
            go.importObject
          );
          wasm = result;
        } catch (streamErr) {
          // Fallback for servers that don't set application/wasm MIME type
          appendLine('  (streaming failed, trying ArrayBuffer fallback...)', 'boot-info');
          const response = await fetch('main.wasm');
          const bytes = await response.arrayBuffer();
          wasm = await WebAssembly.instantiate(bytes, go.importObject);
        }

        appendLine('✓ main.wasm loaded and instantiated', 'boot-ok');

        // Step 3: Start the Go runtime (this calls main() in main_wasm.go)
        // go.run() is async — it starts the Go program and resolves when it exits.
        // Since main_wasm.go uses select{}, it never exits.
        appendLine('$ Starting Go runtime...', 'boot-cmd');
        go.run(wasm.instance);

        // Give the Go runtime a moment to register handleHTTPRequest
        await new Promise(resolve => setTimeout(resolve, 100));

        if (typeof globalThis.handleHTTPRequest === 'function') {
          appendLine('✓ Go HTTP server initialized', 'boot-ok');
        } else {
          appendLine('✗ handleHTTPRequest not registered — WASM may have failed', 'boot-err');
          return;
        }

        // Step 4: Register Service Worker
        appendLine('$ Registering Service Worker...', 'boot-cmd');
        try {
          await registerServiceWorker();
          appendLine('✓ Service Worker active and controlling this page', 'boot-ok');
        } catch (swErr) {
          // Service Worker is optional — direct calls to handleHTTPRequest still work
          appendLine(`⚠ Service Worker registration failed: ${swErr.message}`, 'boot-info');
          appendLine('  (Terminal will still work — requests go directly to WASM)', 'boot-info');
        }

        // Boot complete
        appendLine('', '');
        appendLine(`$ Go HTTP server is running on localhost:${42069}`, 'boot-ok');
        appendLine('Server ready. Type a curl command below, or type "help" for usage.', 'boot-ok');
        appendLine('', '');

        // Enable the terminal input
        commandInput.disabled = false;
        commandInput.placeholder = 'curl localhost:42069/';
        commandInput.focus();

      } catch (err) {
        appendLine(`✗ Boot failed: ${err.message}`, 'boot-err');
        console.error('Boot error:', err);
      }
    }

    // Start the boot sequence when the page loads
    boot();
