/**
     * Parse a curl command string into its components.
     * Supports: -X METHOD, -H "Header: Value", -d "body", -i, -v
     *
     * Examples:
     *   "curl localhost:42069/"
     *   "curl -X POST -H 'Content-Type: text/plain' -d 'hello' localhost:42069/"
     *   "curl -i -v http://localhost:42069/yourproblem"
     *
     * @param {string} input - The full curl command string
     * @returns {object} Parsed command: { method, path, headers, body, showHeaders, verbose }
     */
    function parseCurlCommand(input) {
      // Tokenize respecting quoted strings
      const tokens = [];
      let current = '';
      let inQuote = null;

      for (let i = 0; i < input.length; i++) {
        const ch = input[i];
        if (inQuote) {
          if (ch === inQuote) {
            inQuote = null;
          } else {
            current += ch;
          }
        } else if (ch === '"' || ch === "'") {
          inQuote = ch;
        } else if (ch === ' ' || ch === '\t') {
          if (current) {
            tokens.push(current);
            current = '';
          }
        } else {
          current += ch;
        }
      }
      if (current) tokens.push(current);

      // First token should be "curl"
      if (tokens.length === 0 || tokens[0] !== 'curl') {
        return null;
      }

      const result = {
        method: 'GET',
        path: '/',
        headers: [],
        body: '',
        showHeaders: true,  // -i is default on in our terminal
        verbose: false,
      };

      let i = 1;
      while (i < tokens.length) {
        const token = tokens[i];

        if (token === '-X' && i + 1 < tokens.length) {
          result.method = tokens[++i].toUpperCase();
        } else if (token === '-H' && i + 1 < tokens.length) {
          result.headers.push(tokens[++i]);
        } else if (token === '-d' && i + 1 < tokens.length) {
          result.body = tokens[++i];
          // -d implies POST if method wasn't explicitly set
          if (result.method === 'GET') {
            result.method = 'POST';
          }
        } else if (token === '-i') {
          result.showHeaders = true;
        } else if (token === '-v') {
          result.verbose = true;
        } else if (!token.startsWith('-')) {
          // This is the URL — extract the path
          let url = token;

          // Strip scheme: "http://localhost:42069/path" → "localhost:42069/path"
          url = url.replace(/^https?:\/\//, '');

          // Strip host: "localhost:42069/path" → "/path"
          // Also handle "localhost/path" (no port)
          url = url.replace(/^localhost(:\d+)?/, '');

          // Ensure path starts with /
          if (!url.startsWith('/')) {
            url = '/' + url;
          }

          result.path = url;
        }

        i++;
      }

      return result;
    }
