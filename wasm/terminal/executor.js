/**
     * Execute a parsed curl command by sending a fetch() request through the
     * Service Worker, which routes it to the Go WASM handler.
     *
     * The fetch URL uses the /api/ prefix so the Service Worker knows to
     * intercept it (other requests like loading .js files pass through).
     *
     * @param {object} parsed - Output from parseCurlCommand()
     */
    async function executeRequest(parsed) {
      if (parsed.verbose) {
        appendLine(`> ${parsed.method} ${parsed.path} HTTP/1.1`, 'boot-cmd');
        appendLine(`> Host: localhost:${42069}`, 'boot-cmd');
        for (const h of parsed.headers) {
          appendLine(`> ${h}`, 'boot-cmd');
        }
        appendLine('>', 'boot-cmd');
      }

      try {
        // Build headers string for the Go handler
        // Format: "Header1: Value1\nHeader2: Value2"
        const headersStr = parsed.headers.join('\n');

        // Call the Go WASM handler directly via the global function.
        // This avoids the Service Worker round-trip for simplicity.
        // handleHTTPRequest is registered by wasm_bridge.go via syscall/js.
        if (typeof globalThis.handleHTTPRequest !== 'function') {
          appendLine('Error: Go WASM server is not ready. Please wait for boot to complete.', 'error');
          return;
        }

        let fullResponse = '';
        let resolveStream;
        const streamDone = new Promise(r => resolveStream = r);

        const rawHeaders = await globalThis.handleHTTPRequest(
          parsed.method,
          parsed.path,
          headersStr,
          parsed.body,
          (chunk) => {
             fullResponse += new TextDecoder().decode(chunk);
          },
          () => {
             resolveStream();
          }
        );

        await streamDone;

        if (parsed.verbose) {
          appendLine('', '');
        }

        displayResponse(rawHeaders + fullResponse);
      } catch (err) {
        appendLine(`Error: ${err.message}`, 'error');
      }
    }
