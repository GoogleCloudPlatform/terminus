# Widgets Showcase

Demonstrates core widgets: text input, list with filter, table with selection/sort, spinner styles, and an “all widgets” view.

## Run
- `go run examples/widgets/main.go`
- Optional CLI client: start the server, then `./bin/terminus-cli -addr localhost:8890`

## Controls
- `1`-`5`: switch views (TextInput, List, Table, Spinner, All)
- `Tab` (in All view): move focus across sections (TextInput starts focused)
- `l`: toggle spinner
- `n`: cycle spinner style
- `s`: sort table (when in Table view)
- `q` or `Ctrl+C`: quit
- Typing filters the list when its filter box is focused
