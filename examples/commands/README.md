# Commands Demo

Shows advanced command patterns (sequence, parallel, debounce, throttle, cancellable) plus a debounced search box and spinner updates.

## Run
- `go run examples/commands/main.go`
- Optional CLI client: start the server, then `./bin/terminus-cli -addr localhost:8890`

## Controls
- `q` or `Ctrl+C`: quit
- `1`: run a sequential set of commands
- `2`: run parallel commands
- `3`: HTTP request to api.github.com/zen (requires network)
- `4`: toggle a cancellable 5s timer
- `5`: throttled command (max once/sec)
- `c`: clear the log
- Typing in the search box triggers a 500ms debounced search
