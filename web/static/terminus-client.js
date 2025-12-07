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

// This file contains the client-side logic for the Terminus web frontend.
// It initializes the Ghostty-Web terminal, establishes a WebSocket connection
// to the Go backend, and handles data flow between the terminal and the server.

document.addEventListener('DOMContentLoaded', () => {
    const terminalElement = document.getElementById('terminal');
    if (!terminalElement) {
        console.error('Terminal element not found!');
        return;
    }

    // --- 1. Initialize Ghostty-Web Terminal ---
    // Assuming Ghostty-Web has a similar API to xterm.js for initialization.
    // The Ghostty-Web WASM and JS glue code are expected to be loaded by index.html.
    let terminal;
    try {
        terminal = new GhosttyTerminal({
            // Assuming default options for now. Adjust as needed.
            cursorBlink: true,
            cols: 80, // Initial columns
            rows: 24, // Initial rows
        });
        terminal.open(terminalElement);
        console.log('Ghostty-Web terminal initialized.');
    } catch (e) {
        console.error('Failed to initialize Ghostty-Web Terminal:', e);
        // Fallback to a simple text area or alert if terminal fails
        terminalElement.innerText = 'Terminal failed to load. Please check console for errors.';
        return;
    }

    // --- 2. WebSocket Connection ---
    const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsURL = `${wsProtocol}//${window.location.host}/ws`;
    const ws = new WebSocket(wsURL);

    ws.onopen = () => {
        console.log('WebSocket connection opened.');
        // Initial resize message to server
        sendResizeMessage();
    };

    // --- 3. Data Flow (Server to Terminal) ---
    ws.onmessage = (event) => {
        // The server sends raw ANSI strings directly.
        terminal.write(event.data);
    };

    ws.onclose = () => {
        console.log('WebSocket connection closed.');
        terminal.writeln('\n\rConnection to server lost. Please refresh the page.');
    };

    ws.onerror = (err) => {
        console.error('WebSocket error:', err);
        terminal.writeln(`\n\rWebSocket error: ${err.message || err}`);
    };

    // --- 4. Data Flow (Terminal to Server) ---
    // Assuming Ghostty-Web has an onData event similar to xterm.js
    terminal.onData((data) => {
        // `data` will contain the raw input string from the terminal (e.g., a, b, ,  [A)
        let msgType = 'key';
        let keyData = { keyType: 'runes', runes: [] };

        // Simple mapping for common control keys. Ghostty-Web might provide more structured events.
        switch (data) {
            case '\r':
                keyData.keyType = 'enter';
                break;
            case '\x7f': // Backspace
                keyData.keyType = 'backspace';
                break;
            case '\t':
                keyData.keyType = 'tab';
                break;
            case '\x1b': // Escape
                keyData.keyType = 'escape';
                break;
            case '\x1b[A': // Up arrow
                keyData.keyType = 'up';
                break;
            case '\x1b[B': // Down arrow
                keyData.keyType = 'down';
                break;
            case '\x1b[C': // Right arrow
                keyData.keyType = 'right';
                break;
            case '\x1b[D': // Left arrow
                keyData.keyType = 'left';
                break;
            case '\x03': // Ctrl+C
                keyData.keyType = 'ctrl+c';
                break;
            case ' ':
                keyData.keyType = 'space';
                break;
            default:
                // For regular characters, `data` is the character itself
                // We send it as an array of runes, as expected by the Go backend
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
                    width: terminal.cols, // Assuming Ghostty-Web has .cols and .rows properties
                    height: terminal.rows,
                },
            };
            ws.send(JSON.stringify(clientMessage));
        }
    };

    // Listen for window resize events
    window.addEventListener('resize', () => {
        // This is a simple debounced resize. A more robust solution might use a proper debounce utility.
        // Assuming Ghostty-Web's fit method or manual resize.
        // For xterm.js, it's `terminal.fit()` or `terminal.resize(cols, rows)`.
        // Let's assume GhosttyTerminal has a `fit` method.
        terminal.fit(); // Recalculate cols/rows based on container size

        // Send new dimensions to the server
        sendResizeMessage();
    });

    // Initial fit after terminal is opened
    terminal.fit();
});

