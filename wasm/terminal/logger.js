function appendLogLine(text, type = 'info') {
      const line = document.createElement('div');
      line.className = 'line log-' + type;
      line.textContent = text;
      logsTerminal.appendChild(line);
      logsTerminal.scrollTop = logsTerminal.scrollHeight;
    }

    const originalConsoleLog = console.log;
    const originalConsoleWarn = console.warn;
    const originalConsoleError = console.error;

    console.log = function(...args) {
      originalConsoleLog.apply(console, args);
      appendLogLine(args.map(a => typeof a === 'object' ? JSON.stringify(a) : a).join(' '), 'info');
    };

    console.warn = function(...args) {
      originalConsoleWarn.apply(console, args);
      appendLogLine(args.map(a => typeof a === 'object' ? JSON.stringify(a) : a).join(' '), 'warn');
    };

    console.error = function(...args) {
      originalConsoleError.apply(console, args);
      appendLogLine(args.map(a => typeof a === 'object' ? JSON.stringify(a) : a).join(' '), 'error');
    };
