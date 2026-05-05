# Local Dev

Local Dev is a terminal user interface for orchestrating multi-repository development workflows.
It reads a YAML configuration, starts each pane's long-running process, and gives you hotkeys to trigger additional commands from a single window.

## Highlights

- Manage multiple services from one screen with automatic grid layout and mouse support.
- Run `start` commands as soon as the app launches and stream stdout/stderr into dedicated panes.
- Show Git branch names plus ahead/behind counts directly in each pane title.
- Bind custom commands per pane with optional prompts, silent/background execution, and reserved actions like toggling the pane size.

## Requirements

- Git (optional) to display branch names and ahead/behind information in pane headers.

## Installation

- `go install github.com/jiyeol-lee/localdev@latest`
- Or clone the repository and run `go build ./...` to produce a local binary.

## Configuration

### Where configs live

By default `localdev` loads `config.yml` from `$XDG_CONFIG_HOME/localdev/config.yml`.
If `XDG_CONFIG_HOME` is not set, it falls back to `~/Library/Application Support/localdev/config.yml` (MacOS), `~/.config/localdev/config.yml` (Linux).
Use the `--config` flag to select a different file name under the Local Dev config directory (e.g. `localdev --config staging.yml`). The current loader does not treat `--config` as an arbitrary absolute or relative path.

### Minimal example

```yaml
project_settings:
  dir: /Users/johndoe/workspace
  command: git fetch --all
panes:
  - name: api
    dir: api
    start: go run ./cmd/server
    stop: pkill -f cmd/server
  - name: web
    dir: web
    start: npm run dev
    stop: npm run stop
    commands:
      lowerL:
        command: git pull
        description: Pull latest changes
      upperB:
        command: git fetch && git checkout main
        description: Switch to main branch
      lowerT:
        command: <toggle_pane_size>
        description: Toggle pane size
      lowerS:
        command: <start_pane>
        description: Start pane
      lowerX:
        command: <stop_pane>
        description: Stop pane
```

### Project settings options (optional)

- `dir` (optional) – base directory for all panes. Relative paths in pane `dir` options resolve beneath this path.
- `command` (optional) – command executed once when Local Dev launches, before starting any pane `start` commands.

### Pane options

- `name` (required) – label shown in the pane header.
- `dir` (required) – working directory for the commands. Relative paths resolve beneath `project_settings.dir` when it is set.
- `start` (required) – command executed when Local Dev launches; stdout and stderr stream into the pane (`stderr` is tinted brown).
- `stop` (required) – command executed when you exit; Local Dev runs every stop command concurrently and prefixes each line with the pane name.
- `commands` (optional) – map of hotkeys (`lowerA`–`lowerZ`, `upperA`–`upperZ`) to command objects.
  - `command`: (required) command to run when the keybinding is pressed. it can be a shell command or a reserved command.
  - `description`: (optional) description of the command to show in the help menu.
  - `silent`: (optional) if true, the command will be executed without printing the result in the pane. default is false.
  - `autoExecute`: (optional) if true, the command will be executed automatically when the keybinding is pressed. if false, it will display an input prompt to confirm the execution. default is false.

## Running

- `localdev` – loads the default `config.yml` and starts every pane.
- `localdev --config staging.yml` – loads another configuration file from the config directory.
- `localdev --config config.yaml` – loads `config.yaml` from the config directory.

On startup Local Dev runs `project_settings.command` (if defined) before launching each pane's `start` command.
Press `Ctrl+C` or close the terminal to exit; Local Dev prints a stop banner and executes all pane `stop` commands before returning control to your shell.

## Keybindings

- `1`–`9` and `0` focus the corresponding pane (up to ten panes).
- `?` opens the command list modal for the focused pane; it shows descriptions and lets you trigger commands.
- `Esc` closes the command modal or help modal and returns focus to the pane grid.
- Letter keys defined in the pane's `commands` section run or queue the associated command. Mouse clicks can also change focus when no modal is open.

## Reserved commands

- `<toggle_pane_size>` – toggles the focused pane between its normal size and a larger size that occupies most of the terminal window. Pressing the same keybinding again returns to the normal grid view.
- `<start_pane>` – kills any running process in the focused pane and reruns its `start` command. Prior logs are preserved with a separator line.
- `<stop_pane>` – sends `SIGINT` to the focused pane's process group (followed by `SIGKILL` if it doesn't exit within 3 seconds), then runs the pane's config `stop` command. Output is streamed into the pane between separator lines.

## Five-pane Podman test fixture

The `test_files` directory contains a small fixture for checking pane layout, output streaming, and the reserved start/stop/toggle commands:

- `test_files/Dockerfile` – Alpine-based image that prints `$PANE_NAME tick ####` once per second forever. Podman builds this standard Dockerfile directly.
- `test_files/config.yml` – Local Dev config with five panes. Each pane uses Podman to build the same Dockerfile, runs a dedicated container (`localdev-pane-1` through `localdev-pane-5`), sets a distinct `PANE_NAME`, and binds:
  - `s` (`lowerS`) to `<start_pane>`
  - `x` (`lowerX`) to `<stop_pane>`
  - `t` (`lowerT`) to `<toggle_pane_size>`

No compose file is required or included; each pane directly builds and runs the shared Dockerfile with Podman.

With Podman installed, run the fixture from the repository root:

```sh
export XDG_CONFIG_HOME="$PWD/.localdev-test-config"
mkdir -p "$XDG_CONFIG_HOME/localdev"
cp test_files/config.yml "$XDG_CONFIG_HOME/localdev/config.yml"
localdev
```

Setting `XDG_CONFIG_HOME` keeps the test config repo-local and works consistently across platforms. If you do not set it, copy the file to your platform config directory instead: `~/Library/Application Support/localdev/config.yml` on macOS, or `~/.config/localdev/config.yml` on Linux. The pane `dir` values are relative to the shell where you launch `localdev`, so start it from the repository root. Press `?` inside Local Dev to view the configured hotkeys, or press `Ctrl+C` to exit and remove all Podman test containers.
