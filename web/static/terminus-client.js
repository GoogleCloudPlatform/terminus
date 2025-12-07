// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Terminus Client using xterm.js
document.addEventListener('DOMContentLoaded', () => {
    const terminalElement = document.getElementById('terminal');
    if (!terminalElement) {
        console.error('Terminal element not found!');
        return;
    }

    // --- 1. Initialize xterm.js Terminal ---
    // We use xterm.js as it's a robust, production-ready web terminal component
    // that handles ANSI escape sequences perfectly.
    let terminal;
    let fitAddon;

    try {
        terminal = new Terminal({
            cursorBlink: true,
            fontFamily: 'Menlo, Monaco, "Courier New", monospace',
            fontSize: 14,
            theme: {
                background: '#000000',
                foreground: '#ffffff',
            }
        });

        // Load FitAddon to automatically size the terminal to the container
        fitAddon = new FitAddon.FitAddon();
        terminal.loadAddon(fitAddon);

        terminal.open(terminalElement);
        fitAddon.fit();
        
        console.log('xterm.js terminal initialized.');
    } catch (e) {
        console.error('Failed to initialize Terminal:', e);
        terminalElement.innerText = 'Terminal failed to load. Please check console for errors.';
        return;
    }

    // --- 2. WebSocket Connection ---
    const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsURL = `${wsProtocol}//${window.location.host}/ws`;
    const ws = new WebSocket(wsURL);

    ws.onopen = () => {
        console.log('WebSocket connection opened.');
        // Initial fit and resize message
        fitAddon.fit();
        sendResizeMessage();
        terminal.focus();
    };

    // --- 3. Data Flow (Server to Terminal) ---
    ws.onmessage = (event) => {
        // The server sends raw ANSI strings directly.
        // xterm.js handles ANSI parsing and rendering automatically.
        if (event.data) {
            terminal.write(event.data);
        }
    };

    ws.onclose = () => {
        console.log('WebSocket connection closed.');
        terminal.writeln('\r\n\x1b[31mConnection to server lost. Please refresh the page.\x1b[0m');
    };

    ws.onerror = (err) => {
        console.error('WebSocket error:', err);
        terminal.writeln(`\r\n\x1b[31mWebSocket error: ${err.message || 'Unknown error'}\x1b[0m`);
    };

    // --- 4. Data Flow (Terminal to Server) ---
    terminal.onData((data) => {
        if (ws.readyState !== WebSocket.OPEN) return;

        // `data` contains the raw input sequence from xterm.js
        let msgType = 'key';
        let keyData = { keyType: 'runes', runes: [] };

        // Basic mapping for special keys if needed, but 'runes' usually covers most
        // xterm.js sends the correct escape sequences for arrow keys etc.
        // Terminus backend expects a slightly more structured object for some keys,
        // but raw rune processing is often sufficient for basic apps.
        // Let's keep the structured mapping for robustness with the current Go backend.

        // We check for common control codes
        switch (data) {
            case '\r': // Enter
                keyData.keyType = 'enter';
                break;
            case '\x7f': // Backspace
                keyData.keyType = 'backspace';
                break;
            case '\t': // Tab
                keyData.keyType = 'tab';
                break;
            case '\x1b': // Escape
                keyData.keyType = 'escape';
                break;
            case '\x03': // Ctrl+C
                keyData.keyType = 'ctrl+c';
                break;
            case ' ':
                keyData.keyType = 'space';
                break;
            // Arrow keys come as sequences (e.g. \x1b[A) which are > 1 char
            case '\x1b[A': 
                keyData.keyType = 'up';
                break;
            case '\x1b[B': 
                keyData.keyType = 'down';
                break;
            case '\x1b[C': 
                keyData.keyType = 'right';
                break;
            case '\x1b[D': 
                keyData.keyType = 'left';
                break;
            default:
                // For regular characters or unmapped sequences, send as runes
                keyData.runes = data.split('');
                break;
        }

        const clientMessage = {
            Type: msgType,
            Data: keyData,
        };
        ws.send(JSON.stringify(clientMessage));
    });

    // --- 5. Resize Handling ---
    const sendResizeMessage = () => {
        if (ws.readyState === WebSocket.OPEN) {
            const clientMessage = {
                Type: 'resize',
                Data: {
                    width: terminal.cols,
                    height: terminal.rows,
                },
            };
            ws.send(JSON.stringify(clientMessage));
        }
    };

    // Listen for window resize events
    window.addEventListener('resize', () => {
        // Fit the terminal to the container
        if (fitAddon) {
            fitAddon.fit();
            sendResizeMessage();
        }
    });
});