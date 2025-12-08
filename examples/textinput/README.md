# Text Input Demo

Single-line text inputs with per-field submit/reset, navigation, and a visible block cursor (web and CLI).

## Run
- `go run examples/textinput/main.go`
- Optional CLI client: start the server, then `./bin/terminus-cli -addr localhost:8890`

## Controls
- `Tab` / `Shift+Tab`: move between fields
- `Enter`: submit the focused field
- `Ctrl+S`: submit all fields
- `Ctrl+R`: reset all fields
- `q` or `Ctrl+C`: quit
