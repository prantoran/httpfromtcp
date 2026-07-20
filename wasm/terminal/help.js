function showHelp() {
      const helpText = [
        'Go HTTP Server — WebAssembly Terminal',
        '======================================',
        '',
        'This terminal runs a Go HTTP/1.1 server entirely in your browser',
        'using WebAssembly. Type curl commands to send requests.',
        '',
        'Available routes:',
        '  /              — 200 OK success page',
        '  /yourproblem   — 400 Bad Request error page',
        '  /myproblem     — 500 Internal Server Error page',
        '',
        'Sample commands:',
        '  curl localhost:42069/',
        '  curl localhost:42069/yourproblem',
        '  curl localhost:42069/myproblem',
        '  curl -X GET localhost:42069/',
        '  curl -i localhost:42069/yourproblem',
        '  curl -v localhost:42069/',
        '  curl http://localhost:42069/myproblem',
        '',
        'Supported curl flags:',
        '  -X METHOD          HTTP method (GET, POST, etc.)',
        '  -H "Key: Value"    Add a request header',
        '  -d "body"          Request body (implies POST)',
        '  -i                 Show response headers (default: on)',
        '  -v                 Verbose output (show request details)',
        '',
        'Special commands:',
        '  help     Show this help message',
        '  clear    Clear the terminal',
        '',
        'Note: "localhost:42069" is a simulated address — no real TCP port',
        'is opened. Requests are processed entirely in-browser by the Go',
        'WASM server through a Service Worker bridge.',
      ];

      for (const line of helpText) {
        appendLine(line, 'help');
      }
    }