// Dummy GhosttyTerminal class for local development if Ghostty-Web is not yet integrated
// This allows terminus-client.js to run without immediate errors for basic functionality.
// In a real deployment, this would be replaced by the actual Ghostty-Web library.
class GhosttyTerminal {
    constructor(options) {
        this.options = options || {};
        this.element = null;
        this.cols = options.cols || 80;
        this.rows = options.rows || 24;
        this._dataCallbacks = [];
        console.warn('Using dummy GhosttyTerminal. Actual Ghostty-Web not loaded.');
    }

    open(parentElement) {
        this.element = parentElement;
        this.element.innerHTML = '<pre style="height: 100%; width: 100%; margin: 0; overflow: auto; background-color: black; color: white;"></pre>';
        this.preElement = this.element.querySelector('pre');
        this.preElement.addEventListener('keydown', this._handleKeyDown.bind(this));
        this.preElement.setAttribute('tabindex', '0'); // Make it focusable
        this.preElement.focus();
        this._updateSize();
    }

    write(data) {
        if (this.preElement) {
            // Very basic ANSI interpretation for dummy terminal: 
            // 1. Strip ANSI escape codes to prevent "unreadable" characters.
            // 2. Interpret some basic control codes like \r and \n.
            // Note: A real terminal implementation (like xterm.js or the real Ghostty-Web) would interpret these correctly.
            
            // Regex to match ANSI escape codes (CSI sequences)
            // \x1b\[[0-9;]*[a-zA-Z] matches things like ESC[31m or ESC[2J
            const ansiRegex = /\x1b\[[0-9;]*[a-zA-Z]/g;
            
            // Clean the data of raw ANSI codes for the dummy display
            const cleanData = data.replace(ansiRegex, '');

            // Handle backspaces (basic implementation)
            // This loop removes the character preceding a backspace
            let processedData = '';
            for (let i = 0; i < cleanData.length; i++) {
                if (cleanData[i] === '\b' || cleanData[i] === '\x7f') { // Backspace or Delete
                    processedData = processedData.slice(0, -1);
                } else {
                    processedData += cleanData[i];
                }
            }

            this.preElement.innerHTML += this._escapeHtml(processedData);
            this.preElement.scrollTop = this.preElement.scrollHeight;
        }
    }

    writeln(data) {
        this.write(data + '\n');
    }

    onData(callback) {
        this._dataCallbacks.push(callback);
    }

    fit() {
        if (this.element && this.preElement) {
            const fontHeight = 15; // Approximate font height in pixels
            const fontWidth = 8;  // Approximate font width in pixels

            const newRows = Math.floor(this.element.clientHeight / fontHeight);
            const newCols = Math.floor(this.element.clientWidth / fontWidth);

            if (newRows > 0 && newCols > 0) {
                this.rows = newRows;
                this.cols = newCols;
                console.log(`Dummy terminal resized to: ${this.cols}x${this.rows}`);
            }
            this._updateSize();
        }
    }

    _updateSize() {
        if (this.preElement) {
            this.preElement.style.fontSize = '14px'; // Ensure consistent font size for calculations
            this.preElement.style.lineHeight = '15px';
            this.preElement.style.fontFamily = 'monospace';
        }
    }

    _handleKeyDown(event) {
        // Basic key handling for dummy terminal
        let data = '';
        if (event.key.length === 1) {
            data = event.key;
        } else {
            switch (event.key) {
                case 'Enter': data = '\r'; break;
                case 'Backspace': data = '\x7f'; break;
                case 'Tab': data = '\t'; break;
                case 'Escape': data = '\x1b'; break;
                case 'ArrowUp': data = '\x1b[A'; break;
                case 'ArrowDown': data = '\x1b[B'; break;
                case 'ArrowRight': data = '\x1b[C'; break;
                case 'ArrowLeft': data = '\x1b[D'; break;
                case 'c': if (event.ctrlKey) data = '\x03'; break; // Ctrl+C
            }
        }

        if (data && this._dataCallbacks.length > 0) {
            this._dataCallbacks.forEach(cb => cb(data));
            event.preventDefault(); // Prevent default browser action
        }
    }

    _escapeHtml(text) {
        return text
            .replace(/&/g, "&amp;")
            .replace(/</g, "&lt;")
            .replace(/>/g, "&gt;")
            .replace(/"/g, "&quot;")
            .replace(/'/g, "&#039;")
            .replace(/\n/g, "<br>")
            .replace(/\r/g, ""); // Remove carriage returns for simple rendering
    }
}
