/**
     * Process a command entered in the terminal input.
     * Supports: curl, help, clear, and error messages for unknown commands.
     */
    async function processCommand(input) {
      const trimmed = input.trim();
      if (!trimmed) return;

      // Show the command in the terminal (like a real terminal echo)
      appendLine(`$ ${trimmed}`, 'prompt');

      // Add to history
      commandHistory.push(trimmed);
      historyIndex = commandHistory.length;

      if (trimmed === 'clear') {
        terminal.innerHTML = '';
        return;
      }

      if (trimmed === 'help') {
        showHelp();
        return;
      }

      if (trimmed.startsWith('curl ') || trimmed === 'curl') {
        if (trimmed === 'curl') {
          appendLine('curl: try \'curl --help\' for more information', 'error');
          appendLine('Usage: curl localhost:42069/', 'help');
          return;
        }

        const parsed = parseCurlCommand(trimmed);
        if (!parsed) {
          appendLine('Error: could not parse curl command', 'error');
          return;
        }

        await executeRequest(parsed);
        return;
      }

      // Unknown command
      const cmd = trimmed.split(' ')[0];
      appendLine(`command not found: ${cmd}. Try: curl localhost:42069/`, 'error');
    }
