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

import { init, Terminal, FitAddon } from './ghostty-web.js';

document.addEventListener('DOMContentLoaded', async () => {
    const terminalElement = document.getElementById('terminal');
    if (!terminalElement) {
        console.error('Terminal element not found!');
        return;
    }

    // --- 1. Initialize Ghostty-Web Terminal ---
    try {
        await init(); // Initialize WASM
        console.log('Ghostty-Web WASM initialized.');
    } catch (e) {
        console.error('Failed to initialize Ghostty-Web WASM:', e);
        terminalElement.innerText = 'Terminal failed to load. Check console.';
        return;
    }

    let terminal;
    let fitAddon;

    try {
        terminal = new Terminal({
            cursorBlink: true,
            fontSize: 14,
            fontFamily: 'Menlo, Monaco, "Courier New", monospace',
            theme: {
                background: '#000000',
                foreground: '#ffffff',
            }
        });

        fitAddon = new FitAddon();
        terminal.loadAddon(fitAddon);

        terminal.open(terminalElement);
        fitAddon.fit();
        
        console.log('Ghostty-Web terminal opened.');
    } catch (e) {
        console.error('Failed to open Ghostty Terminal:', e);
        return;
    }

    // --- 2. WebSocket Connection ---
    const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsURL = `${wsProtocol}//${window.location.host}/ws`;
    const ws = new WebSocket(wsURL);

    ws.onopen = () => {
        console.log('WebSocket connection opened.');
        fitAddon.fit();
        sendResizeMessage();
        // Force focus so keyboard works immediately
        // Note: Ghostty-Web might handle focus differently, but we try standard method
        terminalElement.focus();
    };

    // --- 3. Data Flow (Server to Terminal) ---
    ws.onmessage = (event) => {
        // The server sends raw ANSI strings directly.
        // Ghostty-Web's write method handles ANSI parsing and rendering automatically.
        if (event.data) {
            terminal.write(event.data);
        }
    };

    ws.onclose = () => {
        console.log('WebSocket connection closed.');
        terminal.write('\r\n\x1b[31mConnection to server lost. Please refresh the page.\x1b[0m\r\n');
    };

    ws.onerror = (err) => {
        console.error('WebSocket error:', err);
        terminal.write(`\r\n\x1b[31mWebSocket error: ${err.message || 'Unknown error'}\x1b[0m\r\n`);
    };

    // --- 4. Data Flow (Terminal to Server) ---
    terminal.onData((data) => {
        if (ws.readyState !== WebSocket.OPEN) return;

        // Ghostty-Web sends raw input data (ANSI sequences, text, etc.)
        // We map this to the Terminus protocol.
        
        let msgType = 'key';
        let keyData = { keyType: 'runes', runes: [] };

        // Basic control code mapping
        let modifiers = { alt: false, ctrl: false, shift: false };
        
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
            case '\x1b[Z': // Shift+Tab
                keyData.keyType = 'tab';
                modifiers.shift = true;
                break;
            case '\x1b': // Escape
                keyData.keyType = 'escape';
                break;
            case '\x03': // Ctrl+C
                keyData.keyType = 'ctrl+c';
                modifiers.ctrl = true;
                break;
            case '\x12': // Ctrl+R
                keyData.keyType = 'ctrl+r';
                modifiers.ctrl = true;
                break;
            case '\x13': // Ctrl+S
                keyData.keyType = 'ctrl+s';
                modifiers.ctrl = true;
                break;
            case ' ':
                keyData.keyType = 'space';
                break;
            // Ghostty/xterm sequences for arrows
            case '\x1b[A': 
            case '\x1bOA':
                keyData.keyType = 'up';
                break;
            case '\x1b[B': 
            case '\x1bOB':
                keyData.keyType = 'down';
                break;
            case '\x1b[C': 
            case '\x1bOC':
                keyData.keyType = 'right';
                break;
            case '\x1b[D': 
            case '\x1bOD':
                keyData.keyType = 'left';
                break;
            default:
                // For everything else, send as runes
                keyData.runes = data.split('');
                break;
        }

        const clientMessage = {
            Type: msgType,
            Data: { ...keyData, modifiers },
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
        if (fitAddon) {
            fitAddon.fit();
            sendResizeMessage();
        }
    });
});
