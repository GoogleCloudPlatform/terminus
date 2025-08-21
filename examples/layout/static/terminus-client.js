/**
 * Copyright 2025 Google LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Terminus Client - Complete Implementation
(function() {
    'use strict';

    class TerminusClient {
        constructor() {
            this.ws = null;
            this.terminal = document.getElementById('terminal');
            this.connected = false;
            this.reconnectAttempts = 0;
            this.maxReconnectAttempts = 5;
            this.reconnectDelay = 1000;
            this.lines = [];
            this.cursorPosition = { x: 0, y: 0 };
            this.showCursor = true;
            this.cursorBlinkInterval = null;
            this.dimensions = { width: 80, height: 24 };
            this.ansiParser = new ANSIParser();
        }

        connect() {
            const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            const wsUrl = `${protocol}//${window.location.host}/ws`;

            try {
                this.ws = new WebSocket(wsUrl);
                this.setupWebSocketHandlers();
            } catch (err) {
                console.error('WebSocket connection failed:', err);
                this.scheduleReconnect();
            }
        }

        setupWebSocketHandlers() {
            this.ws.onopen = () => {
                console.log('Connected to Terminus server');
                this.connected = true;
                this.reconnectAttempts = 0;
                this.terminal.innerHTML = '';
                this.terminal.classList.remove('disconnected');
                
                // Send initial resize event
                this.calculateAndSendResize();
            };

            this.ws.onclose = () => {
                console.log('Disconnected from Terminus server');
                this.connected = false;
                this.terminal.classList.add('disconnected');
                this.showDisconnectedMessage();
                this.scheduleReconnect();
            };

            this.ws.onerror = (error) => {
                console.error('WebSocket error:', error);
            };

            this.ws.onmessage = (event) => {
                try {
                    const message = JSON.parse(event.data);
                    this.handleServerMessage(message);
                } catch (err) {
                    console.error('Failed to parse server message:', err);
                }
            };
        }

        scheduleReconnect() {
            if (this.reconnectAttempts >= this.maxReconnectAttempts) {
                this.showDisconnectedMessage('Failed to connect. Please refresh the page.');
                return;
            }

            this.reconnectAttempts++;
            const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);
            
            setTimeout(() => {
                console.log(`Reconnection attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts}`);
                this.connect();
            }, delay);
        }

        showDisconnectedMessage(message = 'Disconnected. Attempting to reconnect...') {
            this.terminal.innerHTML = `<div class="disconnected-message">${message}</div>`;
        }

        handleServerMessage(message) {
            switch (message.type) {
                case 'render':
                    this.render(message.data);
                    break;
                case 'clear':
                    this.clearScreen();
                    break;
                case 'updateLine':
                    this.updateLine(message.data.y, message.data.content);
                    break;
                case 'setCell':
                    this.setCell(message.data.x, message.data.y, message.data.rune, message.data.style);
                    break;
                case 'setCursor':
                    this.setCursor(message.data.x, message.data.y, message.data.visible);
                    break;
                case 'batch':
                    this.processBatch(message.data.commands);
                    break;
                default:
                    console.warn('Unknown message type:', message.type);
            }
        }

        render(data) {
            if (typeof data === 'string') {
                // Legacy string render
                this.terminal.innerHTML = this.ansiParser.parse(data);
            } else if (data.content) {
                // Structured render with content
                this.terminal.innerHTML = this.ansiParser.parse(data.content);
            } else if (data.lines) {
                // Line-based render
                this.lines = data.lines.map(line => this.ansiParser.parse(line));
                this.rebuildDisplay();
            }
            this.scrollToBottom();
        }

        clearScreen() {
            this.lines = [];
            this.terminal.innerHTML = '';
            this.cursorPosition = { x: 0, y: 0 };
        }

        updateLine(y, content) {
            this.ensureLines(y + 1);
            this.lines[y] = this.ansiParser.parse(content);
            this.rebuildDisplay();
        }

        setCell(x, y, rune, style) {
            this.ensureLines(y + 1);
            
            // Convert line to character array if needed
            if (!this.lineCharacters) {
                this.lineCharacters = {};
            }
            
            if (!this.lineCharacters[y]) {
                this.lineCharacters[y] = new Array(this.dimensions.width).fill(' ');
            }
            
            // Apply style and character
            const styledChar = style ? 
                `<span style="${this.styleToCSS(style)}">${this.escapeHtml(rune)}</span>` : 
                this.escapeHtml(rune);
            
            this.lineCharacters[y][x] = styledChar;
            
            // Rebuild the line
            this.lines[y] = this.lineCharacters[y].join('');
            this.rebuildDisplay();
        }

        setCursor(x, y, visible = true) {
            this.cursorPosition = { x, y };
            this.showCursor = visible;
            this.updateCursorDisplay();
        }

        processBatch(commands) {
            commands.forEach(cmd => {
                this.handleServerMessage(cmd);
            });
        }

        ensureLines(count) {
            while (this.lines.length < count) {
                this.lines.push('');
            }
        }

        rebuildDisplay() {
            // Lines are already parsed, just join them with <br> tags
            const content = this.lines.join('<br>');
            this.terminal.innerHTML = content;
            this.updateCursorDisplay();
        }

        updateCursorDisplay() {
            // Remove existing cursor
            const existingCursor = this.terminal.querySelector('.cursor');
            if (existingCursor) {
                existingCursor.remove();
            }

            if (!this.showCursor) return;

            // Add cursor at current position
            // This is a simplified implementation
            // A full implementation would insert the cursor at the exact character position
        }

        scrollToBottom() {
            this.terminal.scrollTop = this.terminal.scrollHeight;
        }

        styleToCSS(style) {
            const css = [];
            if (style.foreground) css.push(`color: ${style.foreground}`);
            if (style.background) css.push(`background-color: ${style.background}`);
            if (style.bold) css.push('font-weight: bold');
            if (style.italic) css.push('font-style: italic');
            if (style.underline) css.push('text-decoration: underline');
            if (style.strikethrough) css.push('text-decoration: line-through');
            return css.join('; ');
        }

        escapeHtml(text) {
            const div = document.createElement('div');
            div.textContent = text;
            return div.innerHTML;
        }

        sendMessage(type, data) {
            if (!this.connected || this.ws.readyState !== WebSocket.OPEN) {
                return;
            }

            const message = JSON.stringify({ type, data });
            this.ws.send(message);
        }

        sendKey(keyType, runes = null) {
            const data = { keyType };
            if (runes) {
                data.runes = runes;
            }
            this.sendMessage('key', data);
        }

        calculateAndSendResize() {
            // Get terminal element dimensions
            const rect = this.terminal.getBoundingClientRect();
            const computedStyle = window.getComputedStyle(this.terminal);
            
            // Calculate usable space
            const usableWidth = rect.width - 
                parseFloat(computedStyle.paddingLeft) - 
                parseFloat(computedStyle.paddingRight);
            const usableHeight = rect.height - 
                parseFloat(computedStyle.paddingTop) - 
                parseFloat(computedStyle.paddingBottom);
            
            // Create temporary element to measure character dimensions
            const measurer = document.createElement('span');
            measurer.style.position = 'absolute';
            measurer.style.visibility = 'hidden';
            measurer.style.whiteSpace = 'pre';
            measurer.textContent = 'W'; // Use 'W' as it's typically widest
            this.terminal.appendChild(measurer);
            
            const charWidth = measurer.getBoundingClientRect().width;
            const charHeight = parseFloat(computedStyle.lineHeight);
            
            this.terminal.removeChild(measurer);
            
            // Calculate dimensions
            const width = Math.floor(usableWidth / charWidth);
            const height = Math.floor(usableHeight / charHeight);
            
            // Update dimensions
            this.dimensions = { width, height };
            
            // Send to server
            this.sendMessage('resize', { width, height });
        }

        setupInputHandlers() {
            // Focus terminal on click
            this.terminal.addEventListener('click', () => {
                this.terminal.focus();
            });

            // Keyboard input
            this.terminal.addEventListener('keydown', (e) => {
                if (!this.connected) return;

                let handled = true;

                // Special key combinations
                if (e.ctrlKey || e.metaKey) {
                    switch (e.key.toLowerCase()) {
                        case 'c':
                            this.sendKey('ctrl+c');
                            break;
                        case 'v':
                            // Allow paste
                            handled = false;
                            break;
                        case 'a':
                            this.sendKey('ctrl+a');
                            break;
                        case 'd':
                            this.sendKey('ctrl+d');
                            break;
                        case 'e':
                            this.sendKey('ctrl+e');
                            break;
                        case 'k':
                            this.sendKey('ctrl+k');
                            break;
                        case 'l':
                            this.sendKey('ctrl+l');
                            break;
                        case 'r':
                            this.sendKey('ctrl+r');
                            break;
                        case 's':
                            this.sendKey('ctrl+s');
                            break;
                        case 'u':
                            this.sendKey('ctrl+u');
                            break;
                        case 'w':
                            this.sendKey('ctrl+w');
                            break;
                        case 'z':
                            this.sendKey('ctrl+z');
                            break;
                        default:
                            handled = false;
                    }
                } else if (e.altKey) {
                    switch (e.key.toLowerCase()) {
                        case 'b':
                            this.sendKey('alt+b');
                            break;
                        case 'f':
                            this.sendKey('alt+f');
                            break;
                        case 'd':
                            this.sendKey('alt+d');
                            break;
                        case 'backspace':
                            this.sendKey('alt+backspace');
                            break;
                        default:
                            handled = false;
                    }
                } else {
                    // Regular keys
                    switch (e.key) {
                        case 'Enter':
                            this.sendKey('enter');
                            break;
                        case ' ':
                            this.sendKey('space');
                            break;
                        case 'Backspace':
                            this.sendKey('backspace');
                            break;
                        case 'Delete':
                            this.sendKey('delete');
                            break;
                        case 'Tab':
                            this.sendKey(e.shiftKey ? 'shift+tab' : 'tab');
                            break;
                        case 'Escape':
                            this.sendKey('escape');
                            break;
                        case 'ArrowUp':
                            this.sendKey('up');
                            break;
                        case 'ArrowDown':
                            this.sendKey('down');
                            break;
                        case 'ArrowLeft':
                            this.sendKey('left');
                            break;
                        case 'ArrowRight':
                            this.sendKey('right');
                            break;
                        case 'Home':
                            this.sendKey('home');
                            break;
                        case 'End':
                            this.sendKey('end');
                            break;
                        case 'PageUp':
                            this.sendKey('pageup');
                            break;
                        case 'PageDown':
                            this.sendKey('pagedown');
                            break;
                        case 'Insert':
                            this.sendKey('insert');
                            break;
                        default:
                            // Function keys
                            if (e.key.match(/^F([1-9]|1[0-2])$/)) {
                                this.sendKey(e.key.toLowerCase());
                            }
                            // Regular character input
                            else if (e.key.length === 1) {
                                this.sendKey('runes', [e.key]);
                            } else {
                                handled = false;
                            }
                    }
                }

                if (handled) {
                    e.preventDefault();
                }
            });

            // Paste handling
            this.terminal.addEventListener('paste', (e) => {
                if (!this.connected) return;
                
                e.preventDefault();
                const text = e.clipboardData.getData('text/plain');
                if (text) {
                    // Send paste as individual characters
                    this.sendKey('runes', Array.from(text));
                }
            });

            // Window resize
            let resizeTimeout;
            window.addEventListener('resize', () => {
                clearTimeout(resizeTimeout);
                resizeTimeout = setTimeout(() => {
                    this.calculateAndSendResize();
                }, 300);
            });

            // Visibility change
            document.addEventListener('visibilitychange', () => {
                if (!document.hidden && this.connected) {
                    // Refresh on visibility restore
                    this.sendMessage('refresh', {});
                }
            });
        }

        init() {
            this.setupInputHandlers();
            this.connect();
            
            // Initial focus
            this.terminal.focus();
        }
    }

    // ANSI Parser with full color support
    class ANSIParser {
        constructor() {
            this.colorMap = {
                30: 'black', 31: 'red', 32: 'green', 33: 'yellow',
                34: 'blue', 35: 'magenta', 36: 'cyan', 37: 'white',
                90: 'bright-black', 91: 'bright-red', 92: 'bright-green', 93: 'bright-yellow',
                94: 'bright-blue', 95: 'bright-magenta', 96: 'bright-cyan', 97: 'bright-white'
            };
        }

        parse(text) {
            // Escape HTML first
            text = text
                .replace(/&/g, '&amp;')
                .replace(/</g, '&lt;')
                .replace(/>/g, '&gt;');

            // Parse ANSI sequences
            const regex = /\x1b\[([0-9;]+)m/g;
            let result = '';
            let lastIndex = 0;
            let openSpans = [];

            const getClasses = (codes) => {
                const classes = [];
                const styles = {};

                for (let i = 0; i < codes.length; i++) {
                    const code = parseInt(codes[i]);
                    
                    switch (code) {
                        case 0: // Reset
                            return { reset: true };
                        case 1: classes.push('ansi-bold'); break;
                        case 2: classes.push('ansi-faint'); break;
                        case 3: classes.push('ansi-italic'); break;
                        case 4: classes.push('ansi-underline'); break;
                        case 5: classes.push('ansi-blink'); break;
                        case 7: classes.push('ansi-reverse'); break;
                        case 8: classes.push('ansi-hidden'); break;
                        case 9: classes.push('ansi-strikethrough'); break;
                        case 22: // Normal intensity
                            classes = classes.filter(c => c !== 'ansi-bold' && c !== 'ansi-faint');
                            break;
                        case 23: // Not italic
                            classes = classes.filter(c => c !== 'ansi-italic');
                            break;
                        case 24: // Not underlined
                            classes = classes.filter(c => c !== 'ansi-underline');
                            break;
                        case 38: // 256 color or RGB foreground
                            if (codes[i + 1] === '5' && codes[i + 2]) {
                                // 256 color mode
                                styles.color = this.ansi256ToHex(parseInt(codes[i + 2]));
                                i += 2;
                            } else if (codes[i + 1] === '2' && codes[i + 2] && codes[i + 3] && codes[i + 4]) {
                                // RGB color mode
                                styles.color = `rgb(${codes[i + 2]}, ${codes[i + 3]}, ${codes[i + 4]})`;
                                i += 4;
                            }
                            break;
                        case 48: // 256 color or RGB background
                            if (codes[i + 1] === '5' && codes[i + 2]) {
                                // 256 color mode
                                styles.backgroundColor = this.ansi256ToHex(parseInt(codes[i + 2]));
                                i += 2;
                            } else if (codes[i + 1] === '2' && codes[i + 2] && codes[i + 3] && codes[i + 4]) {
                                // RGB color mode
                                styles.backgroundColor = `rgb(${codes[i + 2]}, ${codes[i + 3]}, ${codes[i + 4]})`;
                                i += 4;
                            }
                            break;
                        default:
                            // Standard colors
                            if (code >= 30 && code <= 37) {
                                classes.push(`ansi-${this.colorMap[code]}`);
                            } else if (code >= 40 && code <= 47) {
                                classes.push(`ansi-bg-${this.colorMap[code - 10]}`);
                            } else if (code >= 90 && code <= 97) {
                                classes.push(`ansi-${this.colorMap[code]}`);
                            } else if (code >= 100 && code <= 107) {
                                classes.push(`ansi-bg-${this.colorMap[code - 10]}`);
                            }
                    }
                }

                return { classes, styles };
            };

            let match;
            while ((match = regex.exec(text)) !== null) {
                // Add text before match
                if (match.index > lastIndex) {
                    result += text.substring(lastIndex, match.index);
                }

                // Parse codes
                const codes = match[1].split(';');
                const { reset, classes, styles } = getClasses(codes);

                if (reset) {
                    // Close all open spans
                    while (openSpans.length > 0) {
                        result += '</span>';
                        openSpans.pop();
                    }
                } else {
                    // Open new span with classes and styles
                    let span = '<span';
                    if (classes.length > 0) {
                        span += ` class="${classes.join(' ')}"`;
                    }
                    if (Object.keys(styles).length > 0) {
                        const styleStr = Object.entries(styles)
                            .map(([k, v]) => `${k}: ${v}`)
                            .join('; ');
                        span += ` style="${styleStr}"`;
                    }
                    span += '>';
                    result += span;
                    openSpans.push(span);
                }

                lastIndex = match.index + match[0].length;
            }

            // Add remaining text
            if (lastIndex < text.length) {
                result += text.substring(lastIndex);
            }

            // Close any remaining spans
            while (openSpans.length > 0) {
                result += '</span>';
                openSpans.pop();
            }

            // Convert newlines to <br>
            result = result.replace(/\n/g, '<br>');

            return result;
        }

        ansi256ToHex(code) {
            // ANSI 256 color palette
            const colors = [
                // Standard colors (0-15)
                '#000000', '#800000', '#008000', '#808000', '#000080', '#800080', '#008080', '#c0c0c0',
                '#808080', '#ff0000', '#00ff00', '#ffff00', '#0000ff', '#ff00ff', '#00ffff', '#ffffff',
                // 216 color cube (16-231)
                ...this.generate216ColorCube(),
                // Grayscale (232-255)
                ...this.generateGrayscale()
            ];
            
            return colors[code] || '#ffffff';
        }

        generate216ColorCube() {
            const colors = [];
            const values = [0, 95, 135, 175, 215, 255];
            
            for (let r = 0; r < 6; r++) {
                for (let g = 0; g < 6; g++) {
                    for (let b = 0; b < 6; b++) {
                        colors.push(`#${values[r].toString(16).padStart(2, '0')}${values[g].toString(16).padStart(2, '0')}${values[b].toString(16).padStart(2, '0')}`);
                    }
                }
            }
            
            return colors;
        }

        generateGrayscale() {
            const colors = [];
            for (let i = 0; i < 24; i++) {
                const value = 8 + i * 10;
                const hex = value.toString(16).padStart(2, '0');
                colors.push(`#${hex}${hex}${hex}`);
            }
            return colors;
        }
    }

    // Initialize client when DOM is ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', () => {
            const client = new TerminusClient();
            client.init();
            window.terminusClient = client; // For debugging
        });
    } else {
        const client = new TerminusClient();
        client.init();
        window.terminusClient = client; // For debugging
    }
})();