commandInput.addEventListener('keydown', async (e) => {
      if (e.key === 'Enter') {
        const value = commandInput.value;
        commandInput.value = '';
        await processCommand(value);
      } else if (e.key === 'ArrowUp') {
        // Navigate command history (older)
        e.preventDefault();
        if (historyIndex > 0) {
          historyIndex--;
          commandInput.value = commandHistory[historyIndex];
        }
      } else if (e.key === 'ArrowDown') {
        // Navigate command history (newer)
        e.preventDefault();
        if (historyIndex < commandHistory.length - 1) {
          historyIndex++;
          commandInput.value = commandHistory[historyIndex];
        } else {
          historyIndex = commandHistory.length;
          commandInput.value = '';
        }
      }
    });

    // Focus input when clicking anywhere in the terminal
    terminal.addEventListener('click', () => {
      commandInput.focus();
    });
