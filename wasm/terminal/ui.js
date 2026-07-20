/**
     * Append a line of text to the terminal output.
     * @param {string} text - The text content to display
     * @param {string} className - CSS class for styling (e.g., 'boot-ok', 'error')
     */
    function appendLine(text, className = '') {
      const line = document.createElement('div');
      line.className = 'line' + (className ? ' ' + className : '');
      line.textContent = text;
      terminal.appendChild(line);
      terminal.scrollTop = terminal.scrollHeight;
    }

    /**
     * Display a multi-line raw HTTP response with syntax highlighting.
     * Status line gets one color, headers another, body gets default color.
     * @param {string} rawResponse - The complete raw HTTP/1.1 response
     */
    function displayResponse(rawResponse) {
      const lines = rawResponse.split('\n');
      let inHeaders = true;
      let headersDone = false;

      for (const line of lines) {
        const trimmed = line.replace(/\r$/, '');

        if (!headersDone && inHeaders) {
          if (trimmed.startsWith('HTTP/')) {
            // Status line: "HTTP/1.1 200 OK"
            appendLine(trimmed, 'response-status');
          } else if (trimmed === '') {
            // Blank line = end of headers
            appendLine('', '');
            headersDone = true;
            inHeaders = false;
          } else {
            // Header: "content-type: text/html"
            appendLine(trimmed, 'response-header');
          }
        } else {
          // Body
          appendLine(trimmed, 'response-body');
        }
      }
    }
