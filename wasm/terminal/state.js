const terminal = document.getElementById('terminal');
    const logsTerminal = document.getElementById('logs-terminal');
    const commandInput = document.getElementById('command-input');

    // Command history for up/down arrow navigation (like a real terminal)
    const commandHistory = [];
    let historyIndex = -1;
